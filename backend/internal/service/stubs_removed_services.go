package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

// Stub types for removed platform-specific services.
// These are kept as empty structs so that existing admin/wire code continues to compile.
// They no longer perform any real work.

type OAuthService struct{}
type OpenAIOAuthService struct{}
type GeminiOAuthService struct{}
type ClaudeTokenProvider struct{}
type GeminiTokenProvider struct{}
type AntigravityGatewayService struct{}
type AntigravityQuotaFetcher struct{}
type OpenAIGatewayService struct{}
type AntigravityOAuthService struct{}
type OAuthRefreshAPI struct{}
type TokenRefreshService struct{}
type SoraGatewayService struct{}
type SoraGenerationService struct{}
type SoraMediaStorage struct{}
type SoraMediaCleanupService struct{}

func (s *SoraMediaCleanupService) Start() {}
func (s *SoraMediaCleanupService) Stop()  {}
type SoraSDKClient struct{}

func (c *SoraSDKClient) SetAccountRepositories(_ ...interface{}) {}
type SoraQuotaService struct{}
type OpenAITokenProvider struct{}
type AntigravityTokenProvider struct{}
type BedrockSigner struct{}
type GeminiMessagesCompatService struct{}

type SoraAccountRepository interface {
	Upsert(ctx context.Context, accountID int64, data map[string]interface{}) error
}
type PrivacyClientFactory interface{}
type SoraClient interface{}

// --- Constructors for wire ---

func NewSoraGatewayService() *SoraGatewayService               { return &SoraGatewayService{} }
func NewOpenAIGatewayService() *OpenAIGatewayService { return &OpenAIGatewayService{} }

func (s *OpenAIGatewayService) CloseOpenAIWSPool() {}
func NewOAuthService() *OAuthService                            { return &OAuthService{} }
func NewOpenAIOAuthService() *OpenAIOAuthService                { return &OpenAIOAuthService{} }
func NewGeminiOAuthService() *GeminiOAuthService                { return &GeminiOAuthService{} }
func NewAntigravityOAuthService() *AntigravityOAuthService      { return &AntigravityOAuthService{} }
func NewOAuthRefreshAPI() *OAuthRefreshAPI                      { return nil }
func NewGeminiMessagesCompatService() *GeminiMessagesCompatService { return &GeminiMessagesCompatService{} }
func NewAntigravityGatewayService() *AntigravityGatewayService  { return &AntigravityGatewayService{} }
func NewAntigravityQuotaFetcher() *AntigravityQuotaFetcher      { return &AntigravityQuotaFetcher{} }

// --- Bedrock stubs ---

func bedrockRuntimeRegion(_ *Account) string                         { return "" }
func ResolveBedrockModelID(_ *Account, model string) (string, bool)  { return model, true }
func BuildBedrockURL(region, model string, stream ...bool) string    { return "" }
func NewBedrockSignerFromAccount(_ *Account) (*BedrockSigner, error) { return &BedrockSigner{}, nil }

func (s *BedrockSigner) SignRequest(_ context.Context, req *http.Request, _ []byte) error { return nil }

// --- Codex/OpenAI stubs ---

func syncOpenAICodexRateLimitFromExtra(_ context.Context, _ AccountRepository, _ *Account, _ time.Time) {}
func buildCodexUsageExtraUpdates(_ ...interface{}) map[string]interface{}                                { return nil }

// CodexRateLimitSnapshot stub
type CodexRateLimitSnapshot struct{}

func (s *CodexRateLimitSnapshot) Normalize() *CodexRateLimitNormalized { return nil }

type CodexRateLimitNormalized struct {
	Used5hPercent  *float64
	Used7dPercent  *float64
	Reset5hSeconds *int
	Reset7dSeconds *int
}

func ParseCodexRateLimitHeaders(_ http.Header) *CodexRateLimitSnapshot {
	return nil
}

func codexRateLimitResetAtFromSnapshot(_ *CodexRateLimitSnapshot, _ ...interface{}) *time.Time {
	return nil
}

const chatgptCodexURL = "https://chatgpt.com"
const codexCLIUserAgent = "codex-cli"

func isImageGenerationModel(_ string) bool { return false }

