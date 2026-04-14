package service

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// PaymentProvider defines the interface all payment providers must implement.
type PaymentProvider interface {
	Name() string
	ProviderKey() string
	SupportedTypes() []string
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error)
	QueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResponse, error)
	VerifyNotification(ctx context.Context, rawBody []byte, headers map[string]string) (*PaymentNotification, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error)
}

// CancelableProvider is an optional extension for providers that support order cancellation.
type CancelableProvider interface {
	PaymentProvider
	CancelPayment(ctx context.Context, tradeNo string) error
}

// CreatePaymentRequest is the provider-agnostic payment creation request.
type CreatePaymentRequest struct {
	OrderID     string
	Amount      decimal.Decimal // in CNY
	PaymentType string
	Subject     string
	NotifyURL   string
	ReturnURL   string
	ClientIP    string
	IsMobile    bool
}

// CreatePaymentResponse is the provider-agnostic payment creation response.
type CreatePaymentResponse struct {
	TradeNo      string
	PayURL       string
	QrCode       string
	ClientSecret string // Stripe Payment Element client secret
}

// QueryOrderResponse is the provider-agnostic order query response.
type QueryOrderResponse struct {
	TradeNo string
	Status  string // "pending", "paid", "failed", "refunded"
	Amount  decimal.Decimal
	PaidAt  *time.Time
}

// PaymentNotification is the parsed webhook/notification data.
type PaymentNotification struct {
	TradeNo string
	OrderID string
	Amount  decimal.Decimal
	Status  string // "success", "failed"
	RawData string
}

// RefundRequest is the provider-agnostic refund request.
type RefundRequest struct {
	TradeNo     string
	OrderID     string
	Amount      decimal.Decimal
	TotalAmount decimal.Decimal // Original order total (required by WeChat Pay)
	Reason      string
}

// RefundResponse is the provider-agnostic refund response.
type RefundResponse struct {
	RefundID string
	Status   string // "success", "pending", "failed"
}

// MethodDefaultLimits defines fallback limits for a payment type when no explicit config exists.
type MethodDefaultLimits struct {
	SingleMax decimal.Decimal // 0 = unlimited
	DailyMax  decimal.Decimal // 0 = unlimited
}

// ProviderWithDefaults is an optional extension for providers that declare default limits.
type ProviderWithDefaults interface {
	PaymentProvider
	DefaultLimits() map[string]MethodDefaultLimits
}
