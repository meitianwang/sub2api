package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ──────────────────────────────────────────────────────────
// Canonical inbound / upstream endpoint paths.
// All normalization and derivation reference this single set
// of constants — add new paths HERE when a new API surface
// is introduced.
// ──────────────────────────────────────────────────────────

const (
	EndpointMessages        = "/v1/messages"
	EndpointChatCompletions = "/v1/chat/completions"
	EndpointResponses       = "/v1/responses"
	EndpointGeminiModels    = "/v1beta/models"
)

// gin.Context keys used by the middleware and helpers below.
const (
	ctxKeyInboundEndpoint = "_gateway_inbound_endpoint"
)

// ──────────────────────────────────────────────────────────
// Normalization functions
// ──────────────────────────────────────────────────────────

// NormalizeInboundEndpoint maps a raw request path to its canonical form.
//
//	"/v1/chat/completions"       → "/v1/chat/completions"
//	"/v1/responses/foo"          → "/v1/responses"
//	"/v1beta/models/gemini:gen"  → "/v1beta/models"
func NormalizeInboundEndpoint(path string) string {
	path = strings.TrimSpace(path)
	switch {
	case strings.Contains(path, EndpointChatCompletions):
		return EndpointChatCompletions
	case strings.Contains(path, EndpointMessages):
		return EndpointMessages
	case strings.Contains(path, EndpointResponses):
		return EndpointResponses
	case strings.Contains(path, EndpointGeminiModels):
		return EndpointGeminiModels
	default:
		return path
	}
}

// DeriveUpstreamEndpoint determines the upstream endpoint from the
// account platform and the normalized inbound endpoint.
// In passthrough mode, the upstream endpoint matches the inbound endpoint.
func DeriveUpstreamEndpoint(inbound, rawRequestPath, platform string) string {
	inbound = strings.TrimSpace(inbound)

	// Preserve subresource suffix for responses (e.g. /v1/responses/compact).
	if inbound == EndpointResponses {
		if suffix := responsesSubpathSuffix(rawRequestPath); suffix != "" {
			return EndpointResponses + suffix
		}
	}

	return inbound
}

// responsesSubpathSuffix extracts the part after "/responses" in a raw
// request path, e.g. "/openai/v1/responses/compact" → "/compact".
// Returns "" when there is no meaningful suffix.
func responsesSubpathSuffix(rawPath string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(rawPath), "/")
	idx := strings.LastIndex(trimmed, "/responses")
	if idx < 0 {
		return ""
	}
	suffix := trimmed[idx+len("/responses"):]
	if suffix == "" || suffix == "/" {
		return ""
	}
	if !strings.HasPrefix(suffix, "/") {
		return ""
	}
	return suffix
}

// ──────────────────────────────────────────────────────────
// Middleware
// ──────────────────────────────────────────────────────────

// InboundEndpointMiddleware normalizes the request path and stores the
// canonical inbound endpoint in gin.Context so that every handler in
// the chain can read it via GetInboundEndpoint.
//
// Apply this middleware to all gateway route groups.
func InboundEndpointMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" && c.Request != nil && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
		c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(path))
		c.Next()
	}
}

// ──────────────────────────────────────────────────────────
// Context helpers — used by handlers before building
// RecordUsageInput / RecordUsageLongContextInput.
// ──────────────────────────────────────────────────────────

// GetInboundEndpoint returns the canonical inbound endpoint stored by
// InboundEndpointMiddleware. If the middleware did not run (e.g. in
// tests), it falls back to normalizing c.FullPath() on the fly.
func GetInboundEndpoint(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyInboundEndpoint); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	// Fallback: normalize on the fly.
	path := ""
	if c != nil {
		path = c.FullPath()
		if path == "" && c.Request != nil && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
	}
	return NormalizeInboundEndpoint(path)
}

// GetUpstreamEndpoint derives the upstream endpoint from the context
// and the account platform. Handlers call this after scheduling an
// account, passing account.Platform.
func GetUpstreamEndpoint(c *gin.Context, platform string) string {
	inbound := GetInboundEndpoint(c)
	rawPath := ""
	if c != nil && c.Request != nil && c.Request.URL != nil {
		rawPath = c.Request.URL.Path
	}
	return DeriveUpstreamEndpoint(inbound, rawPath, platform)
}
