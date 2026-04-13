package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// passthroughProtocol identifies the upstream API protocol for auth and usage parsing.
type passthroughProtocol int

const (
	protocolAnthropic passthroughProtocol = iota
	protocolOpenAI
	protocolGemini
)

// detectProtocol determines the upstream protocol from the request path.
func detectProtocol(path string) passthroughProtocol {
	if strings.HasPrefix(path, "/v1beta/") {
		return protocolGemini
	}
	if strings.HasPrefix(path, "/v1/chat/completions") ||
		strings.HasPrefix(path, "/v1/responses") ||
		strings.HasPrefix(path, "/chat/completions") ||
		strings.HasPrefix(path, "/responses") ||
		strings.HasPrefix(path, "/v1/embeddings") ||
		strings.HasPrefix(path, "/v1/images/") {
		return protocolOpenAI
	}
	// Default: Anthropic (/v1/messages, /v1/messages/count_tokens, /v1/models, /v1/usage, etc.)
	return protocolAnthropic
}

// ForwardPassthrough transparently proxies a request to the upstream relay service.
// It replaces authentication, optionally maps the model name, and extracts usage
// from the response for billing. The request body and response are forwarded as-is.
func (s *GatewayService) ForwardPassthrough(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	originalPath string,
	body []byte,
	parsed *ParsedRequest,
) (*ForwardResult, error) {
	startTime := time.Now()
	if parsed == nil {
		return nil, fmt.Errorf("passthrough: nil parsed request")
	}

	protocol := detectProtocol(originalPath)

	// 1. Model mapping
	originalModel := parsed.Model
	mappedModel := originalModel
	if originalModel != "" {
		if m := account.GetMappedModel(originalModel); m != originalModel {
			body = s.replaceModelInBody(body, m)
			mappedModel = m
		}
	}

	// 2. Get API key
	apiKey := account.GetCredential("api_key")
	if apiKey == "" {
		return nil, errors.New("passthrough: api_key not found in account credentials")
	}

	// 3. Build upstream URL
	baseURL := account.GetBaseURL()
	if baseURL == "" {
		return nil, errors.New("passthrough: base_url not configured on account")
	}
	baseURL = strings.TrimRight(baseURL, "/")

	targetURL := baseURL + originalPath
	// Preserve query string from the original request
	if c != nil && c.Request != nil && c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// For Gemini protocol, append API key as query param
	if protocol == protocolGemini {
		sep := "?"
		if strings.Contains(targetURL, "?") {
			sep = "&"
		}
		targetURL += sep + "key=" + apiKey
	}

	// 4. Build upstream HTTP request
	// For streaming: detach context so client disconnect doesn't cancel upstream read
	// For non-streaming: use the original context directly
	upstreamCtx := ctx
	if parsed.Stream {
		var cancel context.CancelFunc
		upstreamCtx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}
	req, err := http.NewRequestWithContext(upstreamCtx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("passthrough: build request: %w", err)
	}

	// For GET requests (e.g., /v1/models, /v1beta/models), use GET method
	if c != nil && c.Request != nil && c.Request.Method == http.MethodGet {
		req.Method = http.MethodGet
		req.Body = nil
		req.ContentLength = 0
	}

	// Copy allowed headers from inbound request
	if c != nil && c.Request != nil {
		for key, values := range c.Request.Header {
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if !allowedHeaders[lowerKey] {
				continue
			}
			wireKey := resolveWireCasing(key)
			for _, v := range values {
				addHeaderRaw(req.Header, wireKey, v)
			}
		}
	}

	// 5. Set authentication header (remove any inbound auth first)
	req.Header.Del("authorization")
	req.Header.Del("x-api-key")
	req.Header.Del("x-goog-api-key")
	req.Header.Del("cookie")

	switch protocol {
	case protocolAnthropic:
		setHeaderRaw(req.Header, "x-api-key", apiKey)
		if getHeaderRaw(req.Header, "anthropic-version") == "" {
			setHeaderRaw(req.Header, "anthropic-version", "2023-06-01")
		}
	case protocolOpenAI:
		setHeaderRaw(req.Header, "authorization", "Bearer "+apiKey)
	case protocolGemini:
		// Auth via query param (already appended to URL above)
	}

	if getHeaderRaw(req.Header, "content-type") == "" {
		setHeaderRaw(req.Header, "content-type", "application/json")
	}

	// OPS: record upstream request body
	setOpsUpstreamRequestBody(c, body)

	// 6. Get proxy URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	// 7. Send request with retry
	var resp *http.Response
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		resp, err = s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, s.tlsFPProfileService.ResolveTLSProfile(account))
		if err != nil {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			safeErr := sanitizeUpstreamErrorMessage(err.Error())
			setOpsUpstreamError(c, 0, safeErr, "")
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:    "anthropic",
				AccountID:   account.ID,
				AccountName: account.Name,
				Kind:        "network",
				Message:     safeErr,
			})
			return nil, fmt.Errorf("passthrough: upstream request failed: %w", err)
		}
		if resp.StatusCode < 400 {
			break
		}

		// Check for retryable errors
		if s.shouldRetryUpstreamError(account, resp.StatusCode) && attempt < maxRetryAttempts {
			_ = resp.Body.Close()
			delay := retryBackoffDelay(attempt)
			logger.LegacyPrintf("service.gateway", "[Passthrough] Retrying: account=%d status=%d attempt=%d delay=%v",
				account.ID, resp.StatusCode, attempt, delay)
			time.Sleep(delay)

			// Rebuild request for retry (reuse same upstreamCtx)
			req, err = http.NewRequestWithContext(upstreamCtx, req.Method, targetURL, bytes.NewReader(body))
			if err != nil {
				return nil, fmt.Errorf("passthrough: rebuild request for retry: %w", err)
			}
			// Re-set headers for retry
			if c != nil && c.Request != nil {
				for key, values := range c.Request.Header {
					lowerKey := strings.ToLower(strings.TrimSpace(key))
					if !allowedHeaders[lowerKey] {
						continue
					}
					wireKey := resolveWireCasing(key)
					for _, v := range values {
						addHeaderRaw(req.Header, wireKey, v)
					}
				}
			}
			req.Header.Del("authorization")
			req.Header.Del("x-api-key")
			req.Header.Del("x-goog-api-key")
			req.Header.Del("cookie")
			switch protocol {
			case protocolAnthropic:
				setHeaderRaw(req.Header, "x-api-key", apiKey)
				if getHeaderRaw(req.Header, "anthropic-version") == "" {
					setHeaderRaw(req.Header, "anthropic-version", "2023-06-01")
				}
			case protocolOpenAI:
				setHeaderRaw(req.Header, "authorization", "Bearer "+apiKey)
			}
			if getHeaderRaw(req.Header, "content-type") == "" {
				setHeaderRaw(req.Header, "content-type", "application/json")
			}
			continue
		}

		// Check for failover errors
		if s.shouldFailoverUpstreamError(resp.StatusCode) {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()

			logger.LegacyPrintf("service.gateway", "[Passthrough] Upstream failover error: account=%d status=%d body=%s",
				account.ID, resp.StatusCode, truncateString(string(respBody), 1000))

			s.handleFailoverSideEffects(ctx, resp, account)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           "anthropic",
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Passthrough:        true,
				Kind:               "failover",
				Message:            extractUpstreamErrorMessage(respBody),
			})
			return nil, &UpstreamFailoverError{
				StatusCode:   resp.StatusCode,
				ResponseBody: respBody,
			}
		}

		// Non-retryable, non-failover error: pass through to client
		break
	}
	defer resp.Body.Close()

	// 8. Handle response
	if resp.StatusCode >= 400 {
		return s.handlePassthroughErrorResponse(ctx, resp, c, account)
	}

	var usage *ClaudeUsage
	var firstTokenMs *int
	clientDisconnect := false

	if parsed.Stream {
		result, err := s.handlePassthroughStreamingResponse(ctx, resp, c, account, startTime, protocol)
		if err != nil {
			return nil, err
		}
		usage = result.usage
		firstTokenMs = result.firstTokenMs
		clientDisconnect = result.clientDisconnect
	} else {
		usage, err = s.handlePassthroughNonStreamingResponse(ctx, resp, c, account, protocol)
		if err != nil {
			return nil, err
		}
	}

	if usage == nil {
		usage = &ClaudeUsage{}
	}

	upstreamModel := ""
	if mappedModel != originalModel {
		upstreamModel = mappedModel
	}

	return &ForwardResult{
		RequestID:        resp.Header.Get("x-request-id"),
		Usage:            *usage,
		Model:            originalModel,
		UpstreamModel:    upstreamModel,
		Stream:           parsed.Stream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
	}, nil
}

