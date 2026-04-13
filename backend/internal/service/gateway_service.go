package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
	"github.com/cespare/xxhash/v2"
	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	claudeAPIURL            = "https://api.anthropic.com/v1/messages?beta=true"
	claudeAPICountTokensURL = "https://api.anthropic.com/v1/messages/count_tokens?beta=true"
	stickySessionTTL        = time.Hour // 粘性会话TTL
	defaultMaxLineSize      = 500 * 1024 * 1024
	defaultModelsListCacheTTL = 15 * time.Second
	postUsageBillingTimeout      = 15 * time.Second
)

// ForceCacheBillingContextKey 强制缓存计费上下文键
// 用于粘性会话切换时，将 input_tokens 转为 cache_read_input_tokens 计费
type forceCacheBillingKeyType struct{}

var ForceCacheBillingContextKey = forceCacheBillingKeyType{}

var (
	windowCostPrefetchCacheHitTotal  atomic.Int64
	windowCostPrefetchCacheMissTotal atomic.Int64
	windowCostPrefetchBatchSQLTotal  atomic.Int64
	windowCostPrefetchFallbackTotal  atomic.Int64
	windowCostPrefetchErrorTotal     atomic.Int64

	modelsListCacheHitTotal   atomic.Int64
	modelsListCacheMissTotal  atomic.Int64
	modelsListCacheStoreTotal atomic.Int64
)

func GatewayWindowCostPrefetchStats() (cacheHit, cacheMiss, batchSQL, fallback, errCount int64) {
	return windowCostPrefetchCacheHitTotal.Load(),
		windowCostPrefetchCacheMissTotal.Load(),
		windowCostPrefetchBatchSQLTotal.Load(),
		windowCostPrefetchFallbackTotal.Load(),
		windowCostPrefetchErrorTotal.Load()
}

func GatewayModelsListCacheStats() (cacheHit, cacheMiss, store int64) {
	return modelsListCacheHitTotal.Load(), modelsListCacheMissTotal.Load(), modelsListCacheStoreTotal.Load()
}

func openAIStreamEventIsTerminal(data string) bool {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return false
	}
	if trimmed == "[DONE]" {
		return true
	}
	switch gjson.Get(trimmed, "type").String() {
	case "response.completed", "response.done", "response.failed":
		return true
	default:
		return false
	}
}

func anthropicStreamEventIsTerminal(eventName, data string) bool {
	if strings.EqualFold(strings.TrimSpace(eventName), "message_stop") {
		return true
	}
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return false
	}
	if trimmed == "[DONE]" {
		return true
	}
	return gjson.Get(trimmed, "type").String() == "message_stop"
}

func cloneStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

// IsForceCacheBilling 检查是否启用强制缓存计费
func IsForceCacheBilling(ctx context.Context) bool {
	v, _ := ctx.Value(ForceCacheBillingContextKey).(bool)
	return v
}

// WithForceCacheBilling 返回带有强制缓存计费标记的上下文
func WithForceCacheBilling(ctx context.Context) context.Context {
	return context.WithValue(ctx, ForceCacheBillingContextKey, true)
}

func (s *GatewayService) debugModelRoutingEnabled() bool {
	if s == nil {
		return false
	}
	return s.debugModelRouting.Load()
}


func parseDebugEnvBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func shortSessionHash(sessionHash string) string {
	if sessionHash == "" {
		return ""
	}
	if len(sessionHash) <= 8 {
		return sessionHash
	}
	return sessionHash[:8]
}

func redactAuthHeaderValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	// Keep scheme for debugging, redact secret.
	if strings.HasPrefix(strings.ToLower(v), "bearer ") {
		return "Bearer [redacted]"
	}
	return "[redacted]"
}

func safeHeaderValueForLog(key string, v string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	switch key {
	case "authorization", "x-api-key":
		return redactAuthHeaderValue(v)
	default:
		return strings.TrimSpace(v)
	}
}

func extractSystemPreviewFromBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	sys := gjson.GetBytes(body, "system")
	if !sys.Exists() {
		return ""
	}

	switch {
	case sys.IsArray():
		for _, item := range sys.Array() {
			if !item.IsObject() {
				continue
			}
			if strings.EqualFold(item.Get("type").String(), "text") {
				if t := item.Get("text").String(); strings.TrimSpace(t) != "" {
					return t
				}
			}
		}
		return ""
	case sys.Type == gjson.String:
		return sys.String()
	default:
		return ""
	}
}


// derefGroupID safely dereferences *int64 to int64, returning 0 if nil
func derefGroupID(groupID *int64) int64 {
	if groupID == nil {
		return 0
	}
	return *groupID
}

func resolveModelsListCacheTTL(cfg *config.Config) time.Duration {
	if cfg == nil || cfg.Gateway.ModelsListCacheTTLSeconds <= 0 {
		return defaultModelsListCacheTTL
	}
	return time.Duration(cfg.Gateway.ModelsListCacheTTLSeconds) * time.Second
}

func modelsListCacheKey(groupID *int64, platform string) string {
	return fmt.Sprintf("%d|%s", derefGroupID(groupID), strings.TrimSpace(platform))
}

func prefetchedStickyGroupIDFromContext(ctx context.Context) (int64, bool) {
	return PrefetchedStickyGroupIDFromContext(ctx)
}

func prefetchedStickyAccountIDFromContext(ctx context.Context, groupID *int64) int64 {
	prefetchedGroupID, ok := prefetchedStickyGroupIDFromContext(ctx)
	if !ok || prefetchedGroupID != derefGroupID(groupID) {
		return 0
	}
	if accountID, ok := PrefetchedStickyAccountIDFromContext(ctx); ok && accountID > 0 {
		return accountID
	}
	return 0
}

// shouldClearStickySession 检查账号是否处于不可调度状态，需要清理粘性会话绑定。
// 当账号状态为错误、禁用、不可调度、处于临时不可调度期间，
// 或请求的模型处于限流状态时，返回 true。
// 这确保后续请求不会继续使用不可用的账号。
//
// shouldClearStickySession checks if an account is in an unschedulable state
// and the sticky session binding should be cleared.
// Returns true when account status is error/disabled, schedulable is false,
// within temporary unschedulable period, or the requested model is rate-limited.
// This ensures subsequent requests won't continue using unavailable accounts.
func shouldClearStickySession(account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	if account.Status == StatusError || account.Status == StatusDisabled || !account.Schedulable {
		return true
	}
	if account.TempUnschedulableUntil != nil && time.Now().Before(*account.TempUnschedulableUntil) {
		return true
	}
	// 检查模型限流和 scope 限流，有限流即清除粘性会话
	if remaining := account.GetRateLimitRemainingTimeWithContext(context.Background(), requestedModel); remaining > 0 {
		return true
	}
	return false
}

type AccountSelectionResult struct {
	Account     *Account
	Acquired    bool
	ReleaseFunc func()
}

// ClaudeUsage 表示Claude API返回的usage信息
type ClaudeUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreation5mTokens    int // 5分钟缓存创建token（来自嵌套 cache_creation 对象）
	CacheCreation1hTokens    int // 1小时缓存创建token（来自嵌套 cache_creation 对象）
}

// ForwardResult 转发结果
type ForwardResult struct {
	RequestID string
	Usage     ClaudeUsage
	Model     string
	// UpstreamModel is the actual upstream model after mapping.
	// Prefer empty when it is identical to Model; persistence normalizes equal values away as no-op mappings.
	UpstreamModel    string
	Stream           bool
	Duration         time.Duration
	FirstTokenMs     *int // 首字时间（流式请求）
	ClientDisconnect bool // 客户端是否在流式传输过程中断开
	ReasoningEffort  *string

	// 图片生成计费字段（图片生成模型使用）
	ImageCount int    // 生成的图片数量
	ImageSize  string // 图片尺寸 "1K", "2K", "4K"

	// 媒体字段（legacy，透传模式下不使用）
	MediaType string // image / video / prompt
	MediaURL  string // 生成后的媒体地址（可选）
}

// UpstreamFailoverError indicates an upstream error that should trigger account failover.
type UpstreamFailoverError struct {
	StatusCode             int
	ResponseBody           []byte      // 上游响应体，用于错误透传规则匹配
	ResponseHeaders        http.Header // 上游响应头，用于透传 cf-ray/cf-mitigated/content-type 等诊断信息
	ForceCacheBilling      bool        // 粘性会话切换时设为 true
	RetryableOnSameAccount bool        // 临时性错误（如 Google 间歇性 400、空响应），应在同一账号上重试 N 次再切换
}

func (e *UpstreamFailoverError) Error() string {
	return fmt.Sprintf("upstream error: %d (failover)", e.StatusCode)
}

// TempUnscheduleRetryableError 对 RetryableOnSameAccount 类型的 failover 错误触发临时封禁。
// 由 handler 层在同账号重试全部用尽、切换账号时调用。
func (s *GatewayService) TempUnscheduleRetryableError(ctx context.Context, accountID int64, failoverErr *UpstreamFailoverError) {
	if failoverErr == nil || !failoverErr.RetryableOnSameAccount {
		return
	}
	// 根据状态码选择封禁策略
	switch failoverErr.StatusCode {
	case http.StatusBadRequest:
		tempUnscheduleGoogleConfigError(ctx, s.accountRepo, accountID, "[handler]")
	case http.StatusBadGateway:
		tempUnscheduleEmptyResponse(ctx, s.accountRepo, accountID, "[handler]")
	}
}

// GatewayService handles API gateway operations
type GatewayService struct {
	accountRepo           AccountRepository
	groupRepo             GroupRepository
	usageLogRepo          UsageLogRepository
	usageBillingRepo      UsageBillingRepository
	userRepo              UserRepository
	userSubRepo           UserSubscriptionRepository
	cache GatewayCache
	digestStore           *DigestSessionStore
	cfg                   *config.Config
	schedulerSnapshot     *SchedulerSnapshotService
	billingService        *BillingService
	rateLimitService      *RateLimitService
	billingCacheService   *BillingCacheService
	identityService       *IdentityService
	httpUpstream          HTTPUpstream
	deferredService       *DeferredService
	claudeTokenProvider   *ClaudeTokenProvider
	sessionLimitCache     SessionLimitCache // 会话数量限制缓存（仅 Anthropic OAuth/SetupToken）
	rpmCache              RPMCache          // RPM 计数缓存（仅 Anthropic OAuth/SetupToken）
	modelsListCache *gocache.Cache
	modelsListCacheTTL    time.Duration
	settingService        *SettingService
	responseHeaderFilter  *responseheaders.CompiledHeaderFilter
	debugModelRouting     atomic.Bool
	tlsFPProfileService   *TLSFingerprintProfileService
}

// NewGatewayService creates a new GatewayService
func NewGatewayService(
	accountRepo AccountRepository,
	groupRepo GroupRepository,
	usageLogRepo UsageLogRepository,
	usageBillingRepo UsageBillingRepository,
	userRepo UserRepository,
	userSubRepo UserSubscriptionRepository,
	cache GatewayCache,
	cfg *config.Config,
	schedulerSnapshot *SchedulerSnapshotService,
	billingService *BillingService,
	rateLimitService *RateLimitService,
	billingCacheService *BillingCacheService,
	identityService *IdentityService,
	httpUpstream HTTPUpstream,
	deferredService *DeferredService,
	claudeTokenProvider *ClaudeTokenProvider,
	sessionLimitCache SessionLimitCache,
	rpmCache RPMCache,
	digestStore *DigestSessionStore,
	settingService *SettingService,
	tlsFPProfileService *TLSFingerprintProfileService,
) *GatewayService {
	modelsListTTL := resolveModelsListCacheTTL(cfg)

	svc := &GatewayService{
		accountRepo:          accountRepo,
		groupRepo:            groupRepo,
		usageLogRepo:         usageLogRepo,
		usageBillingRepo:     usageBillingRepo,
		userRepo:             userRepo,
		userSubRepo: userSubRepo,
		cache:       cache,
		digestStore:          digestStore,
		cfg:                  cfg,
		schedulerSnapshot:    schedulerSnapshot,
		billingService:       billingService,
		rateLimitService:     rateLimitService,
		billingCacheService:  billingCacheService,
		identityService:      identityService,
		httpUpstream:         httpUpstream,
		deferredService:      deferredService,
		claudeTokenProvider:  claudeTokenProvider,
		sessionLimitCache:    sessionLimitCache,
		rpmCache:       rpmCache,
		settingService: settingService,
		modelsListCache:      gocache.New(modelsListTTL, time.Minute),
		modelsListCacheTTL:   modelsListTTL,
		responseHeaderFilter: compileResponseHeaderFilter(cfg),
		tlsFPProfileService:  tlsFPProfileService,
	}
	svc.debugModelRouting.Store(parseDebugEnvBool(os.Getenv("SUB2API_DEBUG_MODEL_ROUTING")))
	return svc
}

