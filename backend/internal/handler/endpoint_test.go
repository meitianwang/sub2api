package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

// ──────────────────────────────────────────────────────────
// NormalizeInboundEndpoint
// ──────────────────────────────────────────────────────────

func TestNormalizeInboundEndpoint(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// Direct canonical paths.
		{"/v1/messages", EndpointMessages},
		{"/v1/chat/completions", EndpointChatCompletions},
		{"/v1/responses", EndpointResponses},
		{"/v1beta/models", EndpointGeminiModels},

		// Gin route patterns with wildcards.
		{"/v1beta/models/*modelAction", EndpointGeminiModels},
		{"/v1/responses/*subpath", EndpointResponses},

		// Unknown path is returned as-is.
		{"/v1/embeddings", "/v1/embeddings"},
		{"", ""},
		{"  /v1/messages  ", EndpointMessages},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			require.Equal(t, tt.want, NormalizeInboundEndpoint(tt.path))
		})
	}
}

// ──────────────────────────────────────────────────────────
// DeriveUpstreamEndpoint — passthrough mode returns inbound
// ──────────────────────────────────────────────────────────

func TestDeriveUpstreamEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		inbound  string
		rawPath  string
		platform string
		want     string
	}{
		{"messages passthrough", EndpointMessages, "/v1/messages", "anthropic", EndpointMessages},
		{"chat completions passthrough", EndpointChatCompletions, "/v1/chat/completions", "anthropic", EndpointChatCompletions},
		{"gemini models passthrough", EndpointGeminiModels, "/v1beta/models/gemini:gen", "gemini", EndpointGeminiModels},
		{"responses root", EndpointResponses, "/v1/responses", "anthropic", EndpointResponses},
		{"responses compact subpath", EndpointResponses, "/v1/responses/compact", "anthropic", "/v1/responses/compact"},
		{"responses nested subpath", EndpointResponses, "/v1/responses/compact/detail", "anthropic", "/v1/responses/compact/detail"},
		{"unknown platform passthrough", "/v1/embeddings", "/v1/embeddings", "unknown", "/v1/embeddings"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, DeriveUpstreamEndpoint(tt.inbound, tt.rawPath, tt.platform))
		})
	}
}

// ──────────────────────────────────────────────────────────
// responsesSubpathSuffix
// ──────────────────────────────────────────────────────────

func TestResponsesSubpathSuffix(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"/v1/responses", ""},
		{"/v1/responses/", ""},
		{"/v1/responses/compact", "/compact"},
		{"/v1/responses/compact/detail", "/compact/detail"},
		{"/v1/messages", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			require.Equal(t, tt.want, responsesSubpathSuffix(tt.raw))
		})
	}
}

// ──────────────────────────────────────────────────────────
// InboundEndpointMiddleware + context helpers
// ──────────────────────────────────────────────────────────

func TestInboundEndpointMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(InboundEndpointMiddleware())

	var captured string
	router.POST("/v1/messages", func(c *gin.Context) {
		captured = GetInboundEndpoint(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, EndpointMessages, captured)
}

func TestGetInboundEndpoint_FallbackWithoutMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	got := GetInboundEndpoint(c)
	require.Equal(t, EndpointMessages, got)
}

func TestGetUpstreamEndpoint_FullFlow(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", nil)

	c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(c.Request.URL.Path))

	got := GetUpstreamEndpoint(c, "anthropic")
	require.Equal(t, "/v1/responses/compact", got)
}
