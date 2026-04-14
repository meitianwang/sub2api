package payment

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

const wxPayAPIBase = "https://api.mch.weixin.qq.com"

// WxPayConfig holds the configuration for a WeChat Pay v3 provider.
type WxPayConfig struct {
	AppID       string `json:"appId"`
	MchID       string `json:"mchId"`
	APIKeyV3    string `json:"apiKeyV3"`    // exactly 32 bytes, used as AES-256-GCM key
	PrivateKey  string `json:"privateKey"`  // merchant RSA private key PEM
	SerialNo    string `json:"serialNo"`    // merchant certificate serial number
	PublicKey   string `json:"publicKey"`   // WeChat platform RSA public key PEM
	PublicKeyID string `json:"publicKeyId"` // WeChat platform public key ID (serial)
	NotifyURL   string `json:"notifyUrl"`
}

// WxPayProvider implements the PaymentProvider interface for WeChat Pay v3.
type WxPayProvider struct {
	config     WxPayConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	client     *http.Client
}

// NewWxPayProvider creates a new WxPayProvider from config JSON.
func NewWxPayProvider(configJSON string) (service.PaymentProvider, error) {
	var cfg WxPayConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("parse wxpay config: %w", err)
	}
	if cfg.AppID == "" || cfg.MchID == "" || cfg.APIKeyV3 == "" || cfg.PrivateKey == "" || cfg.SerialNo == "" || cfg.PublicKey == "" || cfg.PublicKeyID == "" || cfg.NotifyURL == "" {
		return nil, fmt.Errorf("wxpay config: appId, mchId, apiKeyV3, privateKey, serialNo, publicKey, publicKeyId, notifyUrl are required")
	}
	if len(cfg.APIKeyV3) != 32 {
		return nil, fmt.Errorf("wxpay config: apiKeyV3 must be exactly 32 bytes, got %d", len(cfg.APIKeyV3))
	}

	privKey, err := parseRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("wxpay config: parse private key: %w", err)
	}

	pubKey, err := parseRSAPublicKey(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("wxpay config: parse public key: %w", err)
	}

	return &WxPayProvider{
		config:     cfg,
		privateKey: privKey,
		publicKey:  pubKey,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *WxPayProvider) Name() string        { return "WeChat Pay" }
func (p *WxPayProvider) ProviderKey() string  { return domain.PaymentProviderWxpay }
func (p *WxPayProvider) SupportedTypes() []string {
	return []string{domain.PaymentTypeWxpayDirect}
}
func (p *WxPayProvider) DefaultLimits() map[string]service.MethodDefaultLimits {
	return map[string]service.MethodDefaultLimits{
		domain.PaymentTypeWxpayDirect: {SingleMax: decimal.NewFromInt(1000), DailyMax: decimal.NewFromInt(10000)},
	}
}

// CreatePayment creates a WeChat Pay order via Native (QR) or H5 API.
func (p *WxPayProvider) CreatePayment(ctx context.Context, req service.CreatePaymentRequest) (*service.CreatePaymentResponse, error) {
	totalFen := amountToFen(req.Amount)

	if req.IsMobile {
		resp, err := p.createH5Payment(ctx, req, totalFen)
		if err != nil {
			// Fallback to Native if H5 returns NO_AUTH
			if isNoAuthError(err) {
				return p.createNativePayment(ctx, req, totalFen)
			}
			return nil, err
		}
		return resp, nil
	}

	return p.createNativePayment(ctx, req, totalFen)
}

func (p *WxPayProvider) createNativePayment(ctx context.Context, req service.CreatePaymentRequest, totalFen int64) (*service.CreatePaymentResponse, error) {
	body := map[string]interface{}{
		"appid":        p.config.AppID,
		"mchid":        p.config.MchID,
		"description":  req.Subject,
		"out_trade_no": req.OrderID,
		"notify_url":   p.config.NotifyURL,
		"amount": map[string]interface{}{
			"total":    totalFen,
			"currency": "CNY",
		},
	}

	var result struct {
		CodeURL string `json:"code_url"`
	}
	if err := p.doRequest(ctx, http.MethodPost, "/v3/pay/transactions/native", body, &result); err != nil {
		return nil, fmt.Errorf("wxpay native payment: %w", err)
	}

	return &service.CreatePaymentResponse{
		TradeNo: req.OrderID,
		QrCode:  result.CodeURL,
	}, nil
}

