package admin

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentProviderInstanceHandler handles admin payment provider instance CRUD.
type PaymentProviderInstanceHandler struct {
	instanceRepo service.PaymentProviderInstanceRepository
	encryptor    service.SecretEncryptor
	orderService *service.PaymentOrderService
}

// NewPaymentProviderInstanceHandler creates a new PaymentProviderInstanceHandler.
func NewPaymentProviderInstanceHandler(
	instanceRepo service.PaymentProviderInstanceRepository,
	encryptor service.SecretEncryptor,
	orderService *service.PaymentOrderService,
) *PaymentProviderInstanceHandler {
	return &PaymentProviderInstanceHandler{
		instanceRepo: instanceRepo,
		encryptor:    encryptor,
		orderService: orderService,
	}
}

// List handles GET /api/v1/admin/pay/provider-instances
func (h *PaymentProviderInstanceHandler) List(c *gin.Context) {
	instances, err := h.instanceRepo.List(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	providerKey := c.Query("provider_key")
	out := make([]gin.H, 0, len(instances))
	for i := range instances {
		inst := &instances[i]
		if providerKey != "" && inst.ProviderKey != providerKey {
			continue
		}
		out = append(out, h.instanceToResponse(inst))
	}

	response.Success(c, out)
}

// Create handles POST /api/v1/admin/pay/provider-instances
func (h *PaymentProviderInstanceHandler) Create(c *gin.Context) {
	var req struct {
		ProviderKey    string            `json:"provider_key" binding:"required"`
		Name           string            `json:"name" binding:"required"`
		Config         map[string]string `json:"config" binding:"required"`
		SupportedTypes string            `json:"supported_types"`
		Enabled        *bool             `json:"enabled"`
		SortOrder      int               `json:"sort_order"`
		Limits         *string           `json:"limits"`
		RefundEnabled  bool              `json:"refund_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if !isValidProviderKey(req.ProviderKey) {
		response.BadRequest(c, "Invalid provider_key, must be one of: easypay, alipay, wxpay, stripe")
		return
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		response.InternalError(c, "Failed to serialize config")
		return
	}
	encrypted, err := h.encryptor.Encrypt(string(configJSON))
	if err != nil {
		response.InternalError(c, "Failed to encrypt config")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	instance := &service.PaymentProviderInstance{
		ProviderKey:    req.ProviderKey,
		Name:           req.Name,
		Config:         encrypted,
		SupportedTypes: req.SupportedTypes,
		Enabled:        enabled,
		SortOrder:      req.SortOrder,
		Limits:         req.Limits,
		RefundEnabled:  req.RefundEnabled,
	}

	if err := h.instanceRepo.Create(c.Request.Context(), instance); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, h.instanceToResponse(instance))
}

// GetByID handles GET /api/v1/admin/pay/provider-instances/:id
func (h *PaymentProviderInstanceHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	instance, err := h.instanceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, h.instanceToResponse(instance))
}

// Update handles PUT /api/v1/admin/pay/provider-instances/:id
func (h *PaymentProviderInstanceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	instance, err := h.instanceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var req struct {
		Name           *string            `json:"name"`
		Config         map[string]string  `json:"config"`
		SupportedTypes *string            `json:"supported_types"`
		Enabled        *bool              `json:"enabled"`
		SortOrder      *int               `json:"sort_order"`
		Limits         *string            `json:"limits"`
		RefundEnabled  *bool              `json:"refund_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// If changing credentials, check for active orders
	if req.Config != nil {
		hasActive, err := h.orderService.HasActiveOrdersForProviderInstance(c.Request.Context(), id)
		if err != nil {
			response.InternalError(c, "Failed to check active orders")
			return
		}
		if hasActive {
			response.BadRequest(c, "Cannot modify credentials while there are active orders (pending/paid/recharging)")
			return
		}
	}

	if req.Name != nil {
		instance.Name = *req.Name
	}
	if req.Config != nil {
		// Merge masked values with existing config
		existingConfig, err := h.decryptConfig(instance.Config)
		if err != nil {
			response.InternalError(c, "Failed to decrypt existing config")
			return
		}
		for k, v := range req.Config {
			if strings.Contains(v, "****") {
				continue // skip masked values
			}
			existingConfig[k] = v
		}
		configJSON, err := json.Marshal(existingConfig)
		if err != nil {
			response.InternalError(c, "Failed to serialize config")
			return
		}
		encrypted, err := h.encryptor.Encrypt(string(configJSON))
		if err != nil {
			response.InternalError(c, "Failed to encrypt config")
			return
		}
		instance.Config = encrypted
	}
	if req.SupportedTypes != nil {
		instance.SupportedTypes = *req.SupportedTypes
	}
	if req.Enabled != nil {
		instance.Enabled = *req.Enabled
	}
	if req.SortOrder != nil {
		instance.SortOrder = *req.SortOrder
	}
	if req.Limits != nil {
		instance.Limits = req.Limits
	}
	if req.RefundEnabled != nil {
		instance.RefundEnabled = *req.RefundEnabled
	}

	if err := h.instanceRepo.Update(c.Request.Context(), instance); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, h.instanceToResponse(instance))
}

// Delete handles DELETE /api/v1/admin/pay/provider-instances/:id
func (h *PaymentProviderInstanceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	hasActive, err := h.orderService.HasActiveOrdersForProviderInstance(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to check active orders")
		return
	}
	if hasActive {
		response.BadRequest(c, "Cannot delete instance while there are active orders (pending/paid/recharging)")
		return
	}

	if err := h.instanceRepo.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Instance deleted"})
}

func (h *PaymentProviderInstanceHandler) decryptConfig(encrypted string) (map[string]string, error) {
	decrypted, err := h.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}
	var config map[string]string
	if err := json.Unmarshal([]byte(decrypted), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func (h *PaymentProviderInstanceHandler) instanceToResponse(inst *service.PaymentProviderInstance) gin.H {
	config, err := h.decryptConfig(inst.Config)
	if err != nil {
		slog.Error("failed to decrypt provider instance config", "instance_id", inst.ID, "error", err)
		config = map[string]string{}
	}
	if config == nil {
		config = map[string]string{}
	}
	// Mask sensitive fields — safe keys are shown in full, everything else is masked
	safeKeys := map[string]bool{
		"pid": true, "app_id": true, "appid": true, "mch_id": true,
		"api_base": true, "notify_url": true, "return_url": true,
		"currency": true, "cid": true, "cid_alipay": true, "cid_wxpay": true,
		"cert_serial": true, "public_key_id": true,
	}
	masked := make(map[string]string, len(config))
	for k, v := range config {
		if safeKeys[strings.ToLower(k)] {
			masked[k] = v
		} else if len(v) > 4 {
			masked[k] = "****" + v[len(v)-4:]
		} else {
			masked[k] = "****"
		}
	}

	return gin.H{
		"id":              inst.ID,
		"provider_key":    inst.ProviderKey,
		"name":            inst.Name,
		"config":          masked,
		"supported_types": inst.SupportedTypes,
		"enabled":         inst.Enabled,
		"sort_order":      inst.SortOrder,
		"limits":          inst.Limits,
		"refund_enabled":  inst.RefundEnabled,
		"created_at":      inst.CreatedAt,
		"updated_at":      inst.UpdatedAt,
	}
}

var validProviderKeys = map[string]bool{
	"easypay": true,
	"alipay":  true,
	"wxpay":   true,
	"stripe":  true,
}

func isValidProviderKey(key string) bool {
	return validProviderKeys[key]
}
