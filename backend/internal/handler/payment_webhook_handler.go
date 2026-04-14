package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// maxWebhookBodySize limits webhook request body to 1MB to prevent OOM.
const maxWebhookBodySize = 1 << 20

// PaymentWebhookHandler handles payment provider webhook callbacks.
type PaymentWebhookHandler struct {
	orderService *service.PaymentOrderService
}

// NewPaymentWebhookHandler creates a new PaymentWebhookHandler.
func NewPaymentWebhookHandler(orderService *service.PaymentOrderService) *PaymentWebhookHandler {
	return &PaymentWebhookHandler{orderService: orderService}
}

// NotifyEasyPay handles GET /api/v1/pay/notify/easypay
// EasyPay sends notification data as query parameters.
func (h *PaymentWebhookHandler) NotifyEasyPay(c *gin.Context) {
	rawQuery := c.Request.URL.RawQuery
	headers := extractHeaders(c.Request)

	err := h.orderService.HandleNotification(c.Request.Context(), "easypay", []byte(rawQuery), headers)
	if err != nil {
		slog.Warn("easypay notification failed", "error", err)
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

// NotifyAlipay handles POST /api/v1/pay/notify/alipay
func (h *PaymentWebhookHandler) NotifyAlipay(c *gin.Context) {
	rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodySize))
	if err != nil {
		slog.Warn("alipay notification: failed to read body", "error", err)
		c.String(http.StatusOK, "fail")
		return
	}

	headers := extractHeaders(c.Request)
	err = h.orderService.HandleNotification(c.Request.Context(), "alipay", rawBody, headers)
	if err != nil {
		slog.Warn("alipay notification failed", "error", err)
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

// NotifyWxpay handles POST /api/v1/pay/notify/wxpay
func (h *PaymentWebhookHandler) NotifyWxpay(c *gin.Context) {
	rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodySize))
	if err != nil {
		slog.Warn("wxpay notification: failed to read body", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "failed to read body"})
		return
	}

	headers := map[string]string{
		"Wechatpay-Timestamp": c.GetHeader("Wechatpay-Timestamp"),
		"Wechatpay-Nonce":     c.GetHeader("Wechatpay-Nonce"),
		"Wechatpay-Signature": c.GetHeader("Wechatpay-Signature"),
		"Wechatpay-Serial":    c.GetHeader("Wechatpay-Serial"),
	}

	err = h.orderService.HandleNotification(c.Request.Context(), "wxpay", rawBody, headers)
	if err != nil {
		slog.Warn("wxpay notification failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "processing failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "success"})
}

// NotifyStripe handles POST /api/v1/pay/notify/stripe
func (h *PaymentWebhookHandler) NotifyStripe(c *gin.Context) {
	rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodySize))
	if err != nil {
		slog.Warn("stripe webhook: failed to read body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	headers := map[string]string{
		"Stripe-Signature": c.GetHeader("Stripe-Signature"),
	}

	err = h.orderService.HandleNotification(c.Request.Context(), "stripe", rawBody, headers)
	if err != nil {
		slog.Warn("stripe webhook failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed, will retry"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}

// extractHeaders extracts all HTTP headers into a simple map.
func extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string, len(r.Header))
	for k := range r.Header {
		headers[k] = r.Header.Get(k)
	}
	return headers
}
