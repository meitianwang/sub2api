package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// GeminiV1BetaListModels proxies GET /v1beta/models to upstream relay.
func (h *GatewayHandler) GeminiV1BetaListModels(c *gin.Context) {
	h.geminiPassthrough(c, "/v1beta/models", nil, false, "")
}

// GeminiV1BetaGetModel proxies GET /v1beta/models/:model to upstream relay.
func (h *GatewayHandler) GeminiV1BetaGetModel(c *gin.Context) {
	modelName := c.Param("model")
	h.geminiPassthrough(c, "/v1beta/models/"+modelName, nil, false, modelName)
}

// GeminiV1BetaModels proxies POST /v1beta/models/{model}:{action} to upstream relay.
func (h *GatewayHandler) GeminiV1BetaModels(c *gin.Context) {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil {
		googleError(c, http.StatusUnauthorized, "Invalid API key")
		return
	}
	authSubject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		googleError(c, http.StatusInternalServerError, "User context not found")
		return
	}
	reqLog := requestLogger(c, "handler.gemini_v1beta.models",
		zap.Int64("user_id", authSubject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)

	modelAction := strings.TrimPrefix(c.Param("modelAction"), "/")
	modelName, action, err := parseGeminiModelAction(modelAction)
	if err != nil {
		googleError(c, http.StatusNotFound, err.Error())
		return
	}

	stream := action == "streamGenerateContent"
	reqLog = reqLog.With(zap.String("model", modelName), zap.String("action", action), zap.Bool("stream", stream))

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			googleError(c, http.StatusRequestEntityTooLarge, buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		googleError(c, http.StatusBadRequest, "Failed to read request body")
		return
	}
	if len(body) == 0 {
		googleError(c, http.StatusBadRequest, "Request body is empty")
		return
	}

	setOpsRequestContext(c, modelName, stream, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeFromLegacy(stream, false)))

	subscription, _ := middleware.GetSubscriptionFromContext(c)

	// Billing check
	requestStart := time.Now()
	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription); err != nil {
		status, _, message := billingErrorDetails(err)
		googleError(c, status, message)
		return
	}

	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())

	// Build parsed request
	parsedReq := &service.ParsedRequest{
		Model:  modelName,
		Stream: stream,
		Body:   body,
	}

	// Build the original path for passthrough
	originalPath := fmt.Sprintf("/v1beta/models/%s:%s", modelName, action)

	// Session hash
	parsedReq.SessionContext = &service.SessionContext{
		ClientIP:  ip.GetClientIP(c),
		UserAgent: c.GetHeader("User-Agent"),
		APIKeyID:  apiKey.ID,
	}
	sessionHash := h.gatewayService.GenerateSessionHash(parsedReq)

	// Account selection + failover loop
	fs := NewFailoverState(h.maxAccountSwitches, false)

	for {
		selection, err := h.gatewayService.SelectAccountWithLoadAwareness(c.Request.Context(), apiKey.GroupID, sessionHash, modelName, fs.FailedAccountIDs, "")
		if err != nil {
			if len(fs.FailedAccountIDs) == 0 {
				googleError(c, http.StatusServiceUnavailable, "No available accounts: "+err.Error())
				return
			}
			action := fs.HandleSelectionExhausted(c.Request.Context())
			switch action {
			case FailoverContinue:
				continue
			case FailoverCanceled:
				return
			default:
				googleError(c, http.StatusBadGateway, "All available accounts exhausted")
				return
			}
		}
		account := selection.Account
		setOpsSelectedAccount(c, account.ID, account.Platform)

		// Account already selected, proceed directly
		accountReleaseFunc := selection.ReleaseFunc
		if !selection.Acquired {
			googleError(c, http.StatusServiceUnavailable, "No available accounts")
			return
		}
		accountReleaseFunc = wrapReleaseOnDone(c.Request.Context(), accountReleaseFunc)

		// Forward request via passthrough
		writerSizeBeforeForward := c.Writer.Size()
		result, err := h.gatewayService.ForwardPassthrough(c.Request.Context(), c, account, originalPath, body, parsedReq)

		if accountReleaseFunc != nil {
			accountReleaseFunc()
		}

		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				if c.Writer.Size() != writerSizeBeforeForward {
					googleError(c, http.StatusBadGateway, "Upstream error after partial response")
					return
				}
				action := fs.HandleFailoverError(c.Request.Context(), h.gatewayService, account.ID, account.Platform, failoverErr)
				switch action {
				case FailoverContinue:
					continue
				case FailoverExhausted:
					googleError(c, http.StatusBadGateway, "All available accounts exhausted")
					return
				case FailoverCanceled:
					return
				}
			}
			reqLog.Error("gemini.forward_failed", zap.Int64("account_id", account.ID), zap.Error(err))
			return
		}

		// Record usage
		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)
		requestPayloadHash := service.HashUsageRequestPayload(body)
		inboundEndpoint := GetInboundEndpoint(c)
		upstreamEndpoint := GetUpstreamEndpoint(c, account.Platform)

		h.submitUsageRecordTask(func(ctx context.Context) {
			if err := h.gatewayService.RecordUsage(ctx, &service.RecordUsageInput{
				Result:             result,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				Subscription:       subscription,
				InboundEndpoint:    inboundEndpoint,
				UpstreamEndpoint:   upstreamEndpoint,
				UserAgent:          userAgent,
				IPAddress:          clientIP,
				RequestPayloadHash: requestPayloadHash,
				APIKeyService:      h.apiKeyService,
			}); err != nil {
				reqLog.Error("gemini.record_usage_failed", zap.Int64("account_id", account.ID), zap.Error(err))
			}
		})
		return
	}
}

// geminiPassthrough handles GET requests for Gemini model listing/info via passthrough.
func (h *GatewayHandler) geminiPassthrough(c *gin.Context, path string, body []byte, stream bool, model string) {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil {
		googleError(c, http.StatusUnauthorized, "Invalid API key")
		return
	}

	parsedReq := &service.ParsedRequest{
		Model:  model,
		Stream: stream,
		Body:   body,
	}

	selection, err := h.gatewayService.SelectAccountWithLoadAwareness(c.Request.Context(), apiKey.GroupID, "", model, nil, "")
	if err != nil {
		googleError(c, http.StatusServiceUnavailable, "No available accounts: "+err.Error())
		return
	}
	account := selection.Account
	if selection.ReleaseFunc != nil {
		defer selection.ReleaseFunc()
	}

	_, err = h.gatewayService.ForwardPassthrough(c.Request.Context(), c, account, path, body, parsedReq)
	if err != nil {
		// Error response already written by ForwardPassthrough
		return
	}
}

func parseGeminiModelAction(raw string) (model, action string, err error) {
	idx := strings.LastIndex(raw, ":")
	if idx < 0 || idx == len(raw)-1 {
		return "", "", fmt.Errorf("invalid model action: %s", raw)
	}
	return raw[:idx], raw[idx+1:], nil
}

func googleError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    status,
			"message": message,
			"status":  http.StatusText(status),
		},
	})
}

// parseGeminiModelFromBody extracts model from Gemini request body (unused in passthrough mode but kept for reference)
func parseGeminiModelFromBody(body []byte) string {
	return gjson.GetBytes(body, "model").String()
}
