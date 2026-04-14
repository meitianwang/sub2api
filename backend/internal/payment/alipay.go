package payment

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const alipayGatewayURL = "https://openapi.alipay.com/gateway.do"

// AlipayConfig holds the configuration for an Alipay direct payment instance.
type AlipayConfig struct {
	AppID          string `json:"appId"`
	PrivateKey     string `json:"privateKey"`
	PublicKey      string `json:"publicKey"` // Alipay's public key for signature verification
	NotifyURL      string `json:"notifyUrl"`
	ReturnURL      string `json:"returnUrl"`
	PayPageBaseURL string `json:"payPageBaseUrl"` // Frontend base URL for PC QR code entry page (e.g. "https://example.com/pay")
}

// AlipayProvider implements the PaymentProvider interface for Alipay direct integration.
type AlipayProvider struct {
	config     AlipayConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	client     *http.Client
}

// NewAlipayProvider creates a new AlipayProvider from config JSON.
func NewAlipayProvider(configJSON string) (service.PaymentProvider, error) {
	var cfg AlipayConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("parse alipay config: %w", err)
	}
	if cfg.AppID == "" || cfg.PrivateKey == "" || cfg.PublicKey == "" {
		return nil, fmt.Errorf("alipay config: appId, privateKey, publicKey are required")
	}

	// Try PKCS#8 ("PRIVATE KEY") first — matches the TypeScript original which uses
	// -----BEGIN PRIVATE KEY-----. Fall back to PKCS#1 ("RSA PRIVATE KEY") if PEM decode fails.
	privPEM := formatPEMKey(cfg.PrivateKey, "PRIVATE KEY")
	privBlock, _ := pem.Decode([]byte(privPEM))
	if privBlock == nil {
		privPEM = formatPEMKey(cfg.PrivateKey, "RSA PRIVATE KEY")
		privBlock, _ = pem.Decode([]byte(privPEM))
	}
	if privBlock == nil {
		return nil, fmt.Errorf("alipay config: failed to decode private key PEM")
	}

	privKey, err := parsePrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("alipay config: parse private key: %w", err)
	}

	pubPEM := formatPEMKey(cfg.PublicKey, "PUBLIC KEY")
	pubBlock, _ := pem.Decode([]byte(pubPEM))
	if pubBlock == nil {
		return nil, fmt.Errorf("alipay config: failed to decode public key PEM")
	}

	pubIface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("alipay config: parse public key: %w", err)
	}
	pubKey, ok := pubIface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("alipay config: public key is not RSA")
	}

	return &AlipayProvider{
		config:     cfg,
		privateKey: privKey,
		publicKey:  pubKey,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *AlipayProvider) Name() string        { return "Alipay" }
func (p *AlipayProvider) ProviderKey() string  { return domain.PaymentProviderAlipay }
func (p *AlipayProvider) SupportedTypes() []string {
	return []string{domain.PaymentTypeAlipayDirect}
}
func (p *AlipayProvider) DefaultLimits() map[string]service.MethodDefaultLimits {
	return map[string]service.MethodDefaultLimits{
		domain.PaymentTypeAlipayDirect: {SingleMax: decimal.NewFromInt(1000), DailyMax: decimal.NewFromInt(10000)},
	}
}