// GenerateSessionHash 从预解析请求计算粘性会话 hash
func (s *GatewayService) GenerateSessionHash(parsed *ParsedRequest) string {
	if parsed == nil {
		return ""
	}

	// 1. 最高优先级：从 metadata.user_id 提取 session_xxx
	if parsed.MetadataUserID != "" {
		if uid := ParseMetadataUserID(parsed.MetadataUserID); uid != nil && uid.SessionID != "" {
			return uid.SessionID
		}
	}

	// 2. 提取带 cache_control: {type: "ephemeral"} 的内容
	cacheableContent := s.extractCacheableContent(parsed)
	if cacheableContent != "" {
		return s.hashContent(cacheableContent)
	}

	// 3. 最后 fallback: 使用 session上下文 + system + 所有消息的完整摘要串
	var combined strings.Builder
	// 混入请求上下文区分因子，避免不同用户相同消息产生相同 hash
	if parsed.SessionContext != nil {
		_, _ = combined.WriteString(parsed.SessionContext.ClientIP)
		_, _ = combined.WriteString(":")
		_, _ = combined.WriteString(NormalizeSessionUserAgent(parsed.SessionContext.UserAgent))
		_, _ = combined.WriteString(":")
		_, _ = combined.WriteString(strconv.FormatInt(parsed.SessionContext.APIKeyID, 10))
		_, _ = combined.WriteString("|")
	}
	if parsed.System != nil {
		systemText := s.extractTextFromSystem(parsed.System)
		if systemText != "" {
			_, _ = combined.WriteString(systemText)
		}
	}
	for _, msg := range parsed.Messages {
		if m, ok := msg.(map[string]any); ok {
			if content, exists := m["content"]; exists {
				// Anthropic: messages[].content
				if msgText := s.extractTextFromContent(content); msgText != "" {
					_, _ = combined.WriteString(msgText)
				}
			} else if parts, ok := m["parts"].([]any); ok {
				// Gemini: contents[].parts[].text
				for _, part := range parts {
					if partMap, ok := part.(map[string]any); ok {
						if text, ok := partMap["text"].(string); ok {
							_, _ = combined.WriteString(text)
						}
					}
				}
			}
		}
	}
	if combined.Len() > 0 {
		return s.hashContent(combined.String())
	}

	return ""
}

// BindStickySession sets session -> account binding with standard TTL.
func (s *GatewayService) BindStickySession(ctx context.Context, groupID *int64, sessionHash string, accountID int64) error {
	if sessionHash == "" || accountID <= 0 || s.cache == nil {
		return nil
	}
	return s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), sessionHash, accountID, stickySessionTTL)
}

// GetCachedSessionAccountID retrieves the account ID bound to a sticky session.
// Returns 0 if no binding exists or on error.
func (s *GatewayService) GetCachedSessionAccountID(ctx context.Context, groupID *int64, sessionHash string) (int64, error) {
	if sessionHash == "" || s.cache == nil {
		return 0, nil
	}
	accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
	if err != nil {
		return 0, err
	}
	return accountID, nil
}

// FindGeminiSession 查找 Gemini 会话（基于内容摘要链的 Fallback 匹配）
// 返回最长匹配的会话信息（uuid, accountID）
func (s *GatewayService) FindGeminiSession(_ context.Context, groupID int64, prefixHash, digestChain string) (uuid string, accountID int64, matchedChain string, found bool) {
	if digestChain == "" || s.digestStore == nil {
		return "", 0, "", false
	}
	return s.digestStore.Find(groupID, prefixHash, digestChain)
}

// SaveGeminiSession 保存 Gemini 会话。oldDigestChain 为 Find 返回的 matchedChain，用于删旧 key。
func (s *GatewayService) SaveGeminiSession(_ context.Context, groupID int64, prefixHash, digestChain, uuid string, accountID int64, oldDigestChain string) error {
	if digestChain == "" || s.digestStore == nil {
		return nil
	}
	s.digestStore.Save(groupID, prefixHash, digestChain, uuid, accountID, oldDigestChain)
	return nil
}

// FindAnthropicSession 查找 Anthropic 会话（基于内容摘要链的 Fallback 匹配）
func (s *GatewayService) FindAnthropicSession(_ context.Context, groupID int64, prefixHash, digestChain string) (uuid string, accountID int64, matchedChain string, found bool) {
	if digestChain == "" || s.digestStore == nil {
		return "", 0, "", false
	}
	return s.digestStore.Find(groupID, prefixHash, digestChain)
}

// SaveAnthropicSession 保存 Anthropic 会话
func (s *GatewayService) SaveAnthropicSession(_ context.Context, groupID int64, prefixHash, digestChain, uuid string, accountID int64, oldDigestChain string) error {
	if digestChain == "" || s.digestStore == nil {
		return nil
	}
	s.digestStore.Save(groupID, prefixHash, digestChain, uuid, accountID, oldDigestChain)
	return nil
}

func (s *GatewayService) extractCacheableContent(parsed *ParsedRequest) string {
	if parsed == nil {
		return ""
	}

	var builder strings.Builder

	// 检查 system 中的 cacheable 内容
	if system, ok := parsed.System.([]any); ok {
		for _, part := range system {
			if partMap, ok := part.(map[string]any); ok {
				if cc, ok := partMap["cache_control"].(map[string]any); ok {
					if cc["type"] == "ephemeral" {
						if text, ok := partMap["text"].(string); ok {
							_, _ = builder.WriteString(text)
						}
					}
				}
			}
		}
	}
	systemText := builder.String()

	// 检查 messages 中的 cacheable 内容
	for _, msg := range parsed.Messages {
		if msgMap, ok := msg.(map[string]any); ok {
			if msgContent, ok := msgMap["content"].([]any); ok {
				for _, part := range msgContent {
					if partMap, ok := part.(map[string]any); ok {
						if cc, ok := partMap["cache_control"].(map[string]any); ok {
							if cc["type"] == "ephemeral" {
								return s.extractTextFromContent(msgMap["content"])
							}
						}
					}
				}
			}
		}
	}

	return systemText
}

func (s *GatewayService) extractTextFromSystem(system any) string {
	switch v := system.(type) {
	case string:
		return v
	case []any:
		var texts []string
		for _, part := range v {
			if partMap, ok := part.(map[string]any); ok {
				if text, ok := partMap["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
		return strings.Join(texts, "")
	}
	return ""
}

func (s *GatewayService) extractTextFromContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var texts []string
		for _, part := range v {
			if partMap, ok := part.(map[string]any); ok {
				if partMap["type"] == "text" {
					if text, ok := partMap["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		return strings.Join(texts, "")
	}
	return ""
}

func (s *GatewayService) hashContent(content string) string {
	h := xxhash.Sum64String(content)
	return strconv.FormatUint(h, 36)
}

type anthropicCacheControlPayload struct {
	Type string `json:"type"`
}

type anthropicSystemTextBlockPayload struct {
	Type         string                        `json:"type"`
	Text         string                        `json:"text"`
	CacheControl *anthropicCacheControlPayload `json:"cache_control,omitempty"`
}

type anthropicMetadataPayload struct {
	UserID string `json:"user_id"`
}

// replaceModelInBody 替换请求体中的model字段
// 优先使用定点修改，尽量保持客户端原始字段顺序。
func (s *GatewayService) replaceModelInBody(body []byte, newModel string) []byte {
	if len(body) == 0 {
		return body
	}
	if current := gjson.GetBytes(body, "model"); current.Exists() && current.String() == newModel {
		return body
	}
	newBody, err := sjson.SetBytes(body, "model", newModel)
	if err != nil {
		return body
	}
	return newBody
}

func GenerateSessionUUID(seed string) string {
	return generateSessionUUID(seed)
}

func generateSessionUUID(seed string) string {
	if seed == "" {
		return uuid.NewString()
	}
	hash := sha256.Sum256([]byte(seed))
	bytes := hash[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

// SelectAccount 选择账号（粘性会话+优先级）
func (s *GatewayService) SelectAccount(ctx context.Context, groupID *int64, sessionHash string) (*Account, error) {
	return s.SelectAccountForModel(ctx, groupID, sessionHash, "")
}

// SelectAccountForModel 选择支持指定模型的账号（粘性会话+优先级+模型映射）
func (s *GatewayService) SelectAccountForModel(ctx context.Context, groupID *int64, sessionHash string, requestedModel string) (*Account, error) {
	return s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, nil)
}

// SelectAccountForModelWithExclusions selects an account supporting the requested model while excluding specified accounts.
func (s *GatewayService) SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error) {
	var platform string
	forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
	if hasForcePlatform && forcePlatform != "" {
		platform = forcePlatform
	} else if groupID != nil {
		group, resolvedGroupID, err := s.resolveGatewayGroup(ctx, groupID)
		if err != nil {
			return nil, err
		}
		groupID = resolvedGroupID
		ctx = s.withGroupContext(ctx, group)
		platform = PlatformAnthropic
	} else {
		platform = PlatformAnthropic
	}

	return s.selectAccountForModelWithPlatform(ctx, groupID, sessionHash, requestedModel, excludedIDs, platform)
}

// SelectAccountWithLoadAwareness selects account with load-awareness and wait plan.
// metadataUserID: 已废弃参数，会话限制现在统一使用 sessionHash
func (s *GatewayService) SelectAccountWithLoadAwareness(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, metadataUserID string) (*AccountSelectionResult, error) {
	// 调试日志：记录调度入口参数
	excludedIDsList := make([]int64, 0, len(excludedIDs))
	for id := range excludedIDs {
		excludedIDsList = append(excludedIDsList, id)
	}
	slog.Debug("account_scheduling_starting",
		"group_id", derefGroupID(groupID),
		"model", requestedModel,
		"session", shortSessionHash(sessionHash),
		"excluded_ids", excludedIDsList)

	// 检查 Claude Code 客户端限制（可能会替换 groupID 为降级分组）
	group, groupID, err := s.checkClaudeCodeRestriction(ctx, groupID)
	if err != nil {
		return nil, err
	}
	ctx = s.withGroupContext(ctx, group)

	var stickyAccountID int64
	if prefetch := prefetchedStickyAccountIDFromContext(ctx, groupID); prefetch > 0 {
		stickyAccountID = prefetch
	} else if sessionHash != "" && s.cache != nil {
		if accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash); err == nil {
			stickyAccountID = accountID
		}
	}

	if s.debugModelRoutingEnabled() && requestedModel != "" {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] select entry: group_id=%v model=%s session=%s sticky_account=%d",
			derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), stickyAccountID)
	}

	// 复制排除列表，用于会话限制拒绝时的重试
	localExcluded := make(map[int64]struct{})
	for k, v := range excludedIDs {
		localExcluded[k] = v
	}

	for {
		account, err := s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, localExcluded)
		if err != nil {
			return nil, err
		}

		// 检查会话限制（使用 sessionHash 作为会话标识符）
		if !s.checkAndRegisterSession(ctx, account, sessionHash) {
			localExcluded[account.ID] = struct{}{} // 排除此账号
			continue                               // 重新选择
		}
		return &AccountSelectionResult{
			Account:     account,
			Acquired:    true,
			ReleaseFunc: func() {},
		}, nil
	}
}


func (s *GatewayService) schedulingConfig() config.GatewaySchedulingConfig {
	if s.cfg != nil {
		return s.cfg.Gateway.Scheduling
	}
	return config.GatewaySchedulingConfig{
		StickySessionMaxWaiting:  3,
		StickySessionWaitTimeout: 45 * time.Second,
		FallbackWaitTimeout:      30 * time.Second,
		FallbackMaxWaiting:       100,
		LoadBatchEnabled:         true,
		SlotCleanupInterval:      30 * time.Second,
	}
}

