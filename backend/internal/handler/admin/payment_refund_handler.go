package admin

import (
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// PaymentRefundHandler handles admin refund operations.
type PaymentRefundHandler struct {
	orderService *service.PaymentOrderService
}

// NewPaymentRefundHandler creates a new PaymentRefundHandler.
func NewPaymentRefundHandler(orderService *service.PaymentOrderService) *PaymentRefundHandler {
	return &PaymentRefundHandler{orderService: orderService}
}

// ProcessRefund handles POST /api/v1/admin/pay/refund
func (h *PaymentRefundHandler) ProcessRefund(c *gin.Context) {
	var req struct {
		OrderID int64  `json:"order_id" binding:"required"`
		Amount  string `json:"amount" binding:"required"`
		Reason  string `json:"reason"`
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

	operator := "admin"
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		operator = fmt.Sprintf("admin:%d", subject.UserID)
	}

	if err := h.orderService.RefundOrder(c.Request.Context(), service.RefundOrderRequest{
		OrderID:  req.OrderID,
		Amount:   amount,
		Reason:   req.Reason,
		Operator: operator,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Refund processed"})
}
