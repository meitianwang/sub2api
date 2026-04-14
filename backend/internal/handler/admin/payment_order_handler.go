package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentOrderHandler handles admin payment order management.
type PaymentOrderHandler struct {
	orderService *service.PaymentOrderService
}

// NewPaymentOrderHandler creates a new PaymentOrderHandler.
func NewPaymentOrderHandler(orderService *service.PaymentOrderService) *PaymentOrderHandler {
	return &PaymentOrderHandler{orderService: orderService}
}

// List handles GET /api/v1/admin/pay/orders
func (h *PaymentOrderHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	var filter service.PaymentOrderListFilter
	if v := c.Query("user_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid user_id: must be an integer")
			return
		}
		filter.UserID = &id
	}
	if v := c.Query("status"); v != "" {
		if !isValidOrderStatus(v) {
			response.BadRequest(c, "Invalid status filter")
			return
		}
		filter.Status = &v
	}
	if v := c.Query("order_type"); v != "" {
		if v != "balance" && v != "subscription" {
			response.BadRequest(c, "Invalid order_type filter, must be 'balance' or 'subscription'")
			return
		}
		filter.OrderType = &v
	}
	if v := c.Query("payment_type"); v != "" {
		filter.PaymentType = &v
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err != nil {
			response.BadRequest(c, "Invalid date_from: must be RFC3339 format")
			return
		} else {
			filter.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err != nil {
			response.BadRequest(c, "Invalid date_to: must be RFC3339 format")
			return
		} else {
			filter.DateTo = &t
		}
	}

	orders, paginationResult, err := h.orderService.ListOrders(c.Request.Context(), filter, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentOrderDTO, 0, len(orders))
	for i := range orders {
		out = append(out, *dto.PaymentOrderFromService(&orders[i]))
	}

	response.Paginated(c, out, paginationResult.Total, page, pageSize)
}

// GetByID handles GET /api/v1/admin/pay/orders/:id
func (h *PaymentOrderHandler) GetByID(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, logs, err := h.orderService.GetOrderWithAuditLogs(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	auditLogs := make([]dto.PaymentAuditLogDTO, 0, len(logs))
	for i := range logs {
		auditLogs = append(auditLogs, *dto.PaymentAuditLogFromService(&logs[i]))
	}

	response.Success(c, dto.AdminOrderDetailDTO{
		Order:     *dto.PaymentOrderFromService(order),
		AuditLogs: auditLogs,
	})
}

// Cancel handles POST /api/v1/admin/pay/orders/:id/cancel
func (h *PaymentOrderHandler) Cancel(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.AdminCancelOrder(c.Request.Context(), orderID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Order cancelled"})
}

// RetryRecharge handles POST /api/v1/admin/pay/orders/:id/retry
func (h *PaymentOrderHandler) RetryRecharge(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.RetryRecharge(c.Request.Context(), orderID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Recharge retry initiated"})
}

var validOrderStatuses = map[string]bool{
	domain.PaymentOrderStatusPending:           true,
	domain.PaymentOrderStatusPaid:              true,
	domain.PaymentOrderStatusRecharging:        true,
	domain.PaymentOrderStatusCompleted:         true,
	domain.PaymentOrderStatusExpired:           true,
	domain.PaymentOrderStatusCancelled:         true,
	domain.PaymentOrderStatusFailed:            true,
	domain.PaymentOrderStatusRefundRequested:   true,
	domain.PaymentOrderStatusRefunding:         true,
	domain.PaymentOrderStatusPartiallyRefunded: true,
	domain.PaymentOrderStatusRefunded:          true,
	domain.PaymentOrderStatusRefundFailed:      true,
}

func isValidOrderStatus(s string) bool {
	return validOrderStatuses[s]
}