func (s *GatewayService) withGroupContext(ctx context.Context, group *Group) context.Context {
	if !IsGroupContextValid(group) {
		return ctx
	}
	if existing, ok := ctx.Value(ctxkey.Group).(*Group); ok && existing != nil && existing.ID == group.ID && IsGroupContextValid(existing) {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.Group, group)
}

func (s *GatewayService) groupFromContext(ctx context.Context, groupID int64) *Group {
	if group, ok := ctx.Value(ctxkey.Group).(*Group); ok && IsGroupContextValid(group) && group.ID == groupID {
		return group
	}
	return nil
}

func (s *GatewayService) resolveGroupByID(ctx context.Context, groupID int64) (*Group, error) {
	if group := s.groupFromContext(ctx, groupID); group != nil {
		return group, nil
	}
	group, err := s.groupRepo.GetByIDLite(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group failed: %w", err)
	}
	return group, nil
}

func (s *GatewayService) ResolveGroupByID(ctx context.Context, groupID int64) (*Group, error) {
	return s.resolveGroupByID(ctx, groupID)
}

func (s *GatewayService) routingAccountIDsForRequest(ctx context.Context, groupID *int64, requestedModel string, platform string) []int64 {
	if groupID == nil || requestedModel == "" {
		return nil
	}
	group, err := s.resolveGroupByID(ctx, *groupID)
	if err != nil || group == nil {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] resolve group failed: group_id=%v model=%s platform=%s err=%v", derefGroupID(groupID), requestedModel, platform, err)
		}
		return nil
	}
	ids := group.GetRoutingAccountIDs(requestedModel)
	if s.debugModelRoutingEnabled() {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routing lookup: group_id=%d model=%s enabled=%v rules=%d matched_ids=%v",
			group.ID, requestedModel, group.ModelRoutingEnabled, len(group.ModelRouting), ids)
	}
	return ids
}

func (s *GatewayService) resolveGatewayGroup(ctx context.Context, groupID *int64) (*Group, *int64, error) {
	if groupID == nil {
		return nil, nil, nil
	}

	currentID := *groupID
	visited := map[int64]struct{}{}
	for {
		if _, seen := visited[currentID]; seen {
			return nil, nil, fmt.Errorf("fallback group cycle detected")
		}
		visited[currentID] = struct{}{}

		group, err := s.resolveGroupByID(ctx, currentID)
		if err != nil {
			return nil, nil, err
		}

		if !group.ClaudeCodeOnly || IsClaudeCodeClient(ctx) {
			return group, &currentID, nil
		}

		if group.FallbackGroupID == nil {
			return nil, nil, ErrClaudeCodeOnly
		}
		currentID = *group.FallbackGroupID
	}
}

// checkClaudeCodeRestriction 检查分组的 Claude Code 客户端限制
// 如果分组启用了 claude_code_only 且请求不是来自 Claude Code 客户端：
//   - 有降级分组：返回降级分组的 ID
//   - 无降级分组：返回 ErrClaudeCodeOnly 错误
func (s *GatewayService) checkClaudeCodeRestriction(ctx context.Context, groupID *int64) (*Group, *int64, error) {
	if groupID == nil {
		return nil, groupID, nil
	}

	// 强制平台模式不检查 Claude Code 限制
	if forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string); hasForcePlatform && forcePlatform != "" {
		return nil, groupID, nil
	}

	group, resolvedID, err := s.resolveGatewayGroup(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}

	return group, resolvedID, nil
}

func (s *GatewayService) resolvePlatform(ctx context.Context) (string, bool, error) {
	forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
	if hasForcePlatform && forcePlatform != "" {
		return forcePlatform, true, nil
	}
	return PlatformAnthropic, false, nil
}

func (s *GatewayService) listSchedulableAccounts(ctx context.Context, groupID *int64, platform string, hasForcePlatform bool) ([]Account, bool, error) {
	if s.schedulerSnapshot != nil {
		accounts, useMixed, err := s.schedulerSnapshot.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		if err == nil {
			slog.Debug("account_scheduling_list_snapshot",
				"group_id", derefGroupID(groupID),
				"platform", platform,
				"use_mixed", useMixed,
				"count", len(accounts))
			for _, acc := range accounts {
				slog.Debug("account_scheduling_account_detail",
					"account_id", acc.ID,
					"name", acc.Name,
					"platform", acc.Platform,
					"type", acc.Type,
					"status", acc.Status,
					"tls_fingerprint", acc.IsTLSFingerprintEnabled())
			}
		}
		return accounts, useMixed, err
	}
	const useMixed = false

	var accounts []Account
	var err error
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		accounts, err = s.accountRepo.ListSchedulable(ctx)
	} else if groupID != nil {
		// 透传模式：按分组查询全部可调度账号，不按 platform 过滤。
		// 同一个上游 Key 支持多种协议，ForwardPassthrough 根据请求路径自动选择认证方式。
		accounts, err = s.accountRepo.ListSchedulableByGroupID(ctx, *groupID)
	} else {
		accounts, err = s.accountRepo.ListSchedulableUngrouped(ctx)
	}
	if err != nil {
		slog.Debug("account_scheduling_list_failed",
			"group_id", derefGroupID(groupID),
			"platform", platform,
			"error", err)
		return nil, useMixed, err
	}
	slog.Debug("account_scheduling_list",
		"group_id", derefGroupID(groupID),
		"platform", platform,
		"count", len(accounts))
	for _, acc := range accounts {
		slog.Debug("account_scheduling_account_detail",
			"account_id", acc.ID,
			"name", acc.Name,
			"platform", acc.Platform,
			"type", acc.Type,
			"status", acc.Status,
			"tls_fingerprint", acc.IsTLSFingerprintEnabled())
	}
	return accounts, useMixed, nil
}

func (s *GatewayService) isAccountAllowedForPlatform(account *Account, _ string, _ bool) bool {
	// 透传模式：不按 platform 过滤，同一个上游 Key 支持多种协议。
	return account != nil
}

func (s *GatewayService) isAccountSchedulableForSelection(account *Account) bool {
	if account == nil {
		return false
	}
	return account.IsSchedulable()
}

func (s *GatewayService) isAccountSchedulableForModelSelection(ctx context.Context, account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	return account.IsSchedulableForModelWithContext(ctx, requestedModel)
}

// isAccountInGroup checks if the account belongs to the specified group.
// When groupID is nil, returns true only for ungrouped accounts (no group assignments).
func (s *GatewayService) isAccountInGroup(account *Account, groupID *int64) bool {
	if account == nil {
		return false
	}
	if groupID == nil {
		// 无分组的 API Key 只能使用未分组的账号
		return len(account.AccountGroups) == 0
	}
	for _, ag := range account.AccountGroups {
		if ag.GroupID == *groupID {
			return true
		}
	}
	return false
}



type usageLogWindowStatsBatchProvider interface {
	GetAccountWindowStatsBatch(ctx context.Context, accountIDs []int64, startTime time.Time) (map[int64]*usagestats.AccountStats, error)
}

type windowCostPrefetchContextKeyType struct{}

var windowCostPrefetchContextKey = windowCostPrefetchContextKeyType{}

func windowCostFromPrefetchContext(ctx context.Context, accountID int64) (float64, bool) {
	if ctx == nil || accountID <= 0 {
		return 0, false
	}
	m, ok := ctx.Value(windowCostPrefetchContextKey).(map[int64]float64)
	if !ok || len(m) == 0 {
		return 0, false
	}
	v, exists := m[accountID]
	return v, exists
}

func (s *GatewayService) withWindowCostPrefetch(ctx context.Context, accounts []Account) context.Context {
	if ctx == nil || len(accounts) == 0 || s.sessionLimitCache == nil || s.usageLogRepo == nil {
		return ctx
	}

	accountByID := make(map[int64]*Account)
	accountIDs := make([]int64, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if account == nil || !account.IsAnthropicOAuthOrSetupToken() {
			continue
		}
		if account.GetWindowCostLimit() <= 0 {
			continue
		}
		accountByID[account.ID] = account
		accountIDs = append(accountIDs, account.ID)
	}
	if len(accountIDs) == 0 {
		return ctx
	}

	costs := make(map[int64]float64, len(accountIDs))
	cacheValues, err := s.sessionLimitCache.GetWindowCostBatch(ctx, accountIDs)
	if err == nil {
		for accountID, cost := range cacheValues {
			costs[accountID] = cost
		}
		windowCostPrefetchCacheHitTotal.Add(int64(len(cacheValues)))
	} else {
		windowCostPrefetchErrorTotal.Add(1)
		logger.LegacyPrintf("service.gateway", "window_cost batch cache read failed: %v", err)
	}
	cacheMissCount := len(accountIDs) - len(costs)
	if cacheMissCount < 0 {
		cacheMissCount = 0
	}
	windowCostPrefetchCacheMissTotal.Add(int64(cacheMissCount))

	missingByStart := make(map[int64][]int64)
	startTimes := make(map[int64]time.Time)
	for _, accountID := range accountIDs {
		if _, ok := costs[accountID]; ok {
			continue
		}
		account := accountByID[accountID]
		if account == nil {
			continue
		}
		startTime := account.GetCurrentWindowStartTime()
		startKey := startTime.Unix()
		missingByStart[startKey] = append(missingByStart[startKey], accountID)
		startTimes[startKey] = startTime
	}
	if len(missingByStart) == 0 {
		return context.WithValue(ctx, windowCostPrefetchContextKey, costs)
	}

	batchReader, hasBatch := s.usageLogRepo.(usageLogWindowStatsBatchProvider)
	for startKey, ids := range missingByStart {
		startTime := startTimes[startKey]

		if hasBatch {
			windowCostPrefetchBatchSQLTotal.Add(1)
			queryStart := time.Now()
			statsByAccount, err := batchReader.GetAccountWindowStatsBatch(ctx, ids, startTime)
			if err == nil {
				slog.Debug("window_cost_batch_query_ok",
					"accounts", len(ids),
					"window_start", startTime.Format(time.RFC3339),
					"duration_ms", time.Since(queryStart).Milliseconds())
				for _, accountID := range ids {
					stats := statsByAccount[accountID]
					cost := 0.0
					if stats != nil {
						cost = stats.StandardCost
					}
					costs[accountID] = cost
					_ = s.sessionLimitCache.SetWindowCost(ctx, accountID, cost)
				}
				continue
			}
			windowCostPrefetchErrorTotal.Add(1)
			logger.LegacyPrintf("service.gateway", "window_cost batch db query failed: start=%s err=%v", startTime.Format(time.RFC3339), err)
		}

		// 回退路径：缺少批量仓储能力或批量查询失败时，按账号单查（失败开放）。
		windowCostPrefetchFallbackTotal.Add(int64(len(ids)))
		for _, accountID := range ids {
			stats, err := s.usageLogRepo.GetAccountWindowStats(ctx, accountID, startTime)
			if err != nil {
				windowCostPrefetchErrorTotal.Add(1)
				continue
			}
			cost := stats.StandardCost
			costs[accountID] = cost
			_ = s.sessionLimitCache.SetWindowCost(ctx, accountID, cost)
		}
	}

	return context.WithValue(ctx, windowCostPrefetchContextKey, costs)
}

// isAccountSchedulableForQuota 检查账号是否在配额限制内
// 适用于配置了 quota_limit 的 apikey 类型账号
func (s *GatewayService) isAccountSchedulableForQuota(account *Account) bool {
	if account.Type != AccountTypeAPIKey {
		return true
	}
	return !account.IsQuotaExceeded()
}