// --- AntigravityGatewayService methods ---

// AntigravityTestResult stub
type AntigravityTestResult struct {
	Text string
}

func (s *AntigravityGatewayService) TestConnection(_ context.Context, _ *Account, _ ...interface{}) (*AntigravityTestResult, error) {
	return &AntigravityTestResult{}, nil
}
func (s *AntigravityGatewayService) WriteMappedClaudeError(_ ...interface{}) error { return nil }

// --- GeminiTokenProvider methods ---

func (p *GeminiTokenProvider) GetAccessToken(_ context.Context, _ *Account) (string, error) {
	return "", nil
}

// --- AntigravityQuotaFetcher methods ---

func (f *AntigravityQuotaFetcher) CanFetch(_ *Account) bool { return false }
func (f *AntigravityQuotaFetcher) GetProxyURL(_ context.Context, _ *Account) string { return "" }
func (f *AntigravityQuotaFetcher) FetchQuota(_ context.Context, _ *Account, _ string) (*QuotaInfo, error) {
	return nil, nil
}

// QuotaInfo stub for AntigravityQuotaFetcher
type QuotaInfo struct {
	Used      float64
	Limit     float64
	UsageInfo *UsageInfo
}

const errorCodeRateLimited = "RATE_LIMITED"
const errorCodeNetworkError = "NETWORK_ERROR"
const errorCodeForbidden = "FORBIDDEN"

func classifyForbiddenType(_ string) string                { return "" }
func extractValidationURL(_ string) string                  { return "" }

const forbiddenTypeValidation = "validation"
const forbiddenTypeViolation = "violation"
const creditsExhaustedKey = "credits_exhausted"
const AntigravityPrivacySet = "set"

func shouldSkipOpenAIPrivacyEnsure(_ map[string]interface{}) bool { return true }
func disableOpenAITraining(_ ...interface{}) string               { return "" }
func setAntigravityPrivacy(_ ...interface{}) string               { return "" }
func applyAntigravityPrivacyMode(_ ...interface{})                {}

// --- TokenRefreshService methods ---

func (s *TokenRefreshService) SetSoraAccountRepo(_ SoraAccountRepository)    {}
func (s *TokenRefreshService) SetPrivacyDeps(_ PrivacyClientFactory, _ ProxyRepository) {}
func (s *TokenRefreshService) SetRefreshAPI(_ *OAuthRefreshAPI)                {}
func (s *TokenRefreshService) SetRefreshPolicy(_ interface{})                  {}
func (s *TokenRefreshService) Start()                                          {}
func (s *TokenRefreshService) Stop()                                           {}

// GeminiTokenCache interface for OAuth token caching
type GeminiTokenCache interface {
	GetAccessToken(ctx context.Context, cacheKey string) (string, error)
	SetAccessToken(ctx context.Context, cacheKey string, token string, ttl time.Duration) error
	DeleteAccessToken(ctx context.Context, cacheKey string) error
	AcquireRefreshLock(ctx context.Context, cacheKey string, ttl time.Duration) (bool, error)
	ReleaseRefreshLock(ctx context.Context, cacheKey string) error
}

// --- Token provider methods ---

func NewClaudeTokenProvider(_ AccountRepository, _ GeminiTokenCache, _ *OAuthService) *ClaudeTokenProvider {
	return &ClaudeTokenProvider{}
}
func (p *ClaudeTokenProvider) SetRefreshAPI(_ *OAuthRefreshAPI, _ interface{}) {}
func (p *ClaudeTokenProvider) SetRefreshPolicy(_ interface{})                  {}

func NewOpenAITokenProvider(_ AccountRepository, _ GeminiTokenCache, _ *OpenAIOAuthService) *OpenAITokenProvider {
	return &OpenAITokenProvider{}
}
func (p *OpenAITokenProvider) SetRefreshAPI(_ *OAuthRefreshAPI, _ interface{}) {}
func (p *OpenAITokenProvider) SetRefreshPolicy(_ interface{})                  {}