// handlePassthroughErrorResponse forwards an error response from upstream to the client as-is.
func (s *GatewayService) handlePassthroughErrorResponse(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
) (*ForwardResult, error) {
	if s.rateLimitService != nil {
		s.rateLimitService.UpdateSessionWindow(ctx, account, resp.Header)
	}

	maxBytes := resolveUpstreamResponseReadLimit(s.cfg)
	body, err := readUpstreamResponseBodyLimited(resp.Body, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("passthrough: read error response: %w", err)
	}

	setOpsUpstreamError(c, resp.StatusCode, extractUpstreamErrorMessage(body), "")
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           "anthropic",
		AccountID:          account.ID,
		AccountName:        account.Name,
		UpstreamStatusCode: resp.StatusCode,
		UpstreamRequestID:  resp.Header.Get("x-request-id"),
		Passthrough:        true,
		Kind:               "error",
		Message:            extractUpstreamErrorMessage(body),
	})

	// Forward error response headers and body as-is
	writePassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, body)
	return nil, fmt.Errorf("upstream error: %d", resp.StatusCode)
}

// handlePassthroughStreamingResponse forwards a streaming response from upstream to the client,
// sniffing usage information for billing along the way.
func (s *GatewayService) handlePassthroughStreamingResponse(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
	startTime time.Time,
	protocol passthroughProtocol,
) (*streamingResult, error) {
	if s.rateLimitService != nil {
		s.rateLimitService.UpdateSessionWindow(ctx, account, resp.Header)
	}

	writePassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "text/event-stream"
	}
	c.Header("Content-Type", contentType)
	if c.Writer.Header().Get("Cache-Control") == "" {
		c.Header("Cache-Control", "no-cache")
	}
	if c.Writer.Header().Get("Connection") == "" {
		c.Header("Connection", "keep-alive")
	}
	c.Header("X-Accel-Buffering", "no")
	if v := resp.Header.Get("x-request-id"); v != "" {
		c.Header("x-request-id", v)
	}

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	usage := &ClaudeUsage{}
	var firstTokenMs *int
	clientDisconnected := false
	sawTerminalEvent := false

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)

	type scanEvent struct {
		line string
		err  error
	}
	events := make(chan scanEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev scanEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}
	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func(buf *sseScannerBuf64K) {
		defer putSSEScannerBuf64K(buf)
		defer close(events)
		for scanner.Scan() {
			atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			if !sendEvent(scanEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			_ = sendEvent(scanEvent{err: err})
		}
	}(scanBuf)
	defer close(done)

	streamInterval := time.Duration(0)
	if s.cfg != nil && s.cfg.Gateway.StreamDataIntervalTimeout > 0 {
		streamInterval = time.Duration(s.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
	}
	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	isTerminal := func(eventName, data string) bool {
		switch protocol {
		case protocolAnthropic:
			return anthropicStreamEventIsTerminal(eventName, data)
		case protocolOpenAI:
			return openAIStreamEventIsTerminal(data)
		case protocolGemini:
			return geminiStreamEventIsTerminal(data)
		default:
			return false
		}
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				if !clientDisconnected {
					flusher.Flush()
				}
				if !sawTerminalEvent {
					return &streamingResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: clientDisconnected},
						fmt.Errorf("passthrough: stream usage incomplete: missing terminal event")
				}
				return &streamingResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: clientDisconnected}, nil
			}
			if ev.err != nil {
				if sawTerminalEvent {
					return &streamingResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: clientDisconnected}, nil
				}
				if clientDisconnected {
					return &streamingResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: true},
						fmt.Errorf("passthrough: stream incomplete after disconnect: %w", ev.err)
				}
				if errors.Is(ev.err, bufio.ErrTooLong) {
					logger.L().Warn("passthrough.sse_line_too_long",
						zap.Int64("account_id", account.ID),
						zap.Int("max_size", maxLineSize),
						zap.Error(ev.err))
					return &streamingResult{usage: usage, firstTokenMs: firstTokenMs}, ev.err
				}
				return &streamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("passthrough: stream read error: %w", ev.err)
			}

			line := ev.line
			// Extract SSE data and parse usage + detect terminal event
			if data, ok := extractSSEDataLine(line); ok {
				trimmed := strings.TrimSpace(data)
				if isTerminal("", trimmed) {
					sawTerminalEvent = true
				}
				if firstTokenMs == nil && trimmed != "" && trimmed != "[DONE]" {
					ms := int(time.Since(startTime).Milliseconds())
					firstTokenMs = &ms
				}
				parsePassthroughUsage(protocol, data, usage)
			} else {
				trimmed := strings.TrimSpace(line)
				// Check event: lines for terminal detection
				if strings.HasPrefix(trimmed, "event:") {
					eventName := strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
					if isTerminal(eventName, "") {
						sawTerminalEvent = true
					}
				}
			}

			// Forward line to client
			if !clientDisconnected {
				if _, err := io.WriteString(w, line); err != nil {
					clientDisconnected = true
					logger.LegacyPrintf("service.gateway", "[Passthrough] Client disconnected, continue draining for usage: account=%d", account.ID)
				} else if _, err := io.WriteString(w, "\n"); err != nil {
					clientDisconnected = true
				} else if line == "" {
					flusher.Flush()
				}
			}

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			if clientDisconnected {
				return &streamingResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: true},
					fmt.Errorf("passthrough: stream timeout after disconnect")
			}
			logger.LegacyPrintf("service.gateway", "[Passthrough] Stream interval timeout: account=%d interval=%s", account.ID, streamInterval)
			if s.rateLimitService != nil {
				s.rateLimitService.HandleStreamTimeout(ctx, account, "")
			}
			return &streamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("passthrough: stream data interval timeout")
		}
	}
}