// isAccountSchedulableForWindowCost 检查账号是否可根据窗口费用进行调度
// 仅适用于 Anthropic OAuth/SetupToken 账号
// 返回 true 表示可调度，false 表示不可调度
func (s *GatewayService) isAccountSchedulableForWindowCost(ctx context.Context, account *Account, isSticky bool) bool {
	// 只检查 Anthropic OAuth/SetupToken 账号
	if !account.IsAnthropicOAuthOrSetupToken() {
		return true
	}

	limit := account.GetWindowCostLimit()
	if limit <= 0 {
		return true // 未启用窗口费用限制
	}

	// 尝试从缓存获取窗口费用
	var currentCost float64
	if cost, ok := windowCostFromPrefetchContext(ctx, account.ID); ok {
		currentCost = cost
		goto checkSchedulability
	}
	if s.sessionLimitCache != nil {
		if cost, hit, err := s.sessionLimitCache.GetWindowCost(ctx, account.ID); err == nil && hit {
			currentCost = cost
			goto checkSchedulability
		}
	}

	// 缓存未命中，从数据库查询
	{
		// 使用统一的窗口开始时间计算逻辑（考虑窗口过期情况）
		startTime := account.GetCurrentWindowStartTime()

		stats, err := s.usageLogRepo.GetAccountWindowStats(ctx, account.ID, startTime)
		if err != nil {
			// 失败开放：查询失败时允许调度
			return true
		}

		// 使用标准费用（不含账号倍率）
		currentCost = stats.StandardCost

		// 设置缓存（忽略错误）
		if s.sessionLimitCache != nil {
			_ = s.sessionLimitCache.SetWindowCost(ctx, account.ID, currentCost)
		}
	}

checkSchedulability:
	schedulability := account.CheckWindowCostSchedulability(currentCost)

	switch schedulability {
	case WindowCostSchedulable:
		return true
	case WindowCostStickyOnly:
		return isSticky
	case WindowCostNotSchedulable:
		return false
	}
	return true
}

// rpmPrefetchContextKey is the context key for prefetched RPM counts.
type rpmPrefetchContextKeyType struct{}

var rpmPrefetchContextKey = rpmPrefetchContextKeyType{}

func rpmFromPrefetchContext(ctx context.Context, accountID int64) (int, bool) {
	if v, ok := ctx.Value(rpmPrefetchContextKey).(map[int64]int); ok {
		count, found := v[accountID]
		return count, found
	}
	return 0, false
}

// withRPMPrefetch 批量预取所有候选账号的 RPM 计数
func (s *GatewayService) withRPMPrefetch(ctx context.Context, accounts []Account) context.Context {
	if s.rpmCache == nil {
		return ctx
	}

	var ids []int64
	for i := range accounts {
		if accounts[i].IsAnthropicOAuthOrSetupToken() && accounts[i].GetBaseRPM() > 0 {
			ids = append(ids, accounts[i].ID)
		}
	}
	if len(ids) == 0 {
		return ctx
	}

	counts, err := s.rpmCache.GetRPMBatch(ctx, ids)
	if err != nil {
		return ctx // 失败开放
	}
	return context.WithValue(ctx, rpmPrefetchContextKey, counts)
}

// isAccountSchedulableForRPM 检查账号是否可根据 RPM 进行调度
// 仅适用于 Anthropic OAuth/SetupToken 账号
func (s *GatewayService) isAccountSchedulableForRPM(ctx context.Context, account *Account, isSticky bool) bool {
	if !account.IsAnthropicOAuthOrSetupToken() {
		return true
	}
	baseRPM := account.GetBaseRPM()
	if baseRPM <= 0 {
		return true
	}

	// 尝试从预取缓存获取
	var currentRPM int
	if count, ok := rpmFromPrefetchContext(ctx, account.ID); ok {
		currentRPM = count
	} else if s.rpmCache != nil {
		if count, err := s.rpmCache.GetRPM(ctx, account.ID); err == nil {
			currentRPM = count
		}
		// 失败开放：GetRPM 错误时允许调度
	}

	schedulability := account.CheckRPMSchedulability(currentRPM)
	switch schedulability {
	case WindowCostSchedulable:
		return true
	case WindowCostStickyOnly:
		return isSticky
	case WindowCostNotSchedulable:
		return false
	}
	return true
}

// IncrementAccountRPM increments the RPM counter for the given account.
// 已知 TOCTOU 竞态：调度时读取 RPM 计数与此处递增之间存在时间窗口，
// 高并发下可能短暂超出 RPM 限制。这是与 WindowCost 一致的 soft-limit
// 设计权衡——可接受的少量超额优于加锁带来的延迟和复杂度。
func (s *GatewayService) IncrementAccountRPM(ctx context.Context, accountID int64) error {
	if s.rpmCache == nil {
		return nil
	}
	_, err := s.rpmCache.IncrementRPM(ctx, accountID)
	return err
}

// checkAndRegisterSession 检查并注册会话，用于会话数量限制
// 仅适用于 Anthropic OAuth/SetupToken 账号
// sessionID: 会话标识符（使用粘性会话的 hash）
// 返回 true 表示允许（在限制内或会话已存在），false 表示拒绝（超出限制且是新会话）
func (s *GatewayService) checkAndRegisterSession(ctx context.Context, account *Account, sessionID string) bool {
	// 只检查 Anthropic OAuth/SetupToken 账号
	if !account.IsAnthropicOAuthOrSetupToken() {
		return true
	}

	maxSessions := account.GetMaxSessions()
	if maxSessions <= 0 || sessionID == "" {
		return true // 未启用会话限制或无会话ID
	}

	if s.sessionLimitCache == nil {
		return true // 缓存不可用时允许通过
	}

	idleTimeout := time.Duration(account.GetSessionIdleTimeoutMinutes()) * time.Minute

	allowed, err := s.sessionLimitCache.RegisterSession(ctx, account.ID, sessionID, maxSessions, idleTimeout)
	if err != nil {
		// 失败开放：缓存错误时允许通过
		return true
	}
	return allowed
}

func (s *GatewayService) getSchedulableAccount(ctx context.Context, accountID int64) (*Account, error) {
	if s.schedulerSnapshot != nil {
		return s.schedulerSnapshot.GetAccount(ctx, accountID)
	}
	return s.accountRepo.GetByID(ctx, accountID)
}

func sortAccountsByPriorityAndLastUsed(accounts []*Account, preferOAuth bool) {
	sort.SliceStable(accounts, func(i, j int) bool {
		a, b := accounts[i], accounts[j]
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		switch {
		case a.LastUsedAt == nil && b.LastUsedAt != nil:
			return true
		case a.LastUsedAt != nil && b.LastUsedAt == nil:
			return false
		case a.LastUsedAt == nil && b.LastUsedAt == nil:
			if preferOAuth && a.Type != b.Type {
				return a.Type == AccountTypeOAuth
			}
			return false
		default:
			return a.LastUsedAt.Before(*b.LastUsedAt)
		}
	})
	shuffleWithinPriorityAndLastUsed(accounts, preferOAuth)
}

// shuffleWithinPriorityAndLastUsed 对排序后的 []*Account 切片，按 (Priority, LastUsedAt) 分组后组内随机打乱。
//
// 注意：当 preferOAuth=true 时，需要保证 OAuth 账号在同组内仍然优先，否则会把排序时的偏好打散掉。
// 因此这里采用"组内分区 + 分区内 shuffle"的方式：
// - 先把同组账号按 (OAuth / 非 OAuth) 拆成两段，保持 OAuth 段在前；
// - 再分别在各段内随机打散，避免热点。
func shuffleWithinPriorityAndLastUsed(accounts []*Account, preferOAuth bool) {
	if len(accounts) <= 1 {
		return
	}
	i := 0
	for i < len(accounts) {
		j := i + 1
		for j < len(accounts) && sameAccountGroup(accounts[i], accounts[j]) {
			j++
		}
		if j-i > 1 {
			if preferOAuth {
				oauth := make([]*Account, 0, j-i)
				others := make([]*Account, 0, j-i)
				for _, acc := range accounts[i:j] {
					if acc.Type == AccountTypeOAuth {
						oauth = append(oauth, acc)
					} else {
						others = append(others, acc)
					}
				}
				if len(oauth) > 1 {
					mathrand.Shuffle(len(oauth), func(a, b int) { oauth[a], oauth[b] = oauth[b], oauth[a] })
				}
				if len(others) > 1 {
					mathrand.Shuffle(len(others), func(a, b int) { others[a], others[b] = others[b], others[a] })
				}
				copy(accounts[i:], oauth)
				copy(accounts[i+len(oauth):], others)
			} else {
				mathrand.Shuffle(j-i, func(a, b int) {
					accounts[i+a], accounts[i+b] = accounts[i+b], accounts[i+a]
				})
			}
		}
		i = j
	}
}

// sameAccountGroup 判断两个 Account 是否属于同一排序组（Priority + LastUsedAt）
func sameAccountGroup(a, b *Account) bool {
	if a.Priority != b.Priority {
		return false
	}
	return sameLastUsedAt(a.LastUsedAt, b.LastUsedAt)
}

// sameLastUsedAt 判断两个 LastUsedAt 是否相同（精度到秒）
func sameLastUsedAt(a, b *time.Time) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return a.Unix() == b.Unix()
	}
}


// sortAccountsByPriorityOnly 仅按优先级排序
func sortAccountsByPriorityOnly(accounts []*Account, preferOAuth bool) {
	sort.SliceStable(accounts, func(i, j int) bool {
		a, b := accounts[i], accounts[j]
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		if preferOAuth && a.Type != b.Type {
			return a.Type == AccountTypeOAuth
		}
		return false
	})
}

// shuffleWithinPriority 在同优先级内随机打乱顺序
func shuffleWithinPriority(accounts []*Account) {
	if len(accounts) <= 1 {
		return
	}
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	start := 0
	for start < len(accounts) {
		priority := accounts[start].Priority
		end := start + 1
		for end < len(accounts) && accounts[end].Priority == priority {
			end++
		}
		// 对 [start, end) 范围内的账户随机打乱
		if end-start > 1 {
			r.Shuffle(end-start, func(i, j int) {
				accounts[start+i], accounts[start+j] = accounts[start+j], accounts[start+i]
			})
		}
		start = end
	}
}