func NewAntigravityTokenProvider(_ AccountRepository, _ GeminiTokenCache, _ *AntigravityOAuthService) *AntigravityTokenProvider {
	return &AntigravityTokenProvider{}
}
func (p *AntigravityTokenProvider) SetRefreshAPI(_ *OAuthRefreshAPI, _ interface{}) {}
func (p *AntigravityTokenProvider) SetRefreshPolicy(_ interface{})                  {}
func (p *AntigravityTokenProvider) SetTempUnschedCache(_ TempUnschedCache)          {}

func NewGeminiTokenProvider(_ AccountRepository, _ GeminiTokenCache, _ *GeminiOAuthService) *GeminiTokenProvider {
	return &GeminiTokenProvider{}
}
func (p *GeminiTokenProvider) SetRefreshAPI(_ *OAuthRefreshAPI, _ interface{}) {}
func (p *GeminiTokenProvider) SetRefreshPolicy(_ interface{})                  {}

// --- Token refresher stubs ---

func NewClaudeTokenRefresher(_ *OAuthService) interface{}                            { return nil }
func NewOpenAITokenRefresher(_ *OpenAIOAuthService, _ AccountRepository) interface{} { return nil }
func NewGeminiTokenRefresher(_ *GeminiOAuthService) interface{}                      { return nil }
func NewAntigravityTokenRefresher(_ *AntigravityOAuthService) interface{}             { return nil }

// --- OpenAIOAuthService methods ---

func (s *OpenAIOAuthService) Stop()                     {}
func (s *OpenAIOAuthService) SetPrivacyClientFactory(_ ...interface{}) {}

func (s *OpenAIOAuthService) RefreshAccountToken(_ context.Context, _ *Account) (*OAuthTokenInfo, error) {
	return nil, fmt.Errorf("openai oauth removed")
}
func (s *OpenAIOAuthService) BuildAccountCredentials(_ *OAuthTokenInfo) map[string]interface{} {
	return nil
}

// --- AntigravityOAuthService methods ---

func (s *AntigravityOAuthService) Stop() {}

func (s *AntigravityOAuthService) RefreshAccountToken(_ context.Context, _ *Account) (*OAuthTokenInfo, error) {
	return nil, fmt.Errorf("antigravity oauth removed")
}
func (s *AntigravityOAuthService) BuildAccountCredentials(_ *OAuthTokenInfo) map[string]interface{} {
	return nil
}

// --- Sora stubs ---

func DefaultSoraModels(_ interface{}) []interface{} { return nil }

// --- Error code stubs ---

const errorCodeUnauthenticated = "UNAUTHENTICATED"

// OAuth client interfaces
type ClaudeOAuthClient interface{}
type GeminiOAuthClient interface{}
type OpenAIOAuthClient interface{}

// Sora types
type SoraAccount struct {
	AccountID    int64
	AccessToken  string
	RefreshToken string
	SessionToken string
}
type SoraGeneration struct {
	ID             string
	MediaURLs      []string
	S3ObjectKeys   []string
	UserID         int64
	APIKeyID       *int64
	Model          string
	Prompt         string
	MediaType      string
	Status         string
	MediaURL       string
	FileSizeBytes  int64
	StorageType    string
	UpstreamTaskID string
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	AccountID      int64
	ImageSize      string
	Duration       float64
	CompletedAt    *time.Time
}
type SoraGenerationRepository interface{}
type SoraGenerationListParams struct {
	UserID      int64
	Status      string
	StorageType string
	MediaType   string
	Limit       int
	Offset      int
	Before      *time.Time
	After       *time.Time
	Page        int
	PageSize    int
}

const SoraGenStatusPending = "pending"
const SoraGenStatusGenerating = "generating"
const SoraGenStatusCompleted = "completed"

var ErrSoraGenerationConcurrencyLimit = fmt.Errorf("sora generation concurrency limit")

const SoraGenStatusFailed = "failed"
const SoraGenStatusCancelled = "cancelled"

var ErrSoraStorageQuotaExceeded = fmt.Errorf("sora storage quota exceeded")

type ExchangeCodeInput struct {
	Code        string
	AccountType string
	RedirectURI string
	SessionID   string
	ProxyID     *int64
}

type CookieAuthInput struct {
	SessionKey  string
	AccountType string
	ProxyID     *int64
	Scope       string
}

