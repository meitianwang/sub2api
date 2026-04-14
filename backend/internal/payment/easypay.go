package payment

import (
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

// EasyPayConfig holds the configuration for an EasyPay instance.
type EasyPayConfig struct {
	PID       string `json:"pid"`
	PKey      string `json:"pkey"`
	APIBase   string `json:"apiBase"`
	NotifyURL string `json:"notifyUrl"`
	ReturnURL string `json:"returnUrl"`
	CID       string `json:"cid,omitempty"`
	CIDAlipay string `json:"cidAlipay,omitempty"`
	CIDWxpay  string `json:"cidWxpay,omitempty"`
}

// EasyPayProvider implements the PaymentProvider interface for EasyPay.
type EasyPayProvider struct {
	config EasyPayConfig
	client *http.Client
}

// NewEasyPayProvider creates a new EasyPayProvider from config JSON.
func NewEasyPayProvider(configJSON string) (service.PaymentProvider, error) {
	var cfg EasyPayConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("parse easypay config: %w", err)
	}
	if cfg.PID == "" || cfg.PKey == "" || cfg.APIBase == "" {
		return nil, fmt.Errorf("easypay config: pid, pkey, apiBase are required")
	}
	return &EasyPayProvider{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *EasyPayProvider) Name() string                { return "EasyPay" }
func (p *EasyPayProvider) ProviderKey() string         { return domain.PaymentProviderEasyPay }
func (p *EasyPayProvider) SupportedTypes() []string {
	return []string{domain.PaymentTypeAlipay, domain.PaymentTypeWxpay}
}
func (p *EasyPayProvider) DefaultLimits() map[string]service.MethodDefaultLimits {
	return map[string]service.MethodDefaultLimits{
		domain.PaymentTypeAlipay: {SingleMax: decimal.NewFromInt(1000), DailyMax: decimal.NewFromInt(10000)},
		domain.PaymentTypeWxpay:  {SingleMax: decimal.NewFromInt(1000), DailyMax: decimal.NewFromInt(10000)},
	}
}

func (p *EasyPayProvider) CreatePayment(ctx context.Context, req service.CreatePaymentRequest) (*service.CreatePaymentResponse, error) {
	// Map payment type to EasyPay type
	epType := req.PaymentType // "alipay" or "wxpay"

	clientIP := req.ClientIP
	if clientIP == "" {
		clientIP = "127.0.0.1"
	}

	params := map[string]string{
		"pid":          p.config.PID,
		"type":         epType,
		"out_trade_no": req.OrderID,
		"notify_url":   p.config.NotifyURL,
		"return_url":   p.config.ReturnURL,
		"name":         req.Subject,
		"money":        req.Amount.StringFixed(2),
		"clientip":     clientIP,
	}

	if req.ReturnURL != "" {
		params["return_url"] = req.ReturnURL
	}

	// Resolve channel ID
	cid := p.resolveCID(epType)
	if cid != "" {
		params["cid"] = cid
	}

	if req.IsMobile {
		params["device"] = "mobile"
	}

	// Sign and add signature
	params["sign"] = easyPaySign(params, p.config.PKey)
	params["sign_type"] = "MD5"

	// POST to create order
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.APIBase+"/mapi.php", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("easypay create payment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		TradeNo string `json:"trade_no"`
		PayURL  string `json:"payurl"`
		PayURL2 string `json:"payurl2"` // mobile-specific redirect URL
		QrCode  string `json:"qrcode"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if result.Code != 1 {
		return nil, fmt.Errorf("easypay error: %s", result.Msg)
	}

	payURL := result.PayURL
	if req.IsMobile && result.PayURL2 != "" {
		payURL = result.PayURL2
	}

	return &service.CreatePaymentResponse{
		TradeNo: result.TradeNo,
		PayURL:  payURL,
		QrCode:  result.QrCode,
	}, nil
}

func (p *EasyPayProvider) QueryOrder(ctx context.Context, tradeNo string) (*service.QueryOrderResponse, error) {
	formData := url.Values{
		"act":          {"order"},
		"pid":          {p.config.PID},
		"key":          {p.config.PKey},
		"out_trade_no": {tradeNo},
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.APIBase+"/api.php", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("easypay query order: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		TradeNo string `json:"trade_no"`
		Status  int    `json:"status"`
		Money   string `json:"money"`
		Msg     string `json:"msg"`
		EndTime string `json:"endtime"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Match TypeScript behavior: don't error on code != 1 (e.g. order not found).
	// Instead, return "pending" status so callers can distinguish "query failed" from "not paid".
	status := "pending"
	if result.Code == 1 && result.Status == 1 {
		status = "paid"
	}

	amount, err := decimal.NewFromString(result.Money)
	if err != nil {
		return nil, fmt.Errorf("easypay query: invalid money %q: %w", result.Money, err)
	}

	qResp := &service.QueryOrderResponse{
		TradeNo: result.TradeNo,
		Status:  status,
		Amount:  amount,
	}

	if result.EndTime != "" {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		if loc == nil {
			loc = time.FixedZone("CST", 8*3600)
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", result.EndTime, loc); err == nil {
			qResp.PaidAt = &t
		}
	}

	return qResp, nil
}

// VerifyNotification verifies an EasyPay payment callback.
// NOTE: EasyPay sends notifications via GET request. The HTTP handler must pass
// the URL query string (req.URL.RawQuery) as rawBody, not the request body.
func (p *EasyPayProvider) VerifyNotification(_ context.Context, rawBody []byte, _ map[string]string) (*service.PaymentNotification, error) {
	// Parse form-urlencoded params (from query string for GET callbacks)
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return nil, fmt.Errorf("parse notification body: %w", err)
	}

	params := make(map[string]string)
	for k := range values {
		params[k] = values.Get(k)
	}

	// Verify PID matches
	if params["pid"] != p.config.PID {
		return nil, fmt.Errorf("pid mismatch: expected %s, got %s", p.config.PID, params["pid"])
	}

	// Verify signature
	sign := params["sign"]
	if !easyPayVerifySign(params, p.config.PKey, sign) {
		return nil, fmt.Errorf("signature verification failed")
	}

	// Check trade status
	tradeStatus := params["trade_status"]
	notifStatus := "failed"
	if tradeStatus == "TRADE_SUCCESS" {
		notifStatus = "success"
	}

	// Validate amount
	amount, err := decimal.NewFromString(params["money"])
	if err != nil || !amount.IsPositive() {
		return nil, fmt.Errorf("easypay notification: invalid amount %q", params["money"])
	}

	return &service.PaymentNotification{
		TradeNo: params["trade_no"],
		OrderID: params["out_trade_no"],
		Amount:  amount,
		Status:  notifStatus,
		RawData: string(rawBody),
	}, nil
}

func (p *EasyPayProvider) Refund(ctx context.Context, req service.RefundRequest) (*service.RefundResponse, error) {
	formData := url.Values{
		"pid":          {p.config.PID},
		"key":          {p.config.PKey},
		"trade_no":     {req.TradeNo},
		"out_trade_no": {req.OrderID},
		"money":        {req.Amount.StringFixed(2)},
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.APIBase+"/api.php?act=refund", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("easypay refund: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if result.Code != 1 {
		return nil, fmt.Errorf("easypay refund error: %s", result.Msg)
	}

	return &service.RefundResponse{
		RefundID: req.TradeNo + "-refund",
		Status:   "success",
	}, nil
}

func (p *EasyPayProvider) resolveCID(paymentType string) string {
	if paymentType == "alipay" {
		if cid := normalizeCIDList(p.config.CIDAlipay); cid != "" {
			return cid
		}
		return normalizeCIDList(p.config.CID)
	}
	if cid := normalizeCIDList(p.config.CIDWxpay); cid != "" {
		return cid
	}
	return normalizeCIDList(p.config.CID)
}

// normalizeCIDList trims whitespace from each comma-separated CID entry and filters empty values.
func normalizeCIDList(cid string) string {
	if cid == "" {
		return ""
	}
	parts := strings.Split(cid, ",")
	var clean []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}
	return strings.Join(clean, ",")
}

// Compile-time interface compliance checks.
var _ service.PaymentProvider = (*EasyPayProvider)(nil)
var _ service.ProviderWithDefaults = (*EasyPayProvider)(nil)

// easyPaySign generates an MD5 signature for EasyPay.
func easyPaySign(params map[string]string, pkey string) string {
	// Filter and sort
	type kv struct{ k, v string }
	var pairs []kv
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })

	var sb strings.Builder
	for i, p := range pairs {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(p.k)
		sb.WriteByte('=')
		sb.WriteString(p.v)
	}
	sb.WriteString(pkey)

	h := md5.Sum([]byte(sb.String()))
	return hex.EncodeToString(h[:])
}

// easyPayVerifySign verifies an EasyPay MD5 signature using constant-time comparison.
func easyPayVerifySign(params map[string]string, pkey, expected string) bool {
	computed := easyPaySign(params, pkey)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(expected)) == 1
}