// selectAccountForModelWithPlatform 选择单平台账户（完全隔离）
func (s *GatewayService) selectAccountForModelWithPlatform(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, platform string) (*Account, error) {
	preferOAuth := platform == PlatformGemini
	routingAccountIDs := s.routingAccountIDsForRequest(ctx, groupID, requestedModel, platform)

	var accounts []Account
	accountsLoaded := false

	// ============ Model Routing (legacy path): apply before sticky session ============
	// Model routing is honored so switching model can switch upstream account within the same sticky session.
	if len(routingAccountIDs) > 0 {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] legacy routed begin: group_id=%v model=%s platform=%s session=%s routed_ids=%v",
				derefGroupID(groupID), requestedModel, platform, shortSessionHash(sessionHash), routingAccountIDs)
		}
		// 1) Sticky session only applies if the bound account is within the routing set.
		if sessionHash != "" && s.cache != nil {
			accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
			if err == nil && accountID > 0 && containsInt64(routingAccountIDs, accountID) {
				if _, excluded := excludedIDs[accountID]; !excluded {
					account, err := s.getSchedulableAccount(ctx, accountID)
					// 检查账号分组归属和平台匹配（确保粘性会话不会跨分组或跨平台）
					if err == nil {
						clearSticky := shouldClearStickySession(account, requestedModel)
						if clearSticky {
							_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
						}
						if !clearSticky && s.isAccountInGroup(account, groupID) && (requestedModel == "" || s.isModelSupportedByAccountWithContext(ctx, account, requestedModel)) && s.isAccountSchedulableForModelSelection(ctx, account, requestedModel) && s.isAccountSchedulableForQuota(account) && s.isAccountSchedulableForWindowCost(ctx, account, true) && s.isAccountSchedulableForRPM(ctx, account, true) {
							if s.debugModelRoutingEnabled() {
								logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] legacy routed sticky hit: group_id=%v model=%s session=%s account=%d", derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), accountID)
							}
							return account, nil
						}
					}
				}
			}
		}

		// 2) Select an account from the routed candidates.
		forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
		if hasForcePlatform && forcePlatform == "" {
			hasForcePlatform = false
		}
		var err error
		accounts, _, err = s.listSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
		accountsLoaded = true

		// 提前预取窗口费用+RPM 计数，确保 routing 段内的调度检查调用能命中缓存
		ctx = s.withWindowCostPrefetch(ctx, accounts)
		ctx = s.withRPMPrefetch(ctx, accounts)

		routingSet := make(map[int64]struct{}, len(routingAccountIDs))
		for _, id := range routingAccountIDs {
			if id > 0 {
				routingSet[id] = struct{}{}
			}
		}

		var selected *Account
		for i := range accounts {
			acc := &accounts[i]
			if _, ok := routingSet[acc.ID]; !ok {
				continue
			}
			if _, excluded := excludedIDs[acc.ID]; excluded {
				continue
			}
			// Scheduler snapshots can be temporarily stale; re-check schedulability here to
			// avoid selecting accounts that were recently rate-limited/overloaded.
			if !s.isAccountSchedulableForSelection(acc) {
				continue
			}
			if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
				continue
			}
			if !s.isAccountSchedulableForModelSelection(ctx, acc, requestedModel) {
				continue
			}
			if !s.isAccountSchedulableForQuota(acc) {
				continue
			}
			if !s.isAccountSchedulableForWindowCost(ctx, acc, false) {
				continue
			}
			if !s.isAccountSchedulableForRPM(ctx, acc, false) {
				continue
			}
			if selected == nil {
				selected = acc
				continue
			}
			if acc.Priority < selected.Priority {
				selected = acc
			} else if acc.Priority == selected.Priority {
				switch {
				case acc.LastUsedAt == nil && selected.LastUsedAt != nil:
					selected = acc
				case acc.LastUsedAt != nil && selected.LastUsedAt == nil:
					// keep selected (never used is preferred)
				case acc.LastUsedAt == nil && selected.LastUsedAt == nil:
					if preferOAuth && acc.Type != selected.Type && acc.Type == AccountTypeOAuth {
						selected = acc
					}
				default:
					if acc.LastUsedAt.Before(*selected.LastUsedAt) {
						selected = acc
					}
				}
			}
		}

		if selected != nil {
			if sessionHash != "" && s.cache != nil {
				if err := s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), sessionHash, selected.ID, stickySessionTTL); err != nil {
					logger.LegacyPrintf("service.gateway", "set session account failed: session=%s account_id=%d err=%v", sessionHash, selected.ID, err)
				}
			}
			if s.debugModelRoutingEnabled() {
				logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] legacy routed select: group_id=%v model=%s session=%s account=%d", derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), selected.ID)
			}
			return selected, nil
		}
		logger.LegacyPrintf("service.gateway", "[ModelRouting] No routed accounts available for model=%s, falling back to normal selection", requestedModel)
	}

	// 1. 查询粘性会话
	if sessionHash != "" && s.cache != nil {
		accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
		if err == nil && accountID > 0 {
			if _, excluded := excludedIDs[accountID]; !excluded {
				account, err := s.getSchedulableAccount(ctx, accountID)
				// 检查账号分组归属（确保粘性会话不会跨分组）
				if err == nil {
					clearSticky := shouldClearStickySession(account, requestedModel)
					if clearSticky {
						_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
					}
					if !clearSticky && s.isAccountInGroup(account, groupID) && (requestedModel == "" || s.isModelSupportedByAccountWithContext(ctx, account, requestedModel)) && s.isAccountSchedulableForModelSelection(ctx, account, requestedModel) && s.isAccountSchedulableForQuota(account) && s.isAccountSchedulableForWindowCost(ctx, account, true) && s.isAccountSchedulableForRPM(ctx, account, true) {
						return account, nil
					}
				}
			}
		}
	}

	// 2. 获取可调度账号列表（单平台）
	if !accountsLoaded {
		forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
		if hasForcePlatform && forcePlatform == "" {
			hasForcePlatform = false
		}
		var err error
		accounts, _, err = s.listSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
	}

	// 批量预取窗口费用+RPM 计数，避免逐个账号查询（N+1）
	ctx = s.withWindowCostPrefetch(ctx, accounts)
	ctx = s.withRPMPrefetch(ctx, accounts)

	// 3. 按优先级+最久未用选择（考虑模型支持）
	var selected *Account
	for i := range accounts {
		acc := &accounts[i]
		if _, excluded := excludedIDs[acc.ID]; excluded {
			continue
		}
		// Scheduler snapshots can be temporarily stale; re-check schedulability here to
		// avoid selecting accounts that were recently rate-limited/overloaded.
		if !s.isAccountSchedulableForSelection(acc) {
			continue
		}
		if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForModelSelection(ctx, acc, requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForQuota(acc) {
			continue
		}
		if !s.isAccountSchedulableForWindowCost(ctx, acc, false) {
			continue
		}
		if !s.isAccountSchedulableForRPM(ctx, acc, false) {
			continue
		}
		if selected == nil {
			selected = acc
			continue
		}
		if acc.Priority < selected.Priority {
			selected = acc
		} else if acc.Priority == selected.Priority {
			switch {
			case acc.LastUsedAt == nil && selected.LastUsedAt != nil:
				selected = acc
			case acc.LastUsedAt != nil && selected.LastUsedAt == nil:
				// keep selected (never used is preferred)
			case acc.LastUsedAt == nil && selected.LastUsedAt == nil:
				if preferOAuth && acc.Type != selected.Type && acc.Type == AccountTypeOAuth {
					selected = acc
				}
			default:
				if acc.LastUsedAt.Before(*selected.LastUsedAt) {
					selected = acc
				}
			}
		}
	}

	if selected == nil {
		stats := s.logDetailedSelectionFailure(ctx, groupID, sessionHash, requestedModel, platform, accounts, excludedIDs, false)
		if requestedModel != "" {
			return nil, fmt.Errorf("%w supporting model: %s (%s)", ErrNoAvailableAccounts, requestedModel, summarizeSelectionFailureStats(stats))
		}
		return nil, ErrNoAvailableAccounts
	}

	// 4. 建立粘性绑定
	if sessionHash != "" && s.cache != nil {
		if err := s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), sessionHash, selected.ID, stickySessionTTL); err != nil {
			logger.LegacyPrintf("service.gateway", "set session account failed: session=%s account_id=%d err=%v", sessionHash, selected.ID, err)
		}
	}

	return selected, nil
}


type selectionFailureStats struct {
	Total              int
	Eligible           int
	Excluded           int
	Unschedulable      int
	PlatformFiltered   int
	ModelUnsupported   int
	ModelRateLimited   int
	SamplePlatformIDs  []int64
	SampleMappingIDs   []int64
	SampleRateLimitIDs []string
}

type selectionFailureDiagnosis struct {
	Category string
	Detail   string
}

func (s *GatewayService) logDetailedSelectionFailure(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	platform string,
	accounts []Account,
	excludedIDs map[int64]struct{},
	allowMixedScheduling bool,
) selectionFailureStats {
	stats := s.collectSelectionFailureStats(ctx, accounts, requestedModel, platform, excludedIDs, allowMixedScheduling)
	logger.LegacyPrintf(
		"service.gateway",
		"[SelectAccountDetailed] group_id=%v model=%s platform=%s session=%s total=%d eligible=%d excluded=%d unschedulable=%d platform_filtered=%d model_unsupported=%d model_rate_limited=%d sample_platform_filtered=%v sample_model_unsupported=%v sample_model_rate_limited=%v",
		derefGroupID(groupID),
		requestedModel,
		platform,
		shortSessionHash(sessionHash),
		stats.Total,
		stats.Eligible,
		stats.Excluded,
		stats.Unschedulable,
		stats.PlatformFiltered,
		stats.ModelUnsupported,
		stats.ModelRateLimited,
		stats.SamplePlatformIDs,
		stats.SampleMappingIDs,
		stats.SampleRateLimitIDs,
	)
	return stats
}

func (s *GatewayService) collectSelectionFailureStats(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	platform string,
	excludedIDs map[int64]struct{},
	allowMixedScheduling bool,
) selectionFailureStats {
	stats := selectionFailureStats{
		Total: len(accounts),
	}

	for i := range accounts {
		acc := &accounts[i]
		diagnosis := s.diagnoseSelectionFailure(ctx, acc, requestedModel, platform, excludedIDs, allowMixedScheduling)
		switch diagnosis.Category {
		case "excluded":
			stats.Excluded++
		case "unschedulable":
			stats.Unschedulable++
		case "platform_filtered":
			stats.PlatformFiltered++
			stats.SamplePlatformIDs = appendSelectionFailureSampleID(stats.SamplePlatformIDs, acc.ID)
		case "model_unsupported":
			stats.ModelUnsupported++
			stats.SampleMappingIDs = appendSelectionFailureSampleID(stats.SampleMappingIDs, acc.ID)
		case "model_rate_limited":
			stats.ModelRateLimited++
			remaining := acc.GetRateLimitRemainingTimeWithContext(ctx, requestedModel).Truncate(time.Second)
			stats.SampleRateLimitIDs = appendSelectionFailureRateSample(stats.SampleRateLimitIDs, acc.ID, remaining)
		default:
			stats.Eligible++
		}
	}

	return stats
}

func (s *GatewayService) diagnoseSelectionFailure(
	ctx context.Context,
	acc *Account,
	requestedModel string,
	platform string,
	excludedIDs map[int64]struct{},
	allowMixedScheduling bool,
) selectionFailureDiagnosis {
	if acc == nil {
		return selectionFailureDiagnosis{Category: "unschedulable", Detail: "account_nil"}
	}
	if _, excluded := excludedIDs[acc.ID]; excluded {
		return selectionFailureDiagnosis{Category: "excluded"}
	}
	if !s.isAccountSchedulableForSelection(acc) {
		return selectionFailureDiagnosis{Category: "unschedulable", Detail: "generic_unschedulable"}
	}
	if isPlatformFilteredForSelection(acc, platform, allowMixedScheduling) {
		return selectionFailureDiagnosis{
			Category: "platform_filtered",
			Detail:   fmt.Sprintf("account_platform=%s requested_platform=%s", acc.Platform, strings.TrimSpace(platform)),
		}
	}
	if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
		return selectionFailureDiagnosis{
			Category: "model_unsupported",
			Detail:   fmt.Sprintf("model=%s", requestedModel),
		}
	}
	if !s.isAccountSchedulableForModelSelection(ctx, acc, requestedModel) {
		remaining := acc.GetRateLimitRemainingTimeWithContext(ctx, requestedModel).Truncate(time.Second)
		return selectionFailureDiagnosis{
			Category: "model_rate_limited",
			Detail:   fmt.Sprintf("remaining=%s", remaining),
		}
	}
	return selectionFailureDiagnosis{Category: "eligible"}
}

func isPlatformFilteredForSelection(acc *Account, platform string, _ bool) bool {
	if acc == nil {
		return true
	}
	if strings.TrimSpace(platform) == "" {
		return false
	}
	return acc.Platform != platform
}

func appendSelectionFailureSampleID(samples []int64, id int64) []int64 {
	const limit = 5
	if len(samples) >= limit {
		return samples
	}
	return append(samples, id)
}

func appendSelectionFailureRateSample(samples []string, accountID int64, remaining time.Duration) []string {
	const limit = 5
	if len(samples) >= limit {
		return samples
	}
	return append(samples, fmt.Sprintf("%d(%s)", accountID, remaining))
}

func summarizeSelectionFailureStats(stats selectionFailureStats) string {
	return fmt.Sprintf(
		"total=%d eligible=%d excluded=%d unschedulable=%d platform_filtered=%d model_unsupported=%d model_rate_limited=%d",
		stats.Total,
		stats.Eligible,
		stats.Excluded,
		stats.Unschedulable,
		stats.PlatformFiltered,
		stats.ModelUnsupported,
		stats.ModelRateLimited,
	)
}

// isModelSupportedByAccountWithContext checks if the account supports the requested model.
func (s *GatewayService) isModelSupportedByAccountWithContext(ctx context.Context, account *Account, requestedModel string) bool {
	return s.isModelSupportedByAccount(account, requestedModel)
}

// isModelSupportedByAccount checks model support using the account's model mapping.
func (s *GatewayService) isModelSupportedByAccount(account *Account, requestedModel string) bool {
	return account.IsModelSupported(requestedModel)
}


