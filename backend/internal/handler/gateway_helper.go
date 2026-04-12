package handler

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// claudeCodeValidator is a singleton validator for Claude Code client detection
var claudeCodeValidator = service.NewClaudeCodeValidator()

const claudeCodeParsedRequestContextKey = "claude_code_parsed_request"

// SetClaudeCodeClientContext 检查请求是否来自 Claude Code 客户端，并设置到 context 中
func SetClaudeCodeClientContext(c *gin.Context, body []byte, parsedReq *service.ParsedRequest) {
	if c == nil || c.Request == nil {
		return
	}
	if parsedReq != nil {
		c.Set(claudeCodeParsedRequestContextKey, parsedReq)
	}

	ua := c.GetHeader("User-Agent")
	if !claudeCodeValidator.ValidateUserAgent(ua) {
		ctx := service.SetClaudeCodeClient(c.Request.Context(), false)
		c.Request = c.Request.WithContext(ctx)
		return
	}

	isClaudeCode := false
	if !strings.Contains(c.Request.URL.Path, "messages") {
		isClaudeCode = true
	} else {
		bodyMap := claudeCodeBodyMapFromParsedRequest(parsedReq)
		if bodyMap == nil {
			bodyMap = claudeCodeBodyMapFromContextCache(c)
		}
		if bodyMap == nil && len(body) > 0 {
			_ = json.Unmarshal(body, &bodyMap)
		}
		isClaudeCode = claudeCodeValidator.Validate(c.Request, bodyMap)
	}

	ctx := service.SetClaudeCodeClient(c.Request.Context(), isClaudeCode)

	if isClaudeCode {
		if version := claudeCodeValidator.ExtractVersion(ua); version != "" {
			ctx = service.SetClaudeCodeVersion(ctx, version)
		}
	}

	c.Request = c.Request.WithContext(ctx)
}

func claudeCodeBodyMapFromParsedRequest(parsedReq *service.ParsedRequest) map[string]any {
	if parsedReq == nil {
		return nil
	}
	bodyMap := map[string]any{
		"model": parsedReq.Model,
	}
	if parsedReq.System != nil || parsedReq.HasSystem {
		bodyMap["system"] = parsedReq.System
	}
	if parsedReq.MetadataUserID != "" {
		bodyMap["metadata"] = map[string]any{"user_id": parsedReq.MetadataUserID}
	}
	return bodyMap
}

func claudeCodeBodyMapFromContextCache(c *gin.Context) map[string]any {
	if c == nil {
		return nil
	}
	if cached, ok := c.Get(service.OpenAIParsedRequestBodyKey); ok {
		if bodyMap, ok := cached.(map[string]any); ok {
			return bodyMap
		}
	}
	if cached, ok := c.Get(claudeCodeParsedRequestContextKey); ok {
		switch v := cached.(type) {
		case *service.ParsedRequest:
			return claudeCodeBodyMapFromParsedRequest(v)
		case service.ParsedRequest:
			return claudeCodeBodyMapFromParsedRequest(&v)
		}
	}
	return nil
}

// SSEPingFormat defines the format of SSE ping events for different platforms
type SSEPingFormat string

const (
	SSEPingFormatClaude  SSEPingFormat = "data: {\"type\": \"ping\"}\n\n"
	SSEPingFormatNone    SSEPingFormat = ""
	SSEPingFormatComment SSEPingFormat = ":\n\n"
)

const (
	defaultPingInterval = 10 * time.Second
	initialBackoff      = 100 * time.Millisecond
	backoffMultiplier   = 1.5
	maxBackoff          = 2 * time.Second
)

// wrapReleaseOnDone ensures release runs at most once and still triggers on context cancellation.
func wrapReleaseOnDone(ctx context.Context, releaseFunc func()) func() {
	if releaseFunc == nil {
		return nil
	}
	var once sync.Once

	release := func() {
		once.Do(func() {
			releaseFunc()
		})
	}

	stop := context.AfterFunc(ctx, release)
	return func() {
		once.Do(func() {
			stop()
			releaseFunc()
		})
	}
}

// nextBackoff calculates the next backoff duration with jitter.
func nextBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * backoffMultiplier)
	if next > maxBackoff {
		next = maxBackoff
	}
	jitter := 0.8 + rand.Float64()*0.4
	jittered := time.Duration(float64(next) * jitter)
	if jittered < initialBackoff {
		return initialBackoff
	}
	if jittered > maxBackoff {
		return maxBackoff
	}
	return jittered
}