// handlePassthroughNonStreamingResponse forwards a non-streaming response from upstream.
func (s *GatewayService) handlePassthroughNonStreamingResponse(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
	protocol passthroughProtocol,
) (*ClaudeUsage, error) {
	if s.rateLimitService != nil {
		s.rateLimitService.UpdateSessionWindow(ctx, account, resp.Header)
	}

	maxBytes := resolveUpstreamResponseReadLimit(s.cfg)
	body, err := readUpstreamResponseBodyLimited(resp.Body, maxBytes)
	if err != nil {
		if errors.Is(err, ErrUpstreamResponseBodyTooLarge) {
			setOpsUpstreamError(c, http.StatusBadGateway, "upstream response too large", "")
			c.JSON(http.StatusBadGateway, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "upstream_error",
					"message": "Upstream response too large",
				},
			})
		}
		return nil, err
	}

	usage := parsePassthroughUsageFromBody(protocol, body)

	writePassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, body)
	return usage, nil
}

// extractSSEDataLine extracts the data payload from an SSE "data:" line.
func extractSSEDataLine(line string) (string, bool) {
	if !strings.HasPrefix(line, "data:") {
		return "", false
	}
	start := len("data:")
	for start < len(line) {
		if line[start] != ' ' && line[start] != '\t' {
			break
		}
		start++
	}
	return line[start:], true
}