// GetAccessToken 获取账号凭证
func (s *GatewayService) GetAccessToken(ctx context.Context, account *Account) (string, string, error) {
	switch account.Type {
	case AccountTypeOAuth, AccountTypeSetupToken:
		// Both oauth and setup-token use OAuth token flow
		return s.getOAuthToken(ctx, account)
	case AccountTypeAPIKey:
		apiKey := account.GetCredential("api_key")
		if apiKey == "" {
			return "", "", errors.New("api_key not found in credentials")
		}
		return apiKey, "apikey", nil
	default:
		return "", "", fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func (s *GatewayService) getOAuthToken(ctx context.Context, account *Account) (string, string, error) {
	// 对于 Anthropic OAuth 账号，使用 ClaudeTokenProvider 获取缓存的 token
	if account.Platform == PlatformAnthropic && account.Type == AccountTypeOAuth && s.claudeTokenProvider != nil {
		accessToken, err := s.claudeTokenProvider.GetAccessToken(ctx, account)
		if err != nil {
			return "", "", err
		}
		return accessToken, "oauth", nil
	}

	// 其他情况（Gemini 有自己的 TokenProvider，setup-token 类型等）直接从账号读取
	accessToken := account.GetCredential("access_token")
	if accessToken == "" {
		return "", "", errors.New("access_token not found in credentials")
	}
	// Token刷新由后台 TokenRefreshService 处理，此处只返回当前token
	return accessToken, "oauth", nil
}

// 重试相关常量
const (
	// 最大尝试次数（包含首次请求）。过多重试会导致请求堆积与资源耗尽。
	maxRetryAttempts = 5

	// 指数退避：第 N 次失败后的等待 = retryBaseDelay * 2^(N-1)，并且上限为 retryMaxDelay。
	retryBaseDelay = 300 * time.Millisecond
	retryMaxDelay  = 3 * time.Second

	// 最大重试耗时（包含请求本身耗时 + 退避等待时间）。
	// 用于防止极端情况下 goroutine 长时间堆积导致资源耗尽。
	maxRetryElapsed = 10 * time.Second
)

func (s *GatewayService) shouldRetryUpstreamError(account *Account, statusCode int) bool {
	// OAuth/Setup Token 账号：仅 403 重试
	if account.IsOAuth() {
		return statusCode == 403
	}

	// API Key 账号：未配置的错误码重试
	return !account.ShouldHandleErrorCode(statusCode)
}

// shouldFailoverUpstreamError determines whether an upstream error should trigger account failover.
func (s *GatewayService) shouldFailoverUpstreamError(statusCode int) bool {
	switch statusCode {
	case 401, 403, 429, 529:
		return true
	default:
		return statusCode >= 500
	}
}

func retryBackoffDelay(attempt int) time.Duration {
	// attempt 从 1 开始，表示第 attempt 次请求刚失败，需要等待后进行第 attempt+1 次请求。
	if attempt <= 0 {
		return retryBaseDelay
	}
	delay := retryBaseDelay * time.Duration(1<<(attempt-1))
	if delay > retryMaxDelay {
		return retryMaxDelay
	}
	return delay
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// isClaudeCodeClient 判断请求是否来自 Claude Code 客户端
// 简化判断：User-Agent 匹配 + metadata.user_id 存在
func isClaudeCodeClient(userAgent string, metadataUserID string) bool {
	if metadataUserID == "" {
		return false
	}
	return claudeCliUserAgentRe.MatchString(userAgent)
}

func isClaudeCodeRequest(ctx context.Context, c *gin.Context, parsed *ParsedRequest) bool {
	if IsClaudeCodeClient(ctx) {
		return true
	}
	if parsed == nil || c == nil {
		return false
	}
	return isClaudeCodeClient(c.GetHeader("User-Agent"), parsed.MetadataUserID)
}

// normalizeSystemParam 将 json.RawMessage 类型的 system 参数转为标准 Go 类型（string / []any / nil），
// 避免 type switch 中 json.RawMessage（底层 []byte）无法匹配 case string / case []any / case nil 的问题。
// 这是 Go 的 typed nil 陷阱：(json.RawMessage, nil) ≠ (nil, nil)。
func normalizeSystemParam(system any) any {
	raw, ok := system.(json.RawMessage)
	if !ok {
		return system
	}
	if len(raw) == 0 {
		return nil
	}
	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil
	}
	return parsed
}

// systemIncludesClaudeCodePrompt 检查 system 中是否已包含 Claude Code 提示词
// 使用前缀匹配支持多种变体（标准版、Agent SDK 版等）
func (s *GatewayService) handleFailoverSideEffects(ctx context.Context, resp *http.Response, account *Account) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, body)
}

// handleRetryExhaustedError 处理重试耗尽后的错误
// OAuth 403：标记账号异常
// API Key 未配置错误码：仅返回错误，不标记账号
// RecordUsageInput 记录使用量的输入参数
type RecordUsageInput struct {
	Result             *ForwardResult
	APIKey             *APIKey
	User               *User
	Account            *Account
	Subscription       *UserSubscription  // 可选：订阅信息
	InboundEndpoint    string             // 入站端点（客户端请求路径）
	UpstreamEndpoint   string             // 上游端点（标准化后的上游路径）
	UserAgent          string             // 请求的 User-Agent
	IPAddress          string             // 请求的客户端 IP 地址
	RequestPayloadHash string             // 请求体语义哈希，用于降低 request_id 误复用时的静默误去重风险
	ForceCacheBilling  bool               // 强制缓存计费：将 input_tokens 转为 cache_read 计费（用于粘性会话切换）
	APIKeyService      APIKeyQuotaUpdater // 可选：用于更新API Key配额
}

// APIKeyQuotaUpdater defines the interface for updating API Key quota and rate limit usage
type APIKeyQuotaUpdater interface {
	UpdateQuotaUsed(ctx context.Context, apiKeyID int64, cost float64) error
	UpdateRateLimitUsage(ctx context.Context, apiKeyID int64, cost float64) error
}

type apiKeyAuthCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
}

type usageLogBestEffortWriter interface {
	CreateBestEffort(ctx context.Context, log *UsageLog) error
}

// postUsageBillingParams 统一扣费所需的参数
type postUsageBillingParams struct {
	Cost               *CostBreakdown
	User               *User
	APIKey             *APIKey
	Account            *Account
	Subscription       *UserSubscription
	RequestPayloadHash string
	IsSubscriptionBill bool
	APIKeyService      APIKeyQuotaUpdater
}

// postUsageBilling 统一处理使用量记录后的扣费逻辑：
//   - 订阅/余额扣费
//   - API Key 配额更新
//   - API Key 限速用量更新
//   - 账号配额用量更新（账号口径：TotalCost × 账号计费倍率）
func postUsageBilling(ctx context.Context, p *postUsageBillingParams, deps *billingDeps) {
	billingCtx, cancel := detachedBillingContext(ctx)
	defer cancel()

	cost := p.Cost

	// 1. 订阅 / 余额扣费
	if p.IsSubscriptionBill {
		if cost.TotalCost > 0 {
			if err := deps.userSubRepo.IncrementUsage(billingCtx, p.Subscription.ID, cost.TotalCost); err != nil {
				slog.Error("increment subscription usage failed", "subscription_id", p.Subscription.ID, "error", err)
			}
			deps.billingCacheService.QueueUpdateSubscriptionUsage(p.User.ID, *p.APIKey.GroupID, cost.TotalCost)
		}
	} else {
		if cost.ActualCost > 0 {
			if err := deps.userRepo.DeductBalance(billingCtx, p.User.ID, cost.ActualCost); err != nil {
				slog.Error("deduct balance failed", "user_id", p.User.ID, "error", err)
			}
			deps.billingCacheService.QueueDeductBalance(p.User.ID, cost.ActualCost)
		}
	}

	// 2. API Key 配额
	if cost.ActualCost > 0 && p.APIKey.Quota > 0 && p.APIKeyService != nil {
		if err := p.APIKeyService.UpdateQuotaUsed(billingCtx, p.APIKey.ID, cost.ActualCost); err != nil {
			slog.Error("update api key quota failed", "api_key_id", p.APIKey.ID, "error", err)
		}
	}

	// 3. API Key 限速用量
	if cost.ActualCost > 0 && p.APIKey.HasRateLimits() && p.APIKeyService != nil {
		if err := p.APIKeyService.UpdateRateLimitUsage(billingCtx, p.APIKey.ID, cost.ActualCost); err != nil {
			slog.Error("update api key rate limit usage failed", "api_key_id", p.APIKey.ID, "error", err)
		}
	}

	// 4. 账号配额用量
	if cost.TotalCost > 0 && p.Account.Type == AccountTypeAPIKey && p.Account.HasAnyQuotaLimit() {
		if err := deps.accountRepo.IncrementQuotaUsed(billingCtx, p.Account.ID, cost.TotalCost); err != nil {
			slog.Error("increment account quota used failed", "account_id", p.Account.ID, "cost", cost.TotalCost, "error", err)
		}
	}

	finalizePostUsageBilling(p, deps)
}