func (p *AlipayProvider) CreatePayment(_ context.Context, req service.CreatePaymentRequest) (*service.CreatePaymentResponse, error) {
	method := "alipay.trade.page.pay"
	productCode := "FAST_INSTANT_TRADE_PAY"
	if req.IsMobile {
		method = "alipay.trade.wap.pay"
		productCode = "QUICK_WAP_WAY"
	}

	bizContent, err := json.Marshal(map[string]string{
		"out_trade_no": req.OrderID,
		"total_amount": req.Amount.StringFixed(2),
		"subject":      req.Subject,
		"product_code": productCode,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal biz_content: %w", err)
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	timestamp := time.Now().In(loc).Format("2006-01-02 15:04:05")

	notifyURL := p.config.NotifyURL
	if req.NotifyURL != "" {
		notifyURL = req.NotifyURL
	}
	returnURL := p.config.ReturnURL
	if req.ReturnURL != "" {
		returnURL = req.ReturnURL
	}

	params := map[string]string{
		"app_id":      p.config.AppID,
		"method":      method,
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   timestamp,
		"version":     "1.0",
		"biz_content": string(bizContent),
		"notify_url":  notifyURL,
		"return_url":  returnURL,
	}

	sign, err := p.signParams(params)
	if err != nil {
		return nil, fmt.Errorf("sign params: %w", err)
	}
	params["sign"] = sign

	// Build the full redirect URL
	vals := url.Values{}
	for k, v := range params {
		vals.Set(k, v)
	}

	payURL := alipayGatewayURL + "?" + vals.Encode()

	resp := &service.CreatePaymentResponse{
		TradeNo: req.OrderID,
		PayURL:  payURL,
	}

	// For PC (non-mobile), the QR code should point to a frontend entry page (not the raw
	// Alipay gateway URL). This matches the TypeScript original where PC users see a QR code
	// pointing to /pay/{orderId}, and the entry page handles the Alipay redirect.
	if !req.IsMobile && p.config.PayPageBaseURL != "" {
		entryURL := strings.TrimRight(p.config.PayPageBaseURL, "/") + "/" + req.OrderID
		resp.PayURL = entryURL
		resp.QrCode = entryURL
	} else if !req.IsMobile {
		resp.QrCode = payURL
	}

	return resp, nil
}

func (p *AlipayProvider) QueryOrder(ctx context.Context, tradeNo string) (*service.QueryOrderResponse, error) {
	bizContent, _ := json.Marshal(map[string]string{
		"out_trade_no": tradeNo,
	})

	respBody, err := p.serverExecute(ctx, "alipay.trade.query", string(bizContent))
	if err != nil {
		return nil, fmt.Errorf("alipay query order: %w", err)
	}

	responseKey := "alipay_trade_query_response"
	respData, err := p.extractAndVerifyResponse(respBody, responseKey)
	if err != nil {
		return nil, fmt.Errorf("alipay query order: %w", err)
	}

	var result struct {
		Code        string `json:"code"`
		Msg         string `json:"msg"`
		SubCode     string `json:"sub_code"`
		TradeNo     string `json:"trade_no"`
		TradeStatus string `json:"trade_status"`
		TotalAmount string `json:"total_amount"`
		SendPayDate string `json:"send_pay_date"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}
	// ACQ.TRADE_NOT_EXIST: order hasn't been created at Alipay yet — return pending
	if result.SubCode == "ACQ.TRADE_NOT_EXIST" {
		return &service.QueryOrderResponse{
			TradeNo: tradeNo,
			Status:  "pending",
			Amount:  decimal.Zero,
		}, nil
	}
	if result.Code != "10000" {
		return nil, alipayAPIError(result.Code, result.Msg, result.SubCode, "")
	}

	status := mapAlipayTradeStatus(result.TradeStatus)
	amount, err := decimal.NewFromString(result.TotalAmount)
	if err != nil || !amount.IsPositive() {
		return nil, fmt.Errorf("alipay query: invalid total_amount %q", result.TotalAmount)
	}

	resp := &service.QueryOrderResponse{
		TradeNo: result.TradeNo,
		Status:  status,
		Amount:  amount,
	}

	if result.SendPayDate != "" {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", result.SendPayDate, loc); err == nil {
			resp.PaidAt = &t
		}
	}

	return resp, nil
}

func (p *AlipayProvider) Refund(ctx context.Context, req service.RefundRequest) (*service.RefundResponse, error) {
	bizContent, _ := json.Marshal(map[string]string{
		"out_trade_no":   req.OrderID,
		"refund_amount":  req.Amount.StringFixed(2),
		"refund_reason":  req.Reason,
		"out_request_no": req.OrderID + "-refund",
	})

	respBody, err := p.serverExecute(ctx, "alipay.trade.refund", string(bizContent))
	if err != nil {
		return nil, fmt.Errorf("alipay refund: %w", err)
	}

	responseKey := "alipay_trade_refund_response"
	respData, err := p.extractAndVerifyResponse(respBody, responseKey)
	if err != nil {
		return nil, fmt.Errorf("alipay refund: %w", err)
	}

	var result struct {
		Code       string `json:"code"`
		Msg        string `json:"msg"`
		SubCode    string `json:"sub_code"`
		SubMsg     string `json:"sub_msg"`
		TradeNo    string `json:"trade_no"`
		FundChange string `json:"fund_change"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parse refund response: %w", err)
	}
	if result.Code != "10000" {
		return nil, alipayAPIError(result.Code, result.Msg, result.SubCode, result.SubMsg)
	}

	status := "success"
	if result.FundChange != "Y" {
		status = "pending"
	}

	refundID := result.TradeNo
	if refundID == "" {
		refundID = req.OrderID + "-refund"
	}

	return &service.RefundResponse{
		RefundID: refundID,
		Status:   status,
	}, nil
}

func (p *AlipayProvider) VerifyNotification(_ context.Context, rawBody []byte, headers map[string]string) (*service.PaymentNotification, error) {
	// Decode body — try charset from Content-Type header, fall back to UTF-8.
	bodyStr := decodeAlipayBody(rawBody, headers)

	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		return nil, fmt.Errorf("parse notification body: %w", err)
	}

	params := make(map[string]string, len(values))
	for k := range values {
		params[k] = values.Get(k)
	}

	// Validate sign_type is RSA2
	if signType := params["sign_type"]; signType != "" && strings.ToUpper(signType) != "RSA2" {
		return nil, fmt.Errorf("unsupported sign_type %q, only RSA2 is accepted", signType)
	}

	sign := params["sign"]
	if sign == "" {
		return nil, fmt.Errorf("missing sign in notification")
	}
	// url.ParseQuery decodes '+' as space, but the sign is base64 which uses '+'.
	// Normalize spaces back to '+' for correct base64 decoding.
	sign = strings.ReplaceAll(sign, " ", "+")

	// Build verification string from params excluding sign and sign_type
	verifyStr := buildAlipayNotifySignString(params)

	if err := p.verifySignature(verifyStr, sign); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Validate required notification fields
	for _, requiredKey := range []string{"trade_no", "out_trade_no", "trade_status", "app_id", "total_amount"} {
		if params[requiredKey] == "" {
			return nil, fmt.Errorf("alipay notification: missing required field %q", requiredKey)
		}
	}

	// Validate app_id matches
	if params["app_id"] != p.config.AppID {
		return nil, fmt.Errorf("app_id mismatch: expected %s, got %s", p.config.AppID, params["app_id"])
	}

	// Check trade status
	tradeStatus := params["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		// Return a failed notification so the order service can log a warning,
		// matching the TypeScript original's behavior for debugging visibility.
		return &service.PaymentNotification{
			TradeNo: params["trade_no"],
			OrderID: params["out_trade_no"],
			Amount:  decimal.Zero,
			Status:  "failed",
			RawData: string(rawBody),
		}, nil
	}

	// Validate amount
	amount, err := decimal.NewFromString(params["total_amount"])
	if err != nil || !amount.IsPositive() {
		return nil, fmt.Errorf("alipay notification: invalid total_amount %q", params["total_amount"])
	}

	return &service.PaymentNotification{
		TradeNo: params["trade_no"],
		OrderID: params["out_trade_no"],
		Amount:  amount,
		Status:  "success",
		RawData: string(rawBody),
	}, nil
}

// serverExecute sends a server-side API request to Alipay and returns the raw response body.
func (p *AlipayProvider) serverExecute(ctx context.Context, method, bizContent string) ([]byte, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	timestamp := time.Now().In(loc).Format("2006-01-02 15:04:05")

	params := map[string]string{
		"app_id":      p.config.AppID,
		"method":      method,
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   timestamp,
		"version":     "1.0",
		"biz_content": bizContent,
	}

	sign, err := p.signParams(params)
	if err != nil {
		return nil, fmt.Errorf("sign params: %w", err)
	}
	params["sign"] = sign

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, alipayGatewayURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return body, nil
}

// extractAndVerifyResponse extracts the response JSON substring for signature verification
// and verifies it against the "sign" field in the raw body.
// The responseKey JSON must be extracted via byte-level search (not re-marshaled) to preserve
// the exact bytes Alipay signed.
func (p *AlipayProvider) extractAndVerifyResponse(rawBody []byte, responseKey string) ([]byte, error) {
	// Parse top-level to get the sign
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return nil, fmt.Errorf("parse response envelope: %w", err)
	}

	var sign string
	if signRaw, ok := envelope["sign"]; ok {
		if err := json.Unmarshal(signRaw, &sign); err != nil {
			return nil, fmt.Errorf("parse sign field: %w", err)
		}
	}

	// Extract the exact response substring from raw bytes for signature verification.
	// We must NOT re-marshal the JSON because Alipay signs the exact byte sequence.
	needle := []byte(`"` + responseKey + `":`)
	idx := bytes.Index(rawBody, needle)
	if idx < 0 {
		return nil, fmt.Errorf("response key %q not found in body", responseKey)
	}

	// Find the start of the response object (the '{' after the key)
	start := idx + len(needle)
	for start < len(rawBody) && rawBody[start] != '{' {
		start++
	}
	if start >= len(rawBody) {
		return nil, fmt.Errorf("response object not found after key %q", responseKey)
	}

	// Find the matching closing brace, accounting for nested braces
	depth := 0
	end := start
	for end < len(rawBody) {
		switch rawBody[end] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end++
				goto found
			}
		case '"':
			// Skip string content to avoid counting braces inside strings
			end++
			for end < len(rawBody) {
				if rawBody[end] == '\\' {
					end += 2 // skip backslash and the escaped character
					continue
				}
				if rawBody[end] == '"' {
					break
				}
				end++
			}
		}
		end++
	}
	return nil, fmt.Errorf("unmatched braces in response for key %q", responseKey)

