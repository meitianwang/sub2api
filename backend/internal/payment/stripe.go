package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v82"
	piClient "github.com/stripe/stripe-go/v82/paymentintent"
	refundClient "github.com/stripe/stripe-go/v82/refund"
	"github.com/stripe/stripe-go/v82/webhook"
)

// StripeConfig holds the configuration for a Stripe provider instance.
type StripeConfig struct {
	SecretKey     string `json:"secretKey"`
	WebhookSecret string `json:"webhookSecret"`
	Currency      string `json:"currency"`
}

// StripeProvider implements the PaymentProvider interface for Stripe.
type StripeProvider struct {
	config   StripeConfig
	piClient *piClient.Client
	rfClient *refundClient.Client
}

// NewStripeProvider creates a new StripeProvider from config JSON.
func NewStripeProvider(configJSON string) (service.PaymentProvider, error) {
	var cfg StripeConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("parse stripe config: %w", err)
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("stripe config: secretKey is required")
	}
	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("stripe config: webhookSecret is required")
	}
	if cfg.Currency == "" {
		cfg.Currency = "cny"
	}

	// Use per-instance clients to avoid global state races with multiple Stripe accounts
	backend := stripe.GetBackend(stripe.APIBackend)

	return &StripeProvider{
		config:   cfg,
		piClient: &piClient.Client{B: backend, Key: cfg.SecretKey},
		rfClient: &refundClient.Client{B: backend, Key: cfg.SecretKey},
	}, nil
}

func (p *StripeProvider) Name() string        { return "Stripe" }
func (p *StripeProvider) ProviderKey() string  { return domain.PaymentProviderStripe }
func (p *StripeProvider) SupportedTypes() []string {
	return []string{domain.PaymentTypeStripe}
}
func (p *StripeProvider) DefaultLimits() map[string]service.MethodDefaultLimits {
	return map[string]service.MethodDefaultLimits{
		domain.PaymentTypeStripe: {}, // unlimited by default
	}
}

// zeroDecimalCurrencies lists currencies that use the major unit directly (no cents).
// See: https://docs.stripe.com/currencies#zero-decimal
var zeroDecimalCurrencies = map[string]bool{
	"bif": true, "clp": true, "djf": true, "gnf": true, "jpy": true,
	"kmf": true, "krw": true, "mga": true, "pyg": true, "rwf": true,
	"ugx": true, "vnd": true, "vuv": true, "xaf": true, "xof": true, "xpf": true,
}

// toSmallestUnit converts a decimal amount to the smallest currency unit for Stripe.
func toSmallestUnit(amount decimal.Decimal, currency string) int64 {
	if zeroDecimalCurrencies[strings.ToLower(currency)] {
		return amount.Round(0).IntPart()
	}
	return amount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}

// fromSmallestUnit converts a Stripe smallest-unit int64 back to a decimal amount.
func fromSmallestUnit(units int64, currency string) decimal.Decimal {
	if zeroDecimalCurrencies[strings.ToLower(currency)] {
		return decimal.NewFromInt(units)
	}
	return decimal.NewFromInt(units).Div(decimal.NewFromInt(100))
}

func (p *StripeProvider) CreatePayment(_ context.Context, req service.CreatePaymentRequest) (*service.CreatePaymentResponse, error) {
	amountInCents := toSmallestUnit(req.Amount, p.config.Currency)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInCents),
		Currency: stripe.String(p.config.Currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Description: stripe.String(req.Subject),
	}
	params.SetIdempotencyKey("pi-" + req.OrderID)
	params.AddMetadata("orderId", req.OrderID)

	pi, err := p.piClient.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create payment intent: %w", err)
	}

	return &service.CreatePaymentResponse{
		TradeNo:      pi.ID,
		ClientSecret: pi.ClientSecret,
	}, nil
}

func (p *StripeProvider) QueryOrder(_ context.Context, tradeNo string) (*service.QueryOrderResponse, error) {
	pi, err := p.piClient.Get(tradeNo, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe get payment intent: %w", err)
	}

	status := mapPaymentIntentStatus(pi.Status)

	resp := &service.QueryOrderResponse{
		TradeNo: pi.ID,
		Status:  status,
		Amount:  fromSmallestUnit(pi.Amount, p.config.Currency),
	}

	return resp, nil
}

func (p *StripeProvider) VerifyNotification(_ context.Context, rawBody []byte, headers map[string]string) (*service.PaymentNotification, error) {
	sig := getHeaderCaseInsensitive(headers, "Stripe-Signature")
	if sig == "" {
		return nil, fmt.Errorf("stripe: missing stripe-signature header")
	}

	event, err := webhook.ConstructEvent(rawBody, sig, p.config.WebhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe webhook signature verification failed: %w", err)
	}

	var status string
	switch event.Type {
	case "payment_intent.succeeded":
		status = "success"
	case "payment_intent.payment_failed":
		status = "failed"
	default:
		// Unhandled event type; return nil to indicate no action needed.
		return nil, nil
	}

	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return nil, fmt.Errorf("stripe: unmarshal payment intent from event: %w", err)
	}

	orderID := pi.Metadata["orderId"]
	if orderID == "" {
		return nil, fmt.Errorf("stripe: missing orderId in payment intent metadata")
	}

	return &service.PaymentNotification{
		TradeNo: pi.ID,
		OrderID: orderID,
		Amount:  fromSmallestUnit(pi.Amount, p.config.Currency),
		Status:  status,
		RawData: string(rawBody),
	}, nil
}

func (p *StripeProvider) Refund(_ context.Context, req service.RefundRequest) (*service.RefundResponse, error) {
	amountInCents := toSmallestUnit(req.Amount, p.config.Currency)

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.TradeNo),
		Amount:        stripe.Int64(amountInCents),
		Reason:        stripe.String(string(stripe.RefundReasonRequestedByCustomer)),
	}

	r, err := p.rfClient.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe refund: %w", err)
	}

	return &service.RefundResponse{
		RefundID: r.ID,
		Status:   mapRefundStatus(r.Status),
	}, nil
}

// CancelPayment cancels a Stripe PaymentIntent.
func (p *StripeProvider) CancelPayment(_ context.Context, tradeNo string) error {
	params := &stripe.PaymentIntentCancelParams{}
	_, err := p.piClient.Cancel(tradeNo, params)
	if err != nil {
		return fmt.Errorf("stripe cancel payment intent: %w", err)
	}
	return nil
}

// mapPaymentIntentStatus maps a Stripe PaymentIntentStatus to the internal status string.
func mapPaymentIntentStatus(s stripe.PaymentIntentStatus) string {
	switch s {
	case stripe.PaymentIntentStatusSucceeded:
		return "paid"
	case stripe.PaymentIntentStatusCanceled:
		return "failed"
	default:
		return "pending"
	}
}

// mapRefundStatus maps a Stripe RefundStatus to the internal status string.
func mapRefundStatus(s stripe.RefundStatus) string {
	switch s {
	case stripe.RefundStatusSucceeded:
		return "success"
	case stripe.RefundStatusFailed:
		return "failed"
	case stripe.RefundStatusPending:
		return "pending"
	default:
		return "pending"
	}
}

// Compile-time interface compliance checks.
var _ service.PaymentProvider = (*StripeProvider)(nil)
var _ service.CancelableProvider = (*StripeProvider)(nil)
var _ service.ProviderWithDefaults = (*StripeProvider)(nil)