// parsePassthroughUsage parses usage from an SSE data line based on the protocol.
func parsePassthroughUsage(protocol passthroughProtocol, data string, usage *ClaudeUsage) {
	if usage == nil || data == "" || data == "[DONE]" {
		return
	}

	switch protocol {
	case protocolAnthropic:
		parseAnthropicSSEUsage(data, usage)
	case protocolOpenAI:
		parseOpenAISSEUsage(data, usage)
	case protocolGemini:
		parseGeminiSSEUsage(data, usage)
	}
}

// parseAnthropicSSEUsage extracts usage from Anthropic SSE events.
func parseAnthropicSSEUsage(data string, usage *ClaudeUsage) {
	parsed := gjson.Parse(data)
	switch parsed.Get("type").String() {
	case "message_start":
		msgUsage := parsed.Get("message.usage")
		if msgUsage.Exists() {
			usage.InputTokens = int(msgUsage.Get("input_tokens").Int())
			usage.CacheCreationInputTokens = int(msgUsage.Get("cache_creation_input_tokens").Int())
			usage.CacheReadInputTokens = int(msgUsage.Get("cache_read_input_tokens").Int())

			cc5m := msgUsage.Get("cache_creation.ephemeral_5m_input_tokens")
			cc1h := msgUsage.Get("cache_creation.ephemeral_1h_input_tokens")
			if cc5m.Exists() || cc1h.Exists() {
				usage.CacheCreation5mTokens = int(cc5m.Int())
				usage.CacheCreation1hTokens = int(cc1h.Int())
			}
		}
	case "message_delta":
		deltaUsage := parsed.Get("usage")
		if deltaUsage.Exists() {
			if v := deltaUsage.Get("input_tokens").Int(); v > 0 {
				usage.InputTokens = int(v)
			}
			if v := deltaUsage.Get("output_tokens").Int(); v > 0 {
				usage.OutputTokens = int(v)
			}
			if v := deltaUsage.Get("cache_creation_input_tokens").Int(); v > 0 {
				usage.CacheCreationInputTokens = int(v)
			}
			if v := deltaUsage.Get("cache_read_input_tokens").Int(); v > 0 {
				usage.CacheReadInputTokens = int(v)
			}

			cc5m := deltaUsage.Get("cache_creation.ephemeral_5m_input_tokens")
			cc1h := deltaUsage.Get("cache_creation.ephemeral_1h_input_tokens")
			if cc5m.Exists() && cc5m.Int() > 0 {
				usage.CacheCreation5mTokens = int(cc5m.Int())
			}
			if cc1h.Exists() && cc1h.Int() > 0 {
				usage.CacheCreation1hTokens = int(cc1h.Int())
			}
		}
	}

	// Fallback: check cached_tokens
	if usage.CacheReadInputTokens == 0 {
		if cached := parsed.Get("message.usage.cached_tokens").Int(); cached > 0 {
			usage.CacheReadInputTokens = int(cached)
		}
		if cached := parsed.Get("usage.cached_tokens").Int(); usage.CacheReadInputTokens == 0 && cached > 0 {
			usage.CacheReadInputTokens = int(cached)
		}
	}
	if usage.CacheCreationInputTokens == 0 {
		cc5m := parsed.Get("message.usage.cache_creation.ephemeral_5m_input_tokens").Int()
		cc1h := parsed.Get("message.usage.cache_creation.ephemeral_1h_input_tokens").Int()
		if cc5m == 0 && cc1h == 0 {
			cc5m = parsed.Get("usage.cache_creation.ephemeral_5m_input_tokens").Int()
			cc1h = parsed.Get("usage.cache_creation.ephemeral_1h_input_tokens").Int()
		}
		total := cc5m + cc1h
		if total > 0 {
			usage.CacheCreationInputTokens = int(total)
		}
	}
}