func (p *WxPayProvider) createH5Payment(ctx context.Context, req service.CreatePaymentRequest, totalFen int64) (*service.CreatePaymentResponse, error) {
	clientIP := req.ClientIP
	if clientIP == "" {
		clientIP = "127.0.0.1"
	}

	body := map[string]interface{}{
		"appid":        p.config.AppID,
		"mchid":        p.config.MchID,
		"description":  req.Subject,
		"out_trade_no": req.OrderID,
		"notify_url":   p.config.NotifyURL,
		"amount": map[string]interface{}{
			"total":    totalFen,
			"currency": "CNY",
		},
		"scene_info": map[string]interface{}{
			"payer_client_ip": clientIP,
			"h5_info": map[string]interface{}{
				"type": "Wap",
			},
		},
	}

	var result struct {
		H5URL string `json:"h5_url"`
	}
	if err := p.doRequest(ctx, http.MethodPost, "/v3/pay/transactions/h5", body, &result); err != nil {
		return nil, err
	}

	return &service.CreatePaymentResponse{
		TradeNo: req.OrderID,
		PayURL:  result.H5URL,
	}, nil
}

// QueryOrder queries a WeChat Pay order by out_trade_no.
func (p *WxPayProvider) QueryOrder(ctx context.Context, tradeNo string) (*service.QueryOrderResponse, error) {
	urlPath := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?mchid=%s", tradeNo, p.config.MchID)

	var result struct {
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		Amount        struct {
			Total int64 `json:"total"`
		} `json:"amount"`
		SuccessTime string `json:"success_time"`
	}
	if err := p.doRequest(ctx, http.MethodGet, urlPath, nil, &result); err != nil {
		return nil, fmt.Errorf("wxpay query order: %w", err)
	}

	status := mapWxTradeState(result.TradeState)
	amount := fenToAmount(result.Amount.Total)

	resp := &service.QueryOrderResponse{
		TradeNo: result.TransactionID,
		Status:  status,
		Amount:  amount,
	}

	if result.SuccessTime != "" {
		if t, err := time.Parse(time.RFC3339, result.SuccessTime); err == nil {
			resp.PaidAt = &t
		}
	}

	return resp, nil
}

// Refund creates a refund for a WeChat Pay order.
func (p *WxPayProvider) Refund(ctx context.Context, req service.RefundRequest) (*service.RefundResponse, error) {
	refundFen := amountToFen(req.Amount)
	totalFen := amountToFen(req.TotalAmount)
	if totalFen <= 0 {
		totalFen = refundFen // fallback if total not provided
	}

	outRefundNo := "refund-" + req.OrderID

	body := map[string]interface{}{
		"out_trade_no":  req.OrderID,
		"out_refund_no": outRefundNo,
		"reason":        req.Reason,
		"amount": map[string]interface{}{
			"refund":   refundFen,
			"total":    totalFen,
			"currency": "CNY",
		},
	}

	var result struct {
		RefundID string `json:"refund_id"`
		Status   string `json:"status"`
	}
	if err := p.doRequest(ctx, http.MethodPost, "/v3/refund/domestic/refunds", body, &result); err != nil {
		return nil, fmt.Errorf("wxpay refund: %w", err)
	}

	refundStatus := "pending"
	if result.Status == "SUCCESS" {
		refundStatus = "success"
	}

	return &service.RefundResponse{
		RefundID: result.RefundID,
		Status:   refundStatus,
	}, nil
}

