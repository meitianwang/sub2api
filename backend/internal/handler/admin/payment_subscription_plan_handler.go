package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// PaymentSubscriptionPlanHandler handles admin subscription plan CRUD.
type PaymentSubscriptionPlanHandler struct {
	planRepo     service.SubscriptionPlanRepository
	orderService *service.PaymentOrderService
}

// NewPaymentSubscriptionPlanHandler creates a new PaymentSubscriptionPlanHandler.
func NewPaymentSubscriptionPlanHandler(planRepo service.SubscriptionPlanRepository, orderService *service.PaymentOrderService) *PaymentSubscriptionPlanHandler {
	return &PaymentSubscriptionPlanHandler{planRepo: planRepo, orderService: orderService}
}

// List handles GET /api/v1/admin/pay/subscription-plans
func (h *PaymentSubscriptionPlanHandler) List(c *gin.Context) {
	plans, err := h.planRepo.List(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.SubscriptionPlanDTO, 0, len(plans))
	for i := range plans {
		out = append(out, *dto.SubscriptionPlanFromService(&plans[i]))
	}

	response.Success(c, out)
}

// Create handles POST /api/v1/admin/pay/subscription-plans
func (h *PaymentSubscriptionPlanHandler) Create(c *gin.Context) {
	var req struct {
		GroupID       *int64  `json:"group_id"`
		Name          string  `json:"name" binding:"required"`
		Description   *string `json:"description"`
		Price         string  `json:"price" binding:"required"`
		OriginalPrice *string `json:"original_price"`
		ValidityDays  int     `json:"validity_days"`
		ValidityUnit  string  `json:"validity_unit"`
		Features      *string `json:"features"`
		ProductName   *string `json:"product_name"`
		ForSale       *bool   `json:"for_sale"`
		SortOrder     int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		response.BadRequest(c, "Invalid price")
		return
	}

	var originalPrice *decimal.Decimal
	if req.OriginalPrice != nil {
		d, err := decimal.NewFromString(*req.OriginalPrice)
		if err != nil {
			response.BadRequest(c, "Invalid original_price")
			return
		}
		originalPrice = &d
	}

	if req.ValidityDays <= 0 {
		req.ValidityDays = 30
	}
	if req.ValidityUnit == "" {
		req.ValidityUnit = "day"
	}
	if req.ValidityUnit != "day" && req.ValidityUnit != "week" && req.ValidityUnit != "month" {
		response.BadRequest(c, "Invalid validity_unit, must be 'day', 'week', or 'month'")
		return
	}

	forSale := false
	if req.ForSale != nil {
		forSale = *req.ForSale
	}

	plan := &service.SubscriptionPlan{
		GroupID:       req.GroupID,
		Name:          req.Name,
		Description:   req.Description,
		Price:         price,
		OriginalPrice: originalPrice,
		ValidityDays:  req.ValidityDays,
		ValidityUnit:  req.ValidityUnit,
		Features:      req.Features,
		ProductName:   req.ProductName,
		ForSale:       forSale,
		SortOrder:     req.SortOrder,
	}

	if err := h.planRepo.Create(c.Request.Context(), plan); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, dto.SubscriptionPlanFromService(plan))
}

// Update handles PUT /api/v1/admin/pay/subscription-plans/:id
func (h *PaymentSubscriptionPlanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.planRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var req struct {
		GroupID       *int64  `json:"group_id"`
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		Price         *string `json:"price"`
		OriginalPrice *string `json:"original_price"`
		ValidityDays  *int    `json:"validity_days"`
		ValidityUnit  *string `json:"validity_unit"`
		Features      *string `json:"features"`
		ProductName   *string `json:"product_name"`
		ForSale       *bool   `json:"for_sale"`
		SortOrder     *int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.GroupID != nil {
		plan.GroupID = req.GroupID
	}
	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.Description != nil {
		plan.Description = req.Description
	}
	if req.Price != nil {
		d, err := decimal.NewFromString(*req.Price)
		if err != nil {
			response.BadRequest(c, "Invalid price")
			return
		}
		plan.Price = d
	}
	if req.OriginalPrice != nil {
		d, err := decimal.NewFromString(*req.OriginalPrice)
		if err != nil {
			response.BadRequest(c, "Invalid original_price")
			return
		}
		plan.OriginalPrice = &d
	}
	if req.ValidityDays != nil {
		if *req.ValidityDays <= 0 {
			response.BadRequest(c, "validity_days must be positive")
			return
		}
		plan.ValidityDays = *req.ValidityDays
	}
	if req.ValidityUnit != nil {
		if *req.ValidityUnit != "day" && *req.ValidityUnit != "week" && *req.ValidityUnit != "month" {
			response.BadRequest(c, "Invalid validity_unit, must be 'day', 'week', or 'month'")
			return
		}
		plan.ValidityUnit = *req.ValidityUnit
	}
	if req.Features != nil {
		plan.Features = req.Features
	}
	if req.ProductName != nil {
		plan.ProductName = req.ProductName
	}
	if req.ForSale != nil {
		plan.ForSale = *req.ForSale
	}
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}

	if err := h.planRepo.Update(c.Request.Context(), plan); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.SubscriptionPlanFromService(plan))
}

// Delete handles DELETE /api/v1/admin/pay/subscription-plans/:id
func (h *PaymentSubscriptionPlanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	hasActive, err := h.orderService.HasActiveOrdersForPlan(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to check active orders")
		return
	}
	if hasActive {
		response.BadRequest(c, "Cannot delete plan while there are active orders (pending/paid/recharging) referencing it")
		return
	}

	if err := h.planRepo.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Plan deleted"})
}
