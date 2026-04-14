package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentDashboardHandler handles admin payment dashboard statistics.
type PaymentDashboardHandler struct {
	orderService *service.PaymentOrderService
}

// NewPaymentDashboardHandler creates a new PaymentDashboardHandler.
func NewPaymentDashboardHandler(orderService *service.PaymentOrderService) *PaymentDashboardHandler {
	return &PaymentDashboardHandler{orderService: orderService}
}

// GetStats handles GET /api/v1/admin/pay/dashboard
func (h *PaymentDashboardHandler) GetStats(c *gin.Context) {
	days := 30
	if v := c.Query("days"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	stats, err := h.orderService.GetDashboardStats(c.Request.Context(), days)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	dailySeries := stats.DailySeries
	if dailySeries == nil {
		dailySeries = []service.DailySeriesPoint{}
	}
	paymentMethods := stats.PaymentMethods
	if paymentMethods == nil {
		paymentMethods = []service.PaymentMethodStat{}
	}

	response.Success(c, dto.PaymentDashboardDTO{
		TodayAmount:     stats.TodayAmount.String(),
		TodayOrderCount: stats.TodayOrderCount,
		TotalAmount:     stats.TotalAmount.String(),
		TotalOrderCount: stats.TotalOrderCount,
		DailySeries:     dailySeries,
		PaymentMethods:  paymentMethods,
		Leaderboard:     stats.Leaderboard,
	})
}
