package handler

import (
	"log/slog"
	"net/url"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles user-facing payment operations.
type PaymentHandler struct {
	orderService  *service.PaymentOrderService
	configService *service.PaymentConfigService
	loadBalancer  *service.PaymentLoadBalancer
	channelRepo   service.PaymentChannelRepository
	planRepo      service.SubscriptionPlanRepository
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(
	orderService *service.PaymentOrderService,
	configService *service.PaymentConfigService,
	loadBalancer *service.PaymentLoadBalancer,
	channelRepo service.PaymentChannelRepository,
	planRepo service.SubscriptionPlanRepository,
) *PaymentHandler {
	return &PaymentHandler{
		orderService:  orderService,
		configService: configService,
		loadBalancer:  loadBalancer,
		channelRepo:   channelRepo,
		planRepo:      planRepo,
	}
}

// CreateOrder handles POST /api/v1/pay/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	var req struct {
		Amount      string `json:"amount" binding:"required"`
		PaymentType string `json:"payment_type" binding:"required"`
		OrderType   string `json:"order_type"`
		PlanID      *int64 `json:"plan_id"`
		SrcHost     string `json:"src_host"`
		SrcURL      string `json:"src_url"`
		ReturnURL   string `json:"return_url"`
		IsMobile    bool   `json:"is_mobile"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "Amount must be a positive number")
		return
	}

	// Validate SrcURL if provided
	if req.SrcURL != "" {
		if len(req.SrcURL) > 2048 {
			response.BadRequest(c, "src_url too long (max 2048)")
			return
		}
		if u, err := url.Parse(req.SrcURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			response.BadRequest(c, "src_url must be a valid HTTP/HTTPS URL")
			return
		}
	}
	// Validate ReturnURL if provided
	if req.ReturnURL != "" {
		if len(req.ReturnURL) > 2048 {
			response.BadRequest(c, "return_url too long (max 2048)")
			return
		}
		if u, err := url.Parse(req.ReturnURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			response.BadRequest(c, "return_url must be a valid HTTP/HTTPS URL")
			return
		}
	}
	// Validate SrcHost if provided
	if len(req.SrcHost) > 253 {
		response.BadRequest(c, "src_host too long (max 253)")
		return
	}

	orderType := req.OrderType
	if orderType == "" {
		orderType = "balance"
	}
	if orderType != "balance" && orderType != "subscription" {
		response.BadRequest(c, "Invalid order type, must be 'balance' or 'subscription'")
		return
	}

	result, err := h.orderService.CreateOrder(c.Request.Context(), service.CreateOrderRequest{
		UserID:      subject.UserID,
		Amount:      amount,
		PaymentType: req.PaymentType,
		OrderType:   orderType,
		PlanID:      req.PlanID,
		ClientIP:    c.ClientIP(),
		SrcHost:     req.SrcHost,
		SrcURL:      req.SrcURL,
		ReturnURL:   req.ReturnURL,
		IsMobile:    req.IsMobile,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	order := result.Order
	feeRate := "0"
	payAmount := order.Amount.String()
	if order.FeeRate != nil {
		feeRate = order.FeeRate.String()
	}
	if order.PayAmount != nil {
		payAmount = order.PayAmount.String()
	}

	// Generate status access token for anonymous order polling
	secret := h.configService.GetStatusAccessSecret(c.Request.Context())
	accessToken := service.CreateOrderStatusAccessToken(order.ID, subject.UserID, secret)

	response.Created(c, dto.CreateOrderResponse{
		OrderID:      order.ID,
		Amount:       order.Amount.String(),
		PayAmount:    payAmount,
		FeeRate:      feeRate,
		Status:       order.Status,
		PaymentType:  order.PaymentType,
		OrderType:    order.OrderType,
		PayURL:       result.PayURL,
		QrCode:       result.QrCode,
		ClientSecret: result.ClientSecret,
		ExpiresAt:    order.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		AccessToken:  accessToken,
	})
}

// ListOrders handles GET /api/v1/pay/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	orders, paginationResult, err := h.orderService.ListUserOrders(c.Request.Context(), subject.UserID, paginationParams(page, pageSize))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UserPaymentOrderDTO, 0, len(orders))
	for i := range orders {
		out = append(out, *dto.UserPaymentOrderFromService(&orders[i]))
	}

	response.Paginated(c, out, paginationResult.Total, page, pageSize)
}

// GetOrder handles GET /api/v1/pay/orders/:id
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if order.UserID != subject.UserID {
		response.NotFound(c, "Order not found")
		return
	}

	response.Success(c, dto.UserPaymentOrderFromService(order))
}

// CancelOrder handles POST /api/v1/pay/orders/:id/cancel
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), orderID, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Order cancelled"})
}

// RequestRefund handles POST /api/v1/pay/orders/:id/refund-request
func (h *PaymentHandler) RequestRefund(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req struct {
		Amount string `json:"amount" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "Amount must be a positive number")
		return
	}

	if err := h.orderService.RequestRefund(c.Request.Context(), orderID, subject.UserID, amount, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Refund requested"})
}

// GetConfig handles GET /api/v1/pay/config
func (h *PaymentHandler) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()

	enabledTypes := h.configService.GetEnabledPaymentTypes(ctx)
	if enabledTypes == nil {
		enabledTypes = []string{}
	}

	var methodLimits []dto.MethodLimitDTO
	limits, err := h.loadBalancer.QueryMethodLimits(ctx, enabledTypes)
	if err != nil {
		slog.Warn("failed to query payment method limits", "error", err)
	} else {
		methodLimits = make([]dto.MethodLimitDTO, 0, len(limits))
		for _, l := range limits {
			remaining := "0"
			if l.Remaining != nil {
				remaining = l.Remaining.String()
			}
			methodLimits = append(methodLimits, dto.MethodLimitDTO{
				PaymentType: l.PaymentType,
				Available:   l.Available,
				DailyLimit:  l.DailyLimit.String(),
				DailyUsed:   l.DailyUsed.String(),
				Remaining:   remaining,
				SingleMin:   l.SingleMin.String(),
				SingleMax:   l.SingleMax.String(),
				FeeRate:     l.FeeRate.String(),
			})
		}
	}

	var pendingCount int
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if ok {
		if cnt, err := h.orderService.CountPendingByUserID(ctx, subject.UserID); err == nil {
			pendingCount = cnt
		}
	}

	response.Success(c, dto.PaymentConfigDTO{
		EnabledPaymentTypes:    enabledTypes,
		MinRechargeAmount:      h.configService.GetMinRechargeAmount(ctx).String(),
		MaxRechargeAmount:      h.configService.GetMaxRechargeAmount(ctx).String(),
		MaxDailyRechargeAmount: h.configService.GetMaxDailyRechargeAmount(ctx).String(),
		BalancePaymentDisabled: h.configService.IsBalancePaymentDisabled(ctx),
		MaxPendingOrders:       h.configService.GetMaxPendingOrders(ctx),
		PendingCount:           pendingCount,
		MethodLimits:           methodLimits,
	})
}

// ListChannels handles GET /api/v1/pay/channels
func (h *PaymentHandler) ListChannels(c *gin.Context) {
	channels, err := h.channelRepo.ListEnabled(c.Request.Context())
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

// ListPlans handles GET /api/v1/pay/subscription-plans
func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.planRepo.ListForSale(c.Request.Context())
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

// paginationParams creates pagination params from page/pageSize.
func paginationParams(page, pageSize int) pagination.PaginationParams {
	return pagination.PaginationParams{Page: page, PageSize: pageSize}
}
