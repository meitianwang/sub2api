package admin

import (
	"regexp"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// validFeeRateKeySuffix validates the suffix after "pay_fee_rate_" or "pay_fee_rate_provider_".
var validFeeRateKeySuffix = regexp.MustCompile(`^[a-z0-9_]+$`)

// PaymentConfigHandler handles admin payment configuration.
type PaymentConfigHandler struct {
	configService *service.PaymentConfigService
}

// NewPaymentConfigHandler creates a new PaymentConfigHandler.
func NewPaymentConfigHandler(configService *service.PaymentConfigService) *PaymentConfigHandler {
	return &PaymentConfigHandler{configService: configService}
}

// Get handles GET /api/v1/admin/pay/config
func (h *PaymentConfigHandler) Get(c *gin.Context) {
	all := h.configService.GetAllSettings(c.Request.Context())

	// Mask sensitive values
	masked := make(map[string]string, len(all))
	for k, v := range all {
		if isSensitiveKey(k) {
			if len(v) > 4 {
				masked[k] = "****" + v[len(v)-4:]
			} else {
				masked[k] = "****"
			}
		} else {
			masked[k] = v
		}
	}

	response.Success(c, gin.H{"configs": masked})
}

// allowedPaymentConfigKeys is the explicit allowlist of payment config keys
// that can be modified via the admin API.
var allowedPaymentConfigKeys = map[string]bool{
	"pay_enabled_payment_types":        true,
	"pay_min_recharge_amount":          true,
	"pay_max_recharge_amount":          true,
	"pay_max_daily_recharge_amount":    true,
	"pay_order_timeout_minutes":        true,
	"pay_product_name":                 true,
	"pay_product_name_prefix":          true,
	"pay_product_name_suffix":          true,
	"pay_balance_payment_disabled":     true,
	"pay_cancel_rate_limit_enabled":    true,
	"pay_cancel_rate_limit_window":     true,
	"pay_cancel_rate_limit_unit":       true,
	"pay_cancel_rate_limit_max":        true,
	"pay_cancel_rate_limit_window_mode": true,
	"pay_max_pending_orders":           true,
	"pay_load_balance_strategy":        true,
	"pay_providers":                    true,
	"pay_help_image_url":               true,
	"pay_help_text":                    true,
	"pay_max_daily_amount_alipay":      true,
	"pay_max_daily_amount_wxpay":       true,
	"pay_max_daily_amount_stripe":      true,
	"pay_auto_refund_enabled":          true,
	"pay_grace_period_minutes":         true,
}

// Update handles PUT /api/v1/admin/pay/config
func (h *PaymentConfigHandler) Update(c *gin.Context) {
	var req struct {
		Configs map[string]string `json:"configs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Filter: skip masked values and reject unknown keys
	filtered := make(map[string]string, len(req.Configs))
	for k, v := range req.Configs {
		if strings.Contains(v, "****") && isSensitiveKey(k) {
			continue
		}
		if !allowedPaymentConfigKeys[k] {
			// Allow dynamic fee rate keys with strict suffix validation
			if strings.HasPrefix(k, "pay_fee_rate_") {
				suffix := strings.TrimPrefix(k, "pay_fee_rate_")
				suffix = strings.TrimPrefix(suffix, "provider_")
				if !validFeeRateKeySuffix.MatchString(suffix) || suffix == "" {
					continue
				}
			} else {
				continue
			}
		}
		filtered[k] = v
	}

	if err := h.configService.UpdateSettings(c.Request.Context(), filtered); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Config updated", "updated": len(filtered)})
}

func isSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)
	return strings.Contains(upper, "KEY") ||
		strings.Contains(upper, "SECRET") ||
		strings.Contains(upper, "PASSWORD") ||
		strings.Contains(upper, "PRIVATE")
}