// parseOpenAISSEUsage extracts usage from OpenAI SSE events.
// OpenAI includes usage in the final chunk when stream_options.include_usage is set.
func parseOpenAISSEUsage(data string, usage *ClaudeUsage) {
	parsed := gjson.Parse(data)
	usageNode := parsed.Get("usage")
	if !usageNode.Exists() {
		return
	}
	if v := usageNode.Get("prompt_tokens").Int(); v > 0 {
		usage.InputTokens = int(v)
	}
	if v := usageNode.Get("completion_tokens").Int(); v > 0 {
		usage.OutputTokens = int(v)
	}
	// Also handle prompt_tokens_details.cached_tokens
	if cached := usageNode.Get("prompt_tokens_details.cached_tokens").Int(); cached > 0 {
		usage.CacheReadInputTokens = int(cached)
	}
}

// parseGeminiSSEUsage extracts usage from a Gemini SSE data line.
func parseGeminiSSEUsage(data string, usage *ClaudeUsage) {
	parsed := gjson.Parse(data)
	usageNode := parsed.Get("usageMetadata")
	if !usageNode.Exists() {
		return
	}
	if v := usageNode.Get("promptTokenCount").Int(); v > 0 {
		usage.InputTokens = int(v)
	}
	if v := usageNode.Get("candidatesTokenCount").Int(); v > 0 {
		usage.OutputTokens = int(v)
	}
	if cached := usageNode.Get("cachedContentTokenCount").Int(); cached > 0 {
		usage.CacheReadInputTokens = int(cached)
	}
}

