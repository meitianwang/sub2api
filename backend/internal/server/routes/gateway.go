package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterGatewayRoutes 注册 API 网关路由（透传模式：所有请求原样转发到上游中转站）
func RegisterGatewayRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
) {
	bodyLimit := middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize)
	clientRequestID := middleware.ClientRequestID()
	opsErrorLogger := handler.OpsErrorLoggerMiddleware(opsService)
	endpointNorm := handler.InboundEndpointMiddleware()

	requireGroup := middleware.RequireGroupAssignment(settingService, middleware.AnthropicErrorWriter)
	requireGroupGoogle := middleware.RequireGroupAssignment(settingService, middleware.GoogleErrorWriter)

	// ===== Anthropic API 兼容 =====
	gateway := r.Group("/v1")
	gateway.Use(bodyLimit, clientRequestID, opsErrorLogger, endpointNorm)
	gateway.Use(gin.HandlerFunc(apiKeyAuth))
	gateway.Use(requireGroup)
	{
		gateway.POST("/messages", h.Gateway.Messages)
		gateway.POST("/messages/count_tokens", h.Gateway.CountTokens)
		gateway.GET("/models", h.Gateway.Models)
		gateway.GET("/usage", h.Gateway.Usage)

		// OpenAI Chat Completions API
		gateway.POST("/chat/completions", h.Gateway.ChatCompletions)

		// OpenAI Responses API
		gateway.POST("/responses", h.Gateway.Responses)
		gateway.POST("/responses/*subpath", h.Gateway.Responses)
	}

	// ===== Gemini 原生 API 兼容 =====
	// Use Google-style auth: supports ?key=xxx query param, x-goog-api-key header, and Bearer token.
	// Returns Google-style errors for Gemini SDK compatibility.
	gemini := r.Group("/v1beta")
	gemini.Use(bodyLimit, clientRequestID, opsErrorLogger, endpointNorm)
	gemini.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	gemini.Use(requireGroupGoogle)
	{
		gemini.GET("/models", h.Gateway.GeminiV1BetaListModels)
		gemini.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		gemini.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}

	// ===== 不带 v1 前缀的别名 =====
	commonMiddleware := []gin.HandlerFunc{bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroup}
	r.POST("/chat/completions", append(commonMiddleware, h.Gateway.ChatCompletions)...)
	r.POST("/responses", append(commonMiddleware, h.Gateway.Responses)...)
	r.POST("/responses/*subpath", append(commonMiddleware, h.Gateway.Responses)...)
}
