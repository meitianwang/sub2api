package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// PaymentChannelHandler handles admin payment channel CRUD.
type PaymentChannelHandler struct {
	channelRepo service.PaymentChannelRepository
}

// NewPaymentChannelHandler creates a new PaymentChannelHandler.
func NewPaymentChannelHandler(channelRepo service.PaymentChannelRepository) *PaymentChannelHandler {
	return &PaymentChannelHandler{channelRepo: channelRepo}
}

// List handles GET /api/v1/admin/pay/channels
func (h *PaymentChannelHandler) List(c *gin.Context) {
	channels, err := h.channelRepo.List(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentChannelDTO, 0, len(channels))
	for i := range channels {
		out = append(out, *dto.PaymentChannelFromService(&channels[i]))
	}

	response.Success(c, out)
}

// Create handles POST /api/v1/admin/pay/channels
func (h *PaymentChannelHandler) Create(c *gin.Context) {
	var req struct {
		GroupID        *int64  `json:"group_id"`
		Name           string  `json:"name" binding:"required"`
		Platform       string  `json:"platform" binding:"required"`
		RateMultiplier string  `json:"rate_multiplier"`
		Description    *string `json:"description"`
		Models         *string `json:"models"`
		Features       *string `json:"features"`
		SortOrder      int     `json:"sort_order"`
		Enabled        *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	rateMul := decimal.NewFromInt(1)
	if req.RateMultiplier != "" {
		d, err := decimal.NewFromString(req.RateMultiplier)
		if err != nil {
			response.BadRequest(c, "Invalid rate_multiplier: must be a valid decimal number")
			return
		}
		if d.LessThanOrEqual(decimal.Zero) {
			response.BadRequest(c, "Invalid rate_multiplier: must be positive")
			return
		}
		rateMul = d
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Validate group_id uniqueness if provided
	if req.GroupID != nil {
		existing, err := h.channelRepo.GetByGroupID(c.Request.Context(), *req.GroupID)
		if err == nil && existing != nil {
			response.BadRequest(c, "A channel with this group_id already exists")
			return
		}
	}

	channel := &service.PaymentChannel{
		GroupID:        req.GroupID,
		Name:           req.Name,
		Platform:       req.Platform,
		RateMultiplier: rateMul,
		Description:    req.Description,
		Models:         req.Models,
		Features:       req.Features,
		SortOrder:      req.SortOrder,
		Enabled:        enabled,
	}

	if err := h.channelRepo.Create(c.Request.Context(), channel); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, dto.PaymentChannelFromService(channel))
}

// Update handles PUT /api/v1/admin/pay/channels/:id
func (h *PaymentChannelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	channel, err := h.channelRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var req struct {
		GroupID        *int64  `json:"group_id"`
		Name           *string `json:"name"`
		Platform       *string `json:"platform"`
		RateMultiplier *string `json:"rate_multiplier"`
		Description    *string `json:"description"`
		Models         *string `json:"models"`
		Features       *string `json:"features"`
		SortOrder      *int    `json:"sort_order"`
		Enabled        *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.GroupID != nil {
		// Validate group_id uniqueness if changing
		if channel.GroupID == nil || *req.GroupID != *channel.GroupID {
			existing, err := h.channelRepo.GetByGroupID(c.Request.Context(), *req.GroupID)
			if err == nil && existing != nil && existing.ID != channel.ID {
				response.BadRequest(c, "A channel with this group_id already exists")
				return
			}
		}
		channel.GroupID = req.GroupID
	}
	if req.Name != nil {
		channel.Name = *req.Name
	}
	if req.Platform != nil {
		channel.Platform = *req.Platform
	}
	if req.RateMultiplier != nil {
		d, err := decimal.NewFromString(*req.RateMultiplier)
		if err != nil {
			response.BadRequest(c, "Invalid rate_multiplier: must be a valid decimal number")
			return
		}
		if d.LessThanOrEqual(decimal.Zero) {
			response.BadRequest(c, "Invalid rate_multiplier: must be positive")
			return
		}
		channel.RateMultiplier = d
	}
	if req.Description != nil {
		channel.Description = req.Description
	}
	if req.Models != nil {
		channel.Models = req.Models
	}
	if req.Features != nil {
		channel.Features = req.Features
	}
	if req.SortOrder != nil {
		channel.SortOrder = *req.SortOrder
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
	}

	if err := h.channelRepo.Update(c.Request.Context(), channel); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentChannelFromService(channel))
}

// Delete handles DELETE /api/v1/admin/pay/channels/:id
func (h *PaymentChannelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	if err := h.channelRepo.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Channel deleted"})
}