func resolveUsageBillingRequestID(ctx context.Context, upstreamRequestID string) string {
	if ctx != nil {
		if clientRequestID, _ := ctx.Value(ctxkey.ClientRequestID).(string); strings.TrimSpace(clientRequestID) != "" {
			return "client:" + strings.TrimSpace(clientRequestID)
		}
		if requestID, _ := ctx.Value(ctxkey.RequestID).(string); strings.TrimSpace(requestID) != "" {
			return "local:" + strings.TrimSpace(requestID)
		}
	}
	if requestID := strings.TrimSpace(upstreamRequestID); requestID != "" {
		return requestID
	}
	return "generated:" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func resolveUsageBillingPayloadFingerprint(ctx context.Context, requestPayloadHash string) string {
	if payloadHash := strings.TrimSpace(requestPayloadHash); payloadHash != "" {
		return payloadHash
	}
	if ctx != nil {
		if clientRequestID, _ := ctx.Value(ctxkey.ClientRequestID).(string); strings.TrimSpace(clientRequestID) != "" {
			return "client:" + strings.TrimSpace(clientRequestID)
		}
		if requestID, _ := ctx.Value(ctxkey.RequestID).(string); strings.TrimSpace(requestID) != "" {
			return "local:" + strings.TrimSpace(requestID)
		}
	}
	return ""
}

func buildUsageBillingCommand(requestID string, usageLog *UsageLog, p *postUsageBillingParams) *UsageBillingCommand {
	if p == nil || p.Cost == nil || p.APIKey == nil || p.User == nil || p.Account == nil {
		return nil
	}

	cmd := &UsageBillingCommand{
		RequestID:          requestID,
		APIKeyID:           p.APIKey.ID,
		UserID:             p.User.ID,
		AccountID:          p.Account.ID,
		AccountType:        p.Account.Type,
		RequestPayloadHash: strings.TrimSpace(p.RequestPayloadHash),
	}
	if usageLog != nil {
		cmd.Model = usageLog.Model
		cmd.BillingType = usageLog.BillingType
		cmd.InputTokens = usageLog.InputTokens
		cmd.OutputTokens = usageLog.OutputTokens
		cmd.CacheCreationTokens = usageLog.CacheCreationTokens
		cmd.CacheReadTokens = usageLog.CacheReadTokens
		cmd.ImageCount = usageLog.ImageCount
		if usageLog.MediaType != nil {
			cmd.MediaType = *usageLog.MediaType
		}
		if usageLog.ServiceTier != nil {
			cmd.ServiceTier = *usageLog.ServiceTier
		}
		if usageLog.ReasoningEffort != nil {
			cmd.ReasoningEffort = *usageLog.ReasoningEffort
		}
		if usageLog.SubscriptionID != nil {
			cmd.SubscriptionID = usageLog.SubscriptionID
		}
	}

	if p.IsSubscriptionBill && p.Subscription != nil && p.Cost.TotalCost > 0 {
		cmd.SubscriptionID = &p.Subscription.ID
		cmd.SubscriptionCost = p.Cost.TotalCost
	} else if p.Cost.ActualCost > 0 {
		cmd.BalanceCost = p.Cost.ActualCost
	}

	if p.Cost.ActualCost > 0 && p.APIKey.Quota > 0 && p.APIKeyService != nil {
		cmd.APIKeyQuotaCost = p.Cost.ActualCost
	}
	if p.Cost.ActualCost > 0 && p.APIKey.HasRateLimits() && p.APIKeyService != nil {
		cmd.APIKeyRateLimitCost = p.Cost.ActualCost
	}
	if p.Cost.TotalCost > 0 && p.Account.Type == AccountTypeAPIKey && p.Account.HasAnyQuotaLimit() {
		cmd.AccountQuotaCost = p.Cost.TotalCost
	}

	cmd.Normalize()
	return cmd
}

func applyUsageBilling(ctx context.Context, requestID string, usageLog *UsageLog, p *postUsageBillingParams, deps *billingDeps, repo UsageBillingRepository) (bool, error) {
	if p == nil || deps == nil {
		return false, nil
	}

	cmd := buildUsageBillingCommand(requestID, usageLog, p)
	if cmd == nil || cmd.RequestID == "" || repo == nil {
		postUsageBilling(ctx, p, deps)
		return true, nil
	}

	billingCtx, cancel := detachedBillingContext(ctx)
	defer cancel()

	result, err := repo.Apply(billingCtx, cmd)
	if err != nil {
		return false, err
	}

	if result == nil || !result.Applied {
		deps.deferredService.ScheduleLastUsedUpdate(p.Account.ID)
		return false, nil
	}

	if result.APIKeyQuotaExhausted {
		if invalidator, ok := p.APIKeyService.(apiKeyAuthCacheInvalidator); ok && p.APIKey != nil && p.APIKey.Key != "" {
			invalidator.InvalidateAuthCacheByKey(billingCtx, p.APIKey.Key)
		}
	}

	finalizePostUsageBilling(p, deps)
	return true, nil
}

func finalizePostUsageBilling(p *postUsageBillingParams, deps *billingDeps) {
	if p == nil || p.Cost == nil || deps == nil {
		return
	}

	if p.IsSubscriptionBill {
		if p.Cost.TotalCost > 0 && p.User != nil && p.APIKey != nil && p.APIKey.GroupID != nil {
			deps.billingCacheService.QueueUpdateSubscriptionUsage(p.User.ID, *p.APIKey.GroupID, p.Cost.TotalCost)
		}
	} else if p.Cost.ActualCost > 0 && p.User != nil {
		deps.billingCacheService.QueueDeductBalance(p.User.ID, p.Cost.ActualCost)
	}

	if p.Cost.ActualCost > 0 && p.APIKey != nil && p.APIKey.HasRateLimits() {
		deps.billingCacheService.QueueUpdateAPIKeyRateLimitUsage(p.APIKey.ID, p.Cost.ActualCost)
	}

	deps.deferredService.ScheduleLastUsedUpdate(p.Account.ID)
}

func detachedBillingContext(ctx context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if ctx != nil {
		base = context.WithoutCancel(ctx)
	}
	return context.WithTimeout(base, postUsageBillingTimeout)
}

func detachStreamUpstreamContext(ctx context.Context, stream bool) (context.Context, context.CancelFunc) {
	if !stream {
		return ctx, func() {}
	}
	if ctx == nil {
		return context.Background(), func() {}
	}
	return context.WithoutCancel(ctx), func() {}
}

// billingDeps 扣费逻辑依赖的服务（由各 gateway service 提供）
type billingDeps struct {
	accountRepo         AccountRepository
	userRepo            UserRepository
	userSubRepo         UserSubscriptionRepository
	billingCacheService *BillingCacheService
	deferredService     *DeferredService
}

func (s *GatewayService) billingDeps() *billingDeps {
	return &billingDeps{
		accountRepo:         s.accountRepo,
		userRepo:            s.userRepo,
		userSubRepo:         s.userSubRepo,
		billingCacheService: s.billingCacheService,
		deferredService:     s.deferredService,
	}
}

func writeUsageLogBestEffort(ctx context.Context, repo UsageLogRepository, usageLog *UsageLog, logKey string) {
	if repo == nil || usageLog == nil {
		return
	}
	usageCtx, cancel := detachedBillingContext(ctx)
	defer cancel()

	if writer, ok := repo.(usageLogBestEffortWriter); ok {
		if err := writer.CreateBestEffort(usageCtx, usageLog); err != nil {
			logger.LegacyPrintf(logKey, "Create usage log failed: %v", err)
			if IsUsageLogCreateDropped(err) {
				return
			}
			if _, syncErr := repo.Create(usageCtx, usageLog); syncErr != nil {
				logger.LegacyPrintf(logKey, "Create usage log sync fallback failed: %v", syncErr)
			}
		}
		return
	}

	if _, err := repo.Create(usageCtx, usageLog); err != nil {
		logger.LegacyPrintf(logKey, "Create usage log failed: %v", err)
	}
}

// RecordUsage 记录使用量并扣费（或更新订阅用量）
func (s *GatewayService) RecordUsage(ctx context.Context, input *RecordUsageInput) error {
	result := input.Result
	apiKey := input.APIKey
	user := input.User
	account := input.Account
	subscription := input.Subscription

	// 强制缓存计费：将 input_tokens 转为 cache_read_input_tokens
	// 用于粘性会话切换时的特殊计费处理
	if input.ForceCacheBilling && result.Usage.InputTokens > 0 {
		logger.LegacyPrintf("service.gateway", "force_cache_billing: %d input_tokens → cache_read_input_tokens (account=%d)",
			result.Usage.InputTokens, account.ID)
		result.Usage.CacheReadInputTokens += result.Usage.InputTokens
		result.Usage.InputTokens = 0
	}

	// Cache TTL Override: 确保计费时 token 分类与账号设置一致
	cacheTTLOverridden := false
	if account.IsCacheTTLOverrideEnabled() {
		applyCacheTTLOverride(&result.Usage, account.GetCacheTTLOverrideTarget())
		cacheTTLOverridden = (result.Usage.CacheCreation5mTokens + result.Usage.CacheCreation1hTokens) > 0
	}

	var cost *CostBreakdown
	billingModel := forwardResultBillingModel(result.Model, result.UpstreamModel)

	// 根据请求类型选择计费方式
	if result.MediaType == "prompt" {
		cost = &CostBreakdown{}
	} else if result.ImageCount > 0 {
		// 图片生成计费
		var groupConfig *ImagePriceConfig
		if apiKey.Group != nil {
			groupConfig = &ImagePriceConfig{
				Price1K: apiKey.Group.ImagePrice1K,
				Price2K: apiKey.Group.ImagePrice2K,
				Price4K: apiKey.Group.ImagePrice4K,
			}
		}
		cost = s.billingService.CalculateImageCost(billingModel, result.ImageSize, result.ImageCount, groupConfig)
	} else {
		// Token 计费：仅使用自定义模型定价
		tokens := UsageTokens{
			InputTokens:           result.Usage.InputTokens,
			OutputTokens:          result.Usage.OutputTokens,
			CacheCreationTokens:   result.Usage.CacheCreationInputTokens,
			CacheReadTokens:       result.Usage.CacheReadInputTokens,
			CacheCreation5mTokens: result.Usage.CacheCreation5mTokens,
			CacheCreation1hTokens: result.Usage.CacheCreation1hTokens,
		}

		var customPricing *ModelPricingEntry
		if apiKey.GroupID != nil && apiKey.Group != nil && apiKey.Group.ModelPricing != nil {
			customPricing = apiKey.Group.ModelPricing.GetPricing(billingModel)
		}

		if customPricing != nil {
			cost = s.billingService.CalculateCostWithCustomPricing(billingModel, customPricing, tokens)
		} else {
			// 未配置自定义定价的模型不计费，记录警告便于运维发现漏配
			logger.L().Warn("billing: model has no custom pricing, zero cost applied",
				zap.String("model", billingModel),
				zap.Any("group_id", apiKey.GroupID),
			)
			cost = &CostBreakdown{}
		}
	}

	// 判断计费方式：订阅模式 vs 余额模式
	isSubscriptionBilling := subscription != nil && apiKey.Group != nil && apiKey.Group.IsSubscriptionType()
	billingType := BillingTypeBalance
	if isSubscriptionBilling {
		billingType = BillingTypeSubscription
	}

	// 创建使用日志
	durationMs := int(result.Duration.Milliseconds())
	var imageSize *string
	if result.ImageSize != "" {
		imageSize = &result.ImageSize
	}
	var mediaType *string
	if strings.TrimSpace(result.MediaType) != "" {
		mediaType = &result.MediaType
	}
	fixedRateMultiplier := 1.0
	requestID := resolveUsageBillingRequestID(ctx, result.RequestID)
	usageLog := &UsageLog{
		UserID:                user.ID,
		APIKeyID:              apiKey.ID,
		AccountID:             account.ID,
		RequestID:             requestID,
		Model:                 result.Model,
		RequestedModel:        result.Model,
		UpstreamModel:         optionalNonEqualStringPtr(result.UpstreamModel, result.Model),
		ReasoningEffort:       result.ReasoningEffort,
		InboundEndpoint:       optionalTrimmedStringPtr(input.InboundEndpoint),
		UpstreamEndpoint:      optionalTrimmedStringPtr(input.UpstreamEndpoint),
		InputTokens:           result.Usage.InputTokens,
		OutputTokens:          result.Usage.OutputTokens,
		CacheCreationTokens:   result.Usage.CacheCreationInputTokens,
		CacheReadTokens:       result.Usage.CacheReadInputTokens,
		CacheCreation5mTokens: result.Usage.CacheCreation5mTokens,
		CacheCreation1hTokens: result.Usage.CacheCreation1hTokens,
		InputCost:             cost.InputCost,
		OutputCost:            cost.OutputCost,
		CacheCreationCost:     cost.CacheCreationCost,
		CacheReadCost:         cost.CacheReadCost,
		TotalCost:             cost.TotalCost,
		ActualCost:            cost.ActualCost,
		RateMultiplier:        fixedRateMultiplier,
		AccountRateMultiplier: &fixedRateMultiplier,
		UpstreamCost:          cost.UpstreamCost,
		BillingType:           billingType,
		Stream:                result.Stream,
		DurationMs:            &durationMs,
		FirstTokenMs:          result.FirstTokenMs,
		ImageCount:            result.ImageCount,
		ImageSize:             imageSize,
		MediaType:             mediaType,
		CacheTTLOverridden:    cacheTTLOverridden,
		CreatedAt:             time.Now(),
	}

	// 添加 UserAgent
	if input.UserAgent != "" {
		usageLog.UserAgent = &input.UserAgent
	}

	// 添加 IPAddress
	if input.IPAddress != "" {
		usageLog.IPAddress = &input.IPAddress
	}

	// 添加分组和订阅关联
	if apiKey.GroupID != nil {
		usageLog.GroupID = apiKey.GroupID
	}
	if subscription != nil {
		usageLog.SubscriptionID = &subscription.ID
	}

	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		writeUsageLogBestEffort(ctx, s.usageLogRepo, usageLog, "service.gateway")
		logger.LegacyPrintf("service.gateway", "[SIMPLE MODE] Usage recorded (not billed): user=%d, tokens=%d", usageLog.UserID, usageLog.TotalTokens())
		s.deferredService.ScheduleLastUsedUpdate(account.ID)
		return nil
	}

	billingErr := func() error {
		_, err := applyUsageBilling(ctx, requestID, usageLog, &postUsageBillingParams{
			Cost:               cost,
			User:               user,
			APIKey:             apiKey,
			Account:            account,
			Subscription:       subscription,
			RequestPayloadHash: resolveUsageBillingPayloadFingerprint(ctx, input.RequestPayloadHash),
			IsSubscriptionBill: isSubscriptionBilling,
			APIKeyService:      input.APIKeyService,
		}, s.billingDeps(), s.usageBillingRepo)
		return err
	}()

	if billingErr != nil {
		return billingErr
	}
	writeUsageLogBestEffort(ctx, s.usageLogRepo, usageLog, "service.gateway")

	return nil
}

// RecordUsageLongContextInput 记录使用量的输入参数（支持长上下文双倍计费）
type RecordUsageLongContextInput struct {
	Result                *ForwardResult
	APIKey                *APIKey
	User                  *User
	Account               *Account
	Subscription          *UserSubscription  // 可选：订阅信息
	InboundEndpoint       string             // 入站端点（客户端请求路径）
	UpstreamEndpoint      string             // 上游端点（标准化后的上游路径）
	UserAgent             string             // 请求的 User-Agent
	IPAddress             string             // 请求的客户端 IP 地址
	RequestPayloadHash    string             // 请求体语义哈希，用于降低 request_id 误复用时的静默误去重风险
	LongContextThreshold  int                // 长上下文阈值（如 200000）
	LongContextMultiplier float64            // 超出阈值部分的倍率（如 2.0）
	ForceCacheBilling     bool               // 强制缓存计费：将 input_tokens 转为 cache_read 计费（用于粘性会话切换）
	APIKeyService         APIKeyQuotaUpdater // API Key 配额服务（可选）
}