func (s *GeminiOAuthService) RefreshAccountGoogleOneTier(_ context.Context, _ *Account) (string, map[string]interface{}, map[string]interface{}, error) {
	return "", nil, nil, fmt.Errorf("gemini oauth removed")
}
type SoraS3Storage struct{}

func (s *SoraS3Storage) TestConnectionWithSettings(_ ...interface{}) error { return nil }

func NewSoraS3Storage() *SoraS3Storage { return &SoraS3Storage{} }

// --- Additional missing functions ---

var sensitiveQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:key|client_secret|access_token|refresh_token)=)[^&"\s]+`)

func sanitizeUpstreamErrorMessage(msg string) string {
	if msg == "" {
		return msg
	}
	return sensitiveQueryParamRegex.ReplaceAllString(msg, `$1***`)
}

func normalizeCodexModel(model string) string { return model }

func shortHash(data []byte) string {
	if len(data) < 8 {
		return ""
	}
	return fmt.Sprintf("%x", data[:8])
}

func tempUnscheduleGoogleConfigError(_ ...interface{}) {}

// OAuthTokenInfo stub for token refresh results
type OAuthTokenInfo struct {
	AccessToken     string
	TokenType       string
	ExpiresIn       int64
	ExpiresAt       int64
	RefreshToken    string
	Scope           string
	ProjectIDMissing bool
}

// OAuthService methods
func (s *OAuthService) RefreshAccountToken(_ context.Context, _ *Account) (*OAuthTokenInfo, error) {
	return nil, fmt.Errorf("oauth service removed")
}
func (s *OAuthService) BuildAccountCredentials(_ *OAuthTokenInfo) map[string]interface{} {
	return nil
}
func (s *OAuthService) Stop() {}

func (s *OAuthService) GenerateAuthURL(_ ...interface{}) (string, error) {
	return "", fmt.Errorf("oauth service removed")
}
func (s *OAuthService) GenerateSetupTokenURL(_ ...interface{}) (string, error) {
	return "", fmt.Errorf("oauth service removed")
}
func (s *OAuthService) ExchangeCode(_ ...interface{}) (*OAuthTokenInfo, error) {
	return nil, fmt.Errorf("oauth service removed")
}
func (s *OAuthService) CookieAuth(_ ...interface{}) (*OAuthTokenInfo, error) {
	return nil, fmt.Errorf("oauth service removed")
}

// GeminiOAuthService methods
func (s *GeminiOAuthService) Stop() {}

func (s *GeminiOAuthService) RefreshAccountToken(_ context.Context, _ *Account) (*OAuthTokenInfo, error) {
	return nil, fmt.Errorf("gemini oauth service removed")
}
func (s *GeminiOAuthService) BuildAccountCredentials(_ *OAuthTokenInfo) map[string]interface{} {
	return nil
}

func tempUnscheduleEmptyResponse(_ ...interface{}) {}

// Account method from deleted antigravity_quota_scope.go
func (a *Account) IsSchedulableForModelWithContext(_ context.Context, _ string) bool {
	return true
}

// --- More stubs for gateway_service.go ---

func mapAntigravityModel(_ ...interface{}) string    { return "" }
func applyThinkingModelSuffix(_ ...interface{}) string { return "" }

// SoraModelConfig stub
type SoraModelConfig struct {
	Model string
	Type  string
}

func GetSoraModelConfig(_ string) (*SoraModelConfig, bool) { return nil, false }

// ClaudeTokenProvider methods
func (p *ClaudeTokenProvider) GetAccessToken(_ context.Context, _ *Account) (string, error) {
	return "", fmt.Errorf("claude token provider removed")
}

// Bedrock stubs
func PrepareBedrockRequestBodyWithTokens(_ ...interface{}) ([]byte, error) { return nil, nil }
func (s *GatewayService) handleBedrockStreamingResponse(_ ...interface{}) (*streamingResult, error) {
	return nil, fmt.Errorf("bedrock removed")
}
func transformBedrockInvocationMetrics(_ []byte) []byte    { return nil }
func ResolveBedrockBetaTokens(_ ...interface{}) []string   { return nil }

// Gemini tier constants
const GeminiTierAIStudioFree = "free"
const GeminiTierAIStudioPaid = "paid"
const GeminiTierGoogleOneFree = "google_one_free"
const GeminiTierGoogleAIPro = "google_ai_pro"
const GeminiTierGoogleAIUltra = "google_ai_ultra"
const GeminiTierGCPStandard = "gcp_standard"
const GeminiTierGCPEnterprise = "gcp_enterprise"

func canonicalGeminiTierID(_ string) string { return "" }
func canonicalGeminiTierIDForOAuthType(_ ...interface{}) string { return "" }

// BetaBlockedError stub
type BetaBlockedError struct {
	Message string
}

func (e *BetaBlockedError) Error() string { return e.Message }

// PromptTooLongError stub
type PromptTooLongError struct {
	StatusCode int
	RequestID  string
	Body       []byte
}

func (e *PromptTooLongError) Error() string { return "prompt too long" }

// OpenAI compat fallback metrics
type OpenAICompatFallbackMetrics struct {
	SessionHashLegacyReadFallbackTotal int64
	SessionHashLegacyReadFallbackHit   int64
	SessionHashLegacyDualWriteTotal    int64
	SessionHashLegacyReadHitRate       float64
	MetadataLegacyFallbackTotal        int64
}

func SnapshotOpenAICompatibilityFallbackMetrics() *OpenAICompatFallbackMetrics {
	return &OpenAICompatFallbackMetrics{}
}

// OpenAI parsed request body context key
const OpenAIParsedRequestBodyKey = "openai_parsed_request_body"

const GeminiTierGoogleOneUnknown = "google_one_unknown"

// SoraModelConfig fields
func (c *SoraModelConfig) GetType() string { return "" }

// OpenAIGatewayService methods
func (s *OpenAIGatewayService) SelectAccountWithLoadAwareness(_ ...interface{}) (*AccountSelectionResult, error) {
	return nil, fmt.Errorf("openai gateway removed")
}
func (s *OpenAIGatewayService) Forward(_ ...interface{}) (*ForwardResult, error) {
	return nil, fmt.Errorf("openai gateway removed")
}
func (s *OpenAIGatewayService) RecordUsage(_ ...interface{}) error { return nil }

// AntigravityGatewayService more methods
func (s *AntigravityGatewayService) Forward(_ ...interface{}) (*ForwardResult, error) {
	return nil, fmt.Errorf("antigravity gateway removed")
}
func (s *AntigravityGatewayService) ForwardGemini(_ ...interface{}) (*ForwardResult, error) {
	return nil, fmt.Errorf("antigravity gateway removed")
}

// GeminiMessagesCompatService methods
func (s *GeminiMessagesCompatService) Forward(_ ...interface{}) (*ForwardResult, error) {
	return nil, fmt.Errorf("gemini compat removed")
}
func (s *GeminiMessagesCompatService) ForwardNative(_ ...interface{}) (*ForwardResult, error) {
	return nil, fmt.Errorf("gemini compat removed")
}
func (s *GeminiMessagesCompatService) GetTokenProvider() *GeminiTokenProvider { return nil }

func SetOpenAIClientTransport(_ ...interface{}) {}

const OpenAIClientTransportHTTP = "http"

func ParseGeminiRateLimitResetTime(_ []byte) *int64 { return nil }

var errRefreshSkipped = fmt.Errorf("refresh skipped")

func GeminiTokenCacheKey(_ *Account) string       { return "" }
func AntigravityTokenCacheKey(_ *Account) string   { return "" }

func NewTokenRefreshService(_ ...interface{}) *TokenRefreshService { return &TokenRefreshService{} }

func NewSoraMediaStorage(_ ...interface{}) *SoraMediaStorage               { return &SoraMediaStorage{} }
func NewSoraSDKClient(_ ...interface{}) *SoraSDKClient                     { return &SoraSDKClient{} }
func NewSoraMediaCleanupService(_ ...interface{}) *SoraMediaCleanupService { return &SoraMediaCleanupService{} }

// Account method stub from deleted antigravity_quota_scope.go
func (a *Account) GetRateLimitRemainingTimeWithContext(_ context.Context, _ string) time.Duration {
	return 0
}