// parsePassthroughUsageFromBody extracts usage from a complete (non-streaming) response body.
func parsePassthroughUsageFromBody(protocol passthroughProtocol, body []byte) *ClaudeUsage {
	usage := &ClaudeUsage{}
	if len(body) == 0 {
		return usage
	}

	parsed := gjson.ParseBytes(body)

	switch protocol {
	case protocolAnthropic:
		usageNode := parsed.Get("usage")
		if !usageNode.Exists() {
			return usage
		}
		usage.InputTokens = int(usageNode.Get("input_tokens").Int())
		usage.OutputTokens = int(usageNode.Get("output_tokens").Int())
		usage.CacheCreationInputTokens = int(usageNode.Get("cache_creation_input_tokens").Int())
		usage.CacheReadInputTokens = int(usageNode.Get("cache_read_input_tokens").Int())

	case protocolOpenAI:
		usageNode := parsed.Get("usage")
		if !usageNode.Exists() {
			return usage
		}
		usage.InputTokens = int(usageNode.Get("prompt_tokens").Int())
		usage.OutputTokens = int(usageNode.Get("completion_tokens").Int())
		if cached := usageNode.Get("prompt_tokens_details.cached_tokens").Int(); cached > 0 {
			usage.CacheReadInputTokens = int(cached)
		}

	case protocolGemini:
		usageNode := parsed.Get("usageMetadata")
		if !usageNode.Exists() {
			return usage
		}
		usage.InputTokens = int(usageNode.Get("promptTokenCount").Int())
		usage.OutputTokens = int(usageNode.Get("candidatesTokenCount").Int())
		if cached := usageNode.Get("cachedContentTokenCount").Int(); cached > 0 {
			usage.CacheReadInputTokens = int(cached)
		}
	}

	return usage
}

// geminiStreamEventIsTerminal detects the terminal event in a Gemini SSE stream.
// Gemini signals completion by including "usageMetadata" in the final chunk.
func geminiStreamEventIsTerminal(data string) bool {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return false
	}
	return gjson.Get(trimmed, "usageMetadata").Exists()
}

// writePassthroughResponseHeaders copies response headers from upstream to client.
func writePassthroughResponseHeaders(dst http.Header, src http.Header, filter *responseheaders.CompiledHeaderFilter) {
	if dst == nil || src == nil {
		return
	}
	if filter != nil {
		responseheaders.WriteFilteredHeaders(dst, src, filter)
		return
	}
	if v := strings.TrimSpace(src.Get("Content-Type")); v != "" {
		dst.Set("Content-Type", v)
	}
	if v := strings.TrimSpace(src.Get("x-request-id")); v != "" {
		dst.Set("x-request-id", v)
	}
}