// VerifyNotification verifies and decrypts a WeChat Pay v3 webhook notification.
func (p *WxPayProvider) VerifyNotification(_ context.Context, rawBody []byte, headers map[string]string) (*service.PaymentNotification, error) {
	// Extract headers (case-insensitive)
	timestamp := getHeaderCaseInsensitive(headers, "Wechatpay-Timestamp")
	nonce := getHeaderCaseInsensitive(headers, "Wechatpay-Nonce")
	signature := getHeaderCaseInsensitive(headers, "Wechatpay-Signature")
	serial := getHeaderCaseInsensitive(headers, "Wechatpay-Serial")

	if timestamp == "" || nonce == "" || signature == "" || serial == "" {
		return nil, fmt.Errorf("wxpay notification: missing required headers")
	}

	// Verify serial matches our configured platform public key ID
	if serial != p.config.PublicKeyID {
		return nil, fmt.Errorf("wxpay notification: serial mismatch: expected %s, got %s", p.config.PublicKeyID, serial)
	}

	// Verify timestamp is within 300 seconds
	ts, err := parseUnixTimestamp(timestamp)
	if err != nil {
		return nil, fmt.Errorf("wxpay notification: invalid timestamp: %w", err)
	}
	if absDuration(time.Since(ts)) > 300*time.Second {
		return nil, fmt.Errorf("wxpay notification: timestamp expired, diff=%v", time.Since(ts))
	}

	// Verify RSA-SHA256 signature
	message := timestamp + "\n" + nonce + "\n" + string(rawBody) + "\n"
	if err := verifyRSASHA256(p.publicKey, []byte(message), signature); err != nil {
		return nil, fmt.Errorf("wxpay notification: signature verification failed: %w", err)
	}

	// Parse notification body
	var notification struct {
		EventType string `json:"event_type"`
		Resource  struct {
			Algorithm      string `json:"algorithm"`
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(rawBody, &notification); err != nil {
		return nil, fmt.Errorf("wxpay notification: parse body: %w", err)
	}

	// Only handle TRANSACTION.SUCCESS and TRANSACTION.* events that indicate final state
	if notification.EventType != "TRANSACTION.SUCCESS" {
		// Non-transaction events (e.g. refund notifications) are irrelevant — acknowledge to stop retries
		if !strings.HasPrefix(notification.EventType, "TRANSACTION.") {
			return nil, nil
		}
	}

	if notification.Resource.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("wxpay notification: unsupported algorithm: %s", notification.Resource.Algorithm)
	}

	// Decrypt AES-256-GCM
	plaintext, err := decryptAESGCM(
		[]byte(p.config.APIKeyV3),
		notification.Resource.Nonce,
		notification.Resource.Ciphertext,
		notification.Resource.AssociatedData,
	)
	if err != nil {
		return nil, fmt.Errorf("wxpay notification: decrypt: %w", err)
	}

	// Parse decrypted transaction data
	var txn struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		Amount        struct {
			Total int64 `json:"total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plaintext, &txn); err != nil {
		return nil, fmt.Errorf("wxpay notification: parse decrypted data: %w", err)
	}

	amount := fenToAmount(txn.Amount.Total)

	notifStatus := "failed"
	if txn.TradeState == "SUCCESS" {
		notifStatus = "success"
	}

	return &service.PaymentNotification{
		TradeNo: txn.TransactionID,
		OrderID: txn.OutTradeNo,
		Amount:  amount,
		Status:  notifStatus,
		RawData: string(rawBody),
	}, nil
}

// ---------- HTTP request helpers ----------

// doRequest sends a signed request to the WeChat Pay v3 API.
func (p *WxPayProvider) doRequest(ctx context.Context, method, urlPath string, reqBody interface{}, result interface{}) error {
	var bodyBytes []byte
	if reqBody != nil {
		var err error
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	fullURL := wxPayAPIBase + urlPath

	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if len(bodyBytes) > 0 {
		httpReq.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		httpReq.ContentLength = int64(len(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	// Sign the request
	authHeader, err := p.buildAuthorizationHeader(method, urlPath, bodyBytes)
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}
	httpReq.Header.Set("Authorization", authHeader)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr wxPayAPIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Code != "" {
			return &apiErr
		}
		return fmt.Errorf("wxpay API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}

	return nil
}

// buildAuthorizationHeader constructs the WECHATPAY2-SHA256-RSA2048 Authorization header.
func (p *WxPayProvider) buildAuthorizationHeader(method, urlPath string, body []byte) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr, err := generateNonce()
	if err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	bodyStr := ""
	if len(body) > 0 {
		bodyStr = string(body)
	}

	// Construct the message to sign
	message := method + "\n" + urlPath + "\n" + timestamp + "\n" + nonceStr + "\n" + bodyStr + "\n"

	// SHA256 hash then RSA PKCS1v15 sign
	hashed := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, p.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("sign message: %w", err)
	}
	sigBase64 := base64.StdEncoding.EncodeToString(sig)

	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		p.config.MchID, nonceStr, timestamp, p.config.SerialNo, sigBase64,
	), nil
}

// ---------- Crypto helpers ----------

// parseRSAPrivateKey parses an RSA private key from PEM (with or without headers).
// Tries PKCS#8 ("PRIVATE KEY") first, then PKCS#1 ("RSA PRIVATE KEY"), matching
// the Alipay provider's approach for maximum compatibility.
func parseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	// Try PKCS#8 PEM header first (more common for modern keys)
	pemStr := formatPEMKey(keyStr, "PRIVATE KEY")
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		// Fall back to PKCS#1 PEM header
		pemStr = formatPEMKey(keyStr, "RSA PRIVATE KEY")
		block, _ = pem.Decode([]byte(pemStr))
	}
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try both DER formats regardless of PEM header
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("PKCS8 key is not RSA")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// parseRSAPublicKey parses an RSA public key from PEM (with or without headers).
func parseRSAPublicKey(keyStr string) (*rsa.PublicKey, error) {
	keyStr = formatPEMKey(keyStr, "PUBLIC KEY")
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKIX first
	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := pub.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("PKIX key is not RSA")
	}

	// Try PKCS1
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// verifyRSASHA256 verifies an RSA-SHA256 signature encoded as base64.
func verifyRSASHA256(pub *rsa.PublicKey, message []byte, signatureBase64 string) error {
	sig, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	hashed := sha256.Sum256(message)
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig)
}

// decryptAESGCM decrypts AES-256-GCM ciphertext from WeChat Pay notification.
// The ciphertext is base64-encoded and contains the GCM tag appended to the actual ciphertext,
// which is exactly the format Go's cipher.AEAD.Open() expects.
func decryptAESGCM(key []byte, nonceStr, ciphertextBase64, associatedData string) ([]byte, error) {
	ciphertextWithTag, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := []byte(nonceStr)
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("nonce size mismatch: expected %d, got %d", gcm.NonceSize(), len(nonce))
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("GCM decrypt: %w", err)
	}

	return plaintext, nil
}

// generateNonce generates a 32-character hex nonce using crypto/rand.
func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ---------- Amount conversion ----------

// amountToFen converts a decimal CNY amount to fen (int64).
// 1 CNY = 100 fen. Uses Round(0) to match the TypeScript Math.round() behavior.
func amountToFen(amount decimal.Decimal) int64 {
	return amount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}

// fenToAmount converts fen (int64) to a decimal CNY amount.
func fenToAmount(fen int64) decimal.Decimal {
	return decimal.NewFromInt(fen).Div(decimal.NewFromInt(100))
}

// ---------- Misc helpers ----------

// wxPayAPIError represents an error response from the WeChat Pay API.
type wxPayAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *wxPayAPIError) Error() string {
	return fmt.Sprintf("wxpay API: code=%s message=%s", e.Code, e.Message)
}

// isNoAuthError checks if an error is a WeChat Pay NO_AUTH error.
func isNoAuthError(err error) bool {
	if apiErr, ok := err.(*wxPayAPIError); ok {
		return apiErr.Code == "NO_AUTH"
	}
	return false
}

// mapWxTradeState maps WeChat Pay trade_state to our canonical status.
func mapWxTradeState(state string) string {
	switch state {
	case "SUCCESS":
		return "paid"
	case "REFUND":
		return "refunded"
	case "CLOSED", "PAYERROR":
		return "failed"
	default:
		return "pending"
	}
}

// getHeaderCaseInsensitive retrieves a header value with case-insensitive key matching.
func getHeaderCaseInsensitive(headers map[string]string, key string) string {
	if v, ok := headers[key]; ok {
		return v
	}
	lowerKey := strings.ToLower(key)
	for k, v := range headers {
		if strings.ToLower(k) == lowerKey {
			return v
		}
	}
	return ""
}

// parseUnixTimestamp parses a Unix timestamp string to time.Time.
func parseUnixTimestamp(s string) (time.Time, error) {
	if len(s) == 0 {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	const maxSafe = int64(9999999999999) // year 2286, well beyond any valid timestamp
	var ts int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return time.Time{}, fmt.Errorf("invalid timestamp character: %c", c)
		}
		ts = ts*10 + int64(c-'0')
		if ts > maxSafe {
			return time.Time{}, fmt.Errorf("timestamp overflow")
		}
	}
	return time.Unix(ts, 0), nil
}

// absDuration returns the absolute value of a time.Duration.
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// CancelPayment closes a WeChat Pay order.
func (p *WxPayProvider) CancelPayment(ctx context.Context, tradeNo string) error {
	urlPath := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s/close", tradeNo)
	body := map[string]interface{}{
		"mchid": p.config.MchID,
	}
	// WeChat close returns 204 No Content on success — doRequest handles this.
	return p.doRequest(ctx, http.MethodPost, urlPath, body, nil)
}

// Compile-time interface compliance checks.
var _ service.PaymentProvider = (*WxPayProvider)(nil)
var _ service.CancelableProvider = (*WxPayProvider)(nil)
var _ service.ProviderWithDefaults = (*WxPayProvider)(nil)