found:
	responseData := rawBody[start:end]

	// Verify signature if present
	if sign != "" {
		if err := p.verifySignature(string(responseData), sign); err != nil {
			return nil, fmt.Errorf("response signature verification failed: %w", err)
		}
	}

	return responseData, nil
}

// signParams creates an RSA-SHA256 signature for the given parameters.
func (p *AlipayProvider) signParams(params map[string]string) (string, error) {
	content := buildAlipayRequestSignString(params)

	h := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, p.privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", fmt.Errorf("rsa sign: %w", err)
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

// verifySignature verifies an RSA-SHA256 signature against Alipay's public key.
func (p *AlipayProvider) verifySignature(content, signBase64 string) error {
	sigBytes, err := base64.StdEncoding.DecodeString(signBase64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	h := sha256.Sum256([]byte(content))
	if err := rsa.VerifyPKCS1v15(p.publicKey, crypto.SHA256, h[:], sigBytes); err != nil {
		// Debug logging for signature verification failures — critical for diagnosing
		// production issues where notification signatures don't match.
		contentPreview := content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		signPreview := signBase64
		if len(signPreview) > 40 {
			signPreview = signPreview[:40] + "..."
		}
		slog.Debug("alipay signature verification failed",
			"contentPreview", contentPreview,
			"signPreview", signPreview,
			"error", err)
		return err
	}
	return nil
}

// buildAlipayRequestSignString constructs the string to sign for outgoing requests:
// filter out "sign" and empty values (keep "sign_type"), sort remaining keys alphabetically.
func buildAlipayRequestSignString(params map[string]string) string {
	return buildAlipaySignStringInternal(params, false)
}

// buildAlipayNotifySignString constructs the string to verify for incoming notifications:
// filter out "sign", "sign_type", and empty values, sort remaining keys alphabetically.
func buildAlipayNotifySignString(params map[string]string) string {
	return buildAlipaySignStringInternal(params, true)
}

func buildAlipaySignStringInternal(params map[string]string, excludeSignType bool) string {
	type kv struct{ k, v string }
	var pairs []kv
	for k, v := range params {
		if k == "sign" || v == "" {
			continue
		}
		if excludeSignType && k == "sign_type" {
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
	return sb.String()
}

// formatPEMKey converts a bare base64 key string into proper PEM format.
// If the input already has PEM headers, it is returned after normalizing line endings.
// keyType should be "RSA PRIVATE KEY" or "PUBLIC KEY".
func formatPEMKey(raw, keyType string) string {
	// Handle literal "\n" escape sequences (e.g. from env vars or copy-paste)
	raw = strings.ReplaceAll(raw, `\n`, "\n")
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.TrimSpace(raw)

	header := "-----BEGIN " + keyType + "-----"
	if strings.HasPrefix(raw, header) {
		return raw
	}

	// Strip any whitespace or newlines from the bare base64
	bare := strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(raw)

	// Insert newline every 64 characters
	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteByte('\n')
	for i := 0; i < len(bare); i += 64 {
		end := i + 64
		if end > len(bare) {
			end = len(bare)
		}
		sb.WriteString(bare[i:end])
		sb.WriteByte('\n')
	}
	sb.WriteString("-----END " + keyType + "-----")

	return sb.String()
}

// CancelPayment closes the trade at Alipay via alipay.trade.close.
func (p *AlipayProvider) CancelPayment(ctx context.Context, tradeNo string) error {
	bizContent, _ := json.Marshal(map[string]string{
		"out_trade_no": tradeNo,
	})

	respBody, err := p.serverExecute(ctx, "alipay.trade.close", string(bizContent))
	if err != nil {
		return fmt.Errorf("alipay close order: %w", err)
	}

	responseKey := "alipay_trade_close_response"
	respData, err := p.extractAndVerifyResponse(respBody, responseKey)
	if err != nil {
		return fmt.Errorf("alipay close order: %w", err)
	}

	var result struct {
		Code    string `json:"code"`
		Msg     string `json:"msg"`
		SubCode string `json:"sub_code"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return fmt.Errorf("parse close response: %w", err)
	}
	// 10000 = success; ACQ.TRADE_NOT_EXIST = already gone (treat as success)
	if result.Code != "10000" && result.SubCode != "ACQ.TRADE_NOT_EXIST" {
		return alipayAPIError(result.Code, result.Msg, result.SubCode, "")
	}

	return nil
}

// alipayAPIError formats an Alipay API error, preferring sub_code/sub_msg when available
// for more specific error details (e.g. "ACQ.TRADE_NOT_EXIST" instead of generic "40004").
func alipayAPIError(code, msg, subCode, subMsg string) error {
	errCode := code
	errMsg := msg
	if subCode != "" {
		errCode = subCode
	}
	if subMsg != "" {
		errMsg = subMsg
	}
	return fmt.Errorf("alipay API error: [%s] %s", errCode, errMsg)
}

// Compile-time interface compliance checks.
var _ service.PaymentProvider = (*AlipayProvider)(nil)
var _ service.CancelableProvider = (*AlipayProvider)(nil)
var _ service.ProviderWithDefaults = (*AlipayProvider)(nil)

// decodeAlipayBody decodes the notification body to a string, detecting charset from
// Content-Type header or body parameter. Falls back through UTF-8 → GBK → GB18030.
func decodeAlipayBody(rawBody []byte, headers map[string]string) string {
	primary := detectAlipayCharset(rawBody, headers)
	candidates := uniqueStrings([]string{primary, "utf-8", "gbk", "gb18030"})

	for _, cs := range candidates {
		decoded, err := decodeBytes(rawBody, cs)
		if err == nil && !strings.Contains(decoded, "\uFFFD") {
			return decoded
		}
	}
	// Last resort: treat as UTF-8 (best effort)
	return string(rawBody)
}

// detectAlipayCharset extracts the charset from Content-Type header or a &charset= body parameter.
func detectAlipayCharset(rawBody []byte, headers map[string]string) string {
	// Try Content-Type header first
	ct := headers["content-type"]
	if ct == "" {
		ct = headers["Content-Type"]
	}
	if cs := extractCharsetFromContentType(ct); cs != "" {
		return cs
	}
	// Try &charset= in body (latin1-safe scan)
	if cs := extractCharsetFromBody(rawBody); cs != "" {
		return cs
	}
	return "utf-8"
}

var charsetRe = regexp.MustCompile(`(?i)charset=([^\s;]+)`)
var bodyCharsetRe = regexp.MustCompile(`(?i)(?:^|&)charset=([^&]+)`)

func extractCharsetFromContentType(ct string) string {
	m := charsetRe.FindStringSubmatch(ct)
	if m == nil {
		return ""
	}
	return normalizeCharset(m[1])
}

func extractCharsetFromBody(raw []byte) string {
	m := bodyCharsetRe.FindSubmatch(raw)
	if m == nil {
		return ""
	}
	return normalizeCharset(string(m[1]))
}

func normalizeCharset(cs string) string {
	cs = strings.ToLower(strings.TrimSpace(cs))
	cs = strings.Trim(cs, `"'`)
	switch cs {
	case "utf8":
		return "utf-8"
	case "gb2312", "gb_2312-80":
		return "gbk"
	default:
		return cs
	}
}

func decodeBytes(raw []byte, charset string) (string, error) {
	switch charset {
	case "utf-8", "":
		if !utf8.Valid(raw) {
			return "", fmt.Errorf("invalid utf-8")
		}
		return string(raw), nil
	case "gbk":
		return decodeWithEncoding(raw, simplifiedchinese.GBK)
	case "gb18030":
		return decodeWithEncoding(raw, simplifiedchinese.GB18030)
	default:
		return string(raw), nil
	}
}

func decodeWithEncoding(raw []byte, enc encoding.Encoding) (string, error) {
	reader := transform.NewReader(bytes.NewReader(raw), enc.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

// mapAlipayTradeStatus maps Alipay trade_status to our internal status.
func mapAlipayTradeStatus(status string) string {
	switch status {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return "paid"
	case "TRADE_CLOSED":
		return "failed"
	default:
		return "pending"
	}
}

// parsePrivateKey attempts to parse an RSA private key from DER bytes,
// trying PKCS#1 first, then PKCS#8.
func parsePrivateKey(der []byte) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	keyIface, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse as PKCS#1 or PKCS#8: %w", err)
	}
	rsaKey, ok := keyIface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("PKCS#8 key is not RSA")
	}
	return rsaKey, nil
}