// RecordUsageWithLongContext 记录使用量并扣费，支持长上下文双倍计费（用于 Gemini）
func (s *GatewayService) RecordUsageWithLongContext(ctx context.Context, input *RecordUsageLongContextInput) error {
	result := input.Result
	apiKey := input.APIKey
	user := input.User
	account := input.Account
	subscription := input.Subscription

	// 强制缓存计费：将 input_tokens 转为 cache_read_input_tokens
	// 用于粘性会话切换时的特殊计费处理
	if input.ForceCacheBilling && result.Usage.InputTokens > 0 {
		logger.LegacyPrintf("service.gateway", "force_cache_billing: %d input_tokens → cache_read_input_tokens (account=%d)",
			result.Usage.InputTokens, account.ID)
		result.Usage.CacheReadInputTokens += result.Usage.InputTokens
		result.Usage.InputTokens = 0
	}

	// Cache TTL Override: 确保计费时 token 分类与账号设置一致
	cacheTTLOverridden := false
	if account.IsCacheTTLOverrideEnabled() {
		applyCacheTTLOverride(&result.Usage, account.GetCacheTTLOverrideTarget())
		cacheTTLOverridden = (result.Usage.CacheCreation5mTokens + result.Usage.CacheCreation1hTokens) > 0
	}

	var cost *CostBreakdown
	billingModel := forwardResultBillingModel(result.Model, result.UpstreamModel)

	// 根据请求类型选择计费方式
	if result.ImageCount > 0 {
		// 图片生成计费
		var groupConfig *ImagePriceConfig
		if apiKey.Group != nil {
			groupConfig = &ImagePriceConfig{
				Price1K: apiKey.Group.ImagePrice1K,
				Price2K: apiKey.Group.ImagePrice2K,
				Price4K: apiKey.Group.ImagePrice4K,
			}
		}
		cost = s.billingService.CalculateImageCost(billingModel, result.ImageSize, result.ImageCount, groupConfig)
	} else {
		// Token 计费：仅使用自定义模型定价
		tokens := UsageTokens{
			InputTokens:           result.Usage.InputTokens,
			OutputTokens:          result.Usage.OutputTokens,
			CacheCreationTokens:   result.Usage.CacheCreationInputTokens,
			CacheReadTokens:       result.Usage.CacheReadInputTokens,
			CacheCreation5mTokens: result.Usage.CacheCreation5mTokens,
			CacheCreation1hTokens: result.Usage.CacheCreation1hTokens,
		}

		var customPricing *ModelPricingEntry
		if apiKey.GroupID != nil && apiKey.Group != nil && apiKey.Group.ModelPricing != nil {
			customPricing = apiKey.Group.ModelPricing.GetPricing(billingModel)
		}

		if customPricing != nil {
			cost = s.billingService.CalculateCostWithCustomPricing(billingModel, customPricing, tokens)
		} else {
			logger.L().Warn("billing: model has no custom pricing, zero cost applied",
				zap.String("model", billingModel),
				zap.Any("group_id", apiKey.GroupID),
			)
			cost = &CostBreakdown{}
		}
	}

	// 判断计费方式：订阅模式 vs 余额模式
	isSubscriptionBilling := subscription != nil && apiKey.Group != nil && apiKey.Group.IsSubscriptionType()
	billingType := BillingTypeBalance
	if isSubscriptionBilling {
		billingType = BillingTypeSubscription
	}

	// 创建使用日志
	durationMs := int(result.Duration.Milliseconds())
	var imageSize *string
	if result.ImageSize != "" {
		imageSize = &result.ImageSize
	}
	fixedRateMultiplier := 1.0
	requestID := resolveUsageBillingRequestID(ctx, result.RequestID)
	usageLog := &UsageLog{
		UserID:                user.ID,
		APIKeyID:              apiKey.ID,
		AccountID:             account.ID,
		RequestID:             requestID,
		Model:                 result.Model,
		RequestedModel:        result.Model,
		UpstreamModel:         optionalNonEqualStringPtr(result.UpstreamModel, result.Model),
		ReasoningEffort:       result.ReasoningEffort,
		InboundEndpoint:       optionalTrimmedStringPtr(input.InboundEndpoint),
		UpstreamEndpoint:      optionalTrimmedStringPtr(input.UpstreamEndpoint),
		InputTokens:           result.Usage.InputTokens,
		OutputTokens:          result.Usage.OutputTokens,
		CacheCreationTokens:   result.Usage.CacheCreationInputTokens,
		CacheReadTokens:       result.Usage.CacheReadInputTokens,
		CacheCreation5mTokens: result.Usage.CacheCreation5mTokens,
		CacheCreation1hTokens: result.Usage.CacheCreation1hTokens,
		InputCost:             cost.InputCost,
		OutputCost:            cost.OutputCost,
		CacheCreationCost:     cost.CacheCreationCost,
		CacheReadCost:         cost.CacheReadCost,
		TotalCost:             cost.TotalCost,
		ActualCost:            cost.ActualCost,
		RateMultiplier:        fixedRateMultiplier,
		AccountRateMultiplier: &fixedRateMultiplier,
		UpstreamCost:          cost.UpstreamCost,
		BillingType:           billingType,
		Stream:                result.Stream,
		DurationMs:            &durationMs,
		FirstTokenMs:          result.FirstTokenMs,
		ImageCount:            result.ImageCount,
		ImageSize:             imageSize,
		CacheTTLOverridden:    cacheTTLOverridden,
		CreatedAt:             time.Now(),
	}

	// 添加 UserAgent
	if input.UserAgent != "" {
		usageLog.UserAgent = &input.UserAgent
	}

	// 添加 IPAddress
	if input.IPAddress != "" {
		usageLog.IPAddress = &input.IPAddress
	}

	// 添加分组和订阅关联
	if apiKey.GroupID != nil {
		usageLog.GroupID = apiKey.GroupID
	}
	if subscription != nil {
		usageLog.SubscriptionID = &subscription.ID
	}

	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		writeUsageLogBestEffort(ctx, s.usageLogRepo, usageLog, "service.gateway")
		logger.LegacyPrintf("service.gateway", "[SIMPLE MODE] Usage recorded (not billed): user=%d, tokens=%d", usageLog.UserID, usageLog.TotalTokens())
		s.deferredService.ScheduleLastUsedUpdate(account.ID)
		return nil
	}

	billingErr := func() error {
		_, err := applyUsageBilling(ctx, requestID, usageLog, &postUsageBillingParams{
			Cost:               cost,
			User:               user,
			APIKey:             apiKey,
			Account:            account,
			Subscription:       subscription,
			RequestPayloadHash: resolveUsageBillingPayloadFingerprint(ctx, input.RequestPayloadHash),
			IsSubscriptionBill: isSubscriptionBilling,
			APIKeyService:      input.APIKeyService,
		}, s.billingDeps(), s.usageBillingRepo)
		return err
	}()

	if billingErr != nil {
		return billingErr
	}
	writeUsageLogBestEffort(ctx, s.usageLogRepo, usageLog, "service.gateway")

	return nil
}

// ForwardCountTokens 通过透传转发 count_tokens 请求到上游 API（不记录使用量）
func (s *GatewayService) ForwardCountTokens(ctx context.Context, c *gin.Context, account *Account, parsed *ParsedRequest) error {
	if parsed == nil {
		s.countTokensError(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return fmt.Errorf("parse request: empty request")
	}
	parsed.Stream = false
	_, err := s.ForwardPassthrough(ctx, c, account, "/v1/messages/count_tokens", parsed.Body, parsed)
	return err
}

// countTokensError 返回 count_tokens 错误响应
func (s *GatewayService) countTokensError(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

// buildCustomRelayURL 构建自定义中继转发 URL
// 在 path 后附加 beta=true 和可选的 proxy 查询参数
func (s *GatewayService) buildCustomRelayURL(baseURL, path string, account *Account) string {
	u := strings.TrimRight(baseURL, "/") + path + "?beta=true"
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL := account.Proxy.URL()
		if proxyURL != "" {
			u += "&proxy=" + url.QueryEscape(proxyURL)
		}
	}
	return u
}

func (s *GatewayService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.cfg != nil && !s.cfg.Security.URLAllowlist.Enabled {
		normalized, err := urlvalidator.ValidateURLFormat(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
		if err != nil {
			return "", fmt.Errorf("invalid base_url: %w", err)
		}
		return normalized, nil
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	return normalized, nil
}

// GetAvailableModels returns the list of models available for a group
// It aggregates model_mapping keys from all schedulable accounts in the group
func (s *GatewayService) GetAvailableModels(ctx context.Context, groupID *int64, platform string) []string {
	cacheKey := modelsListCacheKey(groupID, platform)
	if s.modelsListCache != nil {
		if cached, found := s.modelsListCache.Get(cacheKey); found {
			if models, ok := cached.([]string); ok {
				modelsListCacheHitTotal.Add(1)
				return cloneStringSlice(models)
			}
		}
	}
	modelsListCacheMissTotal.Add(1)

	var accounts []Account
	var err error

	if groupID != nil {
		accounts, err = s.accountRepo.ListSchedulableByGroupID(ctx, *groupID)
	} else {
		accounts, err = s.accountRepo.ListSchedulable(ctx)
	}

	if err != nil || len(accounts) == 0 {
		return nil
	}

	// Platform filter removed: in passthrough mode, all accounts in the group are valid regardless of platform.

	// Collect unique models from all accounts
	modelSet := make(map[string]struct{})
	hasAnyMapping := false

	for _, acc := range accounts {
		mapping := acc.GetModelMapping()
		if len(mapping) > 0 {
			hasAnyMapping = true
			for model := range mapping {
				modelSet[model] = struct{}{}
			}
		}
	}

	// If no account has model_mapping, return nil (use default)
	if !hasAnyMapping {
		if s.modelsListCache != nil {
			s.modelsListCache.Set(cacheKey, []string(nil), s.modelsListCacheTTL)
			modelsListCacheStoreTotal.Add(1)
		}
		return nil
	}

	// Convert to slice
	models := make([]string, 0, len(modelSet))
	for model := range modelSet {
		models = append(models, model)
	}
	sort.Strings(models)

	if s.modelsListCache != nil {
		s.modelsListCache.Set(cacheKey, cloneStringSlice(models), s.modelsListCacheTTL)
		modelsListCacheStoreTotal.Add(1)
	}
	return cloneStringSlice(models)
}

func (s *GatewayService) InvalidateAvailableModelsCache(groupID *int64, platform string) {
	if s == nil || s.modelsListCache == nil {
		return
	}

	normalizedPlatform := strings.TrimSpace(platform)
	// 完整匹配时精准失效；否则按维度批量失效。
	if groupID != nil && normalizedPlatform != "" {
		s.modelsListCache.Delete(modelsListCacheKey(groupID, normalizedPlatform))
		return
	}

	targetGroup := derefGroupID(groupID)
	for key := range s.modelsListCache.Items() {
		parts := strings.SplitN(key, "|", 2)
		if len(parts) != 2 {
			continue
		}
		groupPart, parseErr := strconv.ParseInt(parts[0], 10, 64)
		if parseErr != nil {
			continue
		}
		if groupID != nil && groupPart != targetGroup {
			continue
		}
		if normalizedPlatform != "" && parts[1] != normalizedPlatform {
			continue
		}
		s.modelsListCache.Delete(key)
	}
}

// reconcileCachedTokens 兼容 Kimi 等上游：
// 将 OpenAI 风格的 cached_tokens 映射到 Claude 标准的 cache_read_input_tokens
func reconcileCachedTokens(usage map[string]any) bool {
	if usage == nil {
		return false
	}
	cacheRead, _ := usage["cache_read_input_tokens"].(float64)
	if cacheRead > 0 {
		return false // 已有标准字段，无需处理
	}
	cached, _ := usage["cached_tokens"].(float64)
	if cached <= 0 {
		return false
	}
	usage["cache_read_input_tokens"] = cached
	return true
}

