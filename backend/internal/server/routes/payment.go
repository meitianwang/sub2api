package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPaymentWebhookRoutes registers unauthenticated payment webhook endpoints.
func RegisterPaymentWebhookRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	notify := v1.Group("/pay/notify")
	{
		notify.GET("/easypay", h.PaymentWebhook.NotifyEasyPay)
		notify.POST("/alipay", h.PaymentWebhook.NotifyAlipay)
		notify.POST("/wxpay", h.PaymentWebhook.NotifyWxpay)
		notify.POST("/stripe", h.PaymentWebhook.NotifyStripe)
	}
}

// registerPaymentUserRoutes registers authenticated user payment endpoints.
func registerPaymentUserRoutes(authenticated *gin.RouterGroup, h *handler.Handlers) {
	pay := authenticated.Group("/pay")
	{
		pay.POST("/orders", h.Payment.CreateOrder)
		pay.GET("/orders", h.Payment.ListOrders)
		pay.GET("/orders/:id", h.Payment.GetOrder)
		pay.POST("/orders/:id/cancel", h.Payment.CancelOrder)
		pay.POST("/orders/:id/refund-request", h.Payment.RequestRefund)
		pay.GET("/config", h.Payment.GetConfig)
		pay.GET("/channels", h.Payment.ListChannels)
		pay.GET("/subscription-plans", h.Payment.ListPlans)
	}
}

// registerPaymentAdminRoutes registers admin payment management endpoints.
func registerPaymentAdminRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	pay := admin.Group("/pay")
	{
		// Orders
		orders := pay.Group("/orders")
		{
			orders.GET("", h.Admin.PaymentOrder.List)
			orders.GET("/:id", h.Admin.PaymentOrder.GetByID)
			orders.POST("/:id/cancel", h.Admin.PaymentOrder.Cancel)
			orders.POST("/:id/retry", h.Admin.PaymentOrder.RetryRecharge)
		}

		// Refund
		pay.POST("/refund", h.Admin.PaymentRefund.ProcessRefund)

		// Config
		config := pay.Group("/config")
		{
			config.GET("", h.Admin.PaymentConfig.Get)
			config.PUT("", h.Admin.PaymentConfig.Update)
		}

		// Provider instances
		instances := pay.Group("/provider-instances")
		{
			instances.GET("", h.Admin.PaymentProviderInstance.List)
			instances.POST("", h.Admin.PaymentProviderInstance.Create)
			instances.GET("/:id", h.Admin.PaymentProviderInstance.GetByID)
			instances.PUT("/:id", h.Admin.PaymentProviderInstance.Update)
			instances.DELETE("/:id", h.Admin.PaymentProviderInstance.Delete)
		}

		// Channels
		channels := pay.Group("/channels")
		{
			channels.GET("", h.Admin.PaymentChannel.List)
			channels.POST("", h.Admin.PaymentChannel.Create)
			channels.PUT("/:id", h.Admin.PaymentChannel.Update)
			channels.DELETE("/:id", h.Admin.PaymentChannel.Delete)
		}

		// Subscription plans
		plans := pay.Group("/subscription-plans")
		{
			plans.GET("", h.Admin.PaymentSubscriptionPlan.List)
			plans.POST("", h.Admin.PaymentSubscriptionPlan.Create)
			plans.PUT("/:id", h.Admin.PaymentSubscriptionPlan.Update)
			plans.DELETE("/:id", h.Admin.PaymentSubscriptionPlan.Delete)
		}

		// Dashboard
		pay.GET("/dashboard", h.Admin.PaymentDashboard.GetStats)
	}
}
