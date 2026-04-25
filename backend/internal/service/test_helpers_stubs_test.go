package service

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

func ptrTime(t time.Time) *time.Time { return &t }

// mockAccountRepoForGemini is a mock AccountRepository used by many test files.
type mockAccountRepoForGemini struct {
	accounts     []Account
	accountsByID map[int64]*Account
}

func (m *mockAccountRepoForGemini) GetByID(_ context.Context, id int64) (*Account, error) {
	if acc, ok := m.accountsByID[id]; ok {
		return acc, nil
	}
	return nil, errors.New("account not found")
}
func (m *mockAccountRepoForGemini) GetByIDs(_ context.Context, ids []int64) ([]*Account, error) {
	var result []*Account
	for _, id := range ids {
		if acc, ok := m.accountsByID[id]; ok {
			result = append(result, acc)
		}
	}
	return result, nil
}
func (m *mockAccountRepoForGemini) ExistsByID(_ context.Context, id int64) (bool, error) {
	if m.accountsByID == nil {
		return false, nil
	}
	_, ok := m.accountsByID[id]
	return ok, nil
}
func (m *mockAccountRepoForGemini) ListSchedulableByPlatform(_ context.Context, _ string) ([]Account, error) {
	var result []Account
	for _, acc := range m.accounts {
		if acc.IsSchedulable() {
			result = append(result, acc)
		}
	}
	return result, nil
}
func (m *mockAccountRepoForGemini) ListSchedulableByGroupIDAndPlatform(_ context.Context, _ int64, platform string) ([]Account, error) {
	return m.ListSchedulableByPlatform(context.Background(), platform)
}
func (m *mockAccountRepoForGemini) Create(_ context.Context, _ *Account) error { return nil }
func (m *mockAccountRepoForGemini) GetByCRSAccountID(_ context.Context, _ string) (*Account, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) FindByExtraField(_ context.Context, _ string, _ any) ([]Account, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) ListCRSAccountIDs(_ context.Context) (map[string]int64, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) Update(_ context.Context, _ *Account) error { return nil }
func (m *mockAccountRepoForGemini) Delete(_ context.Context, _ int64) error     { return nil }
func (m *mockAccountRepoForGemini) List(_ context.Context, _ pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (m *mockAccountRepoForGemini) ListWithFilters(_ context.Context, _ pagination.PaginationParams, _, _, _, _ string, _ int64, _ string) ([]Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (m *mockAccountRepoForGemini) ListByGroup(_ context.Context, _ int64) ([]Account, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) ListActive(_ context.Context) ([]Account, error)            { return nil, nil }
func (m *mockAccountRepoForGemini) ListByPlatform(_ context.Context, _ string) ([]Account, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) UpdateLastUsed(_ context.Context, _ int64) error { return nil }
func (m *mockAccountRepoForGemini) BatchUpdateLastUsed(_ context.Context, _ map[int64]time.Time) error {
	return nil
}
func (m *mockAccountRepoForGemini) SetError(_ context.Context, _ int64, _ string) error   { return nil }
func (m *mockAccountRepoForGemini) ClearError(_ context.Context, _ int64) error            { return nil }
func (m *mockAccountRepoForGemini) SetSchedulable(_ context.Context, _ int64, _ bool) error { return nil }
func (m *mockAccountRepoForGemini) AutoPauseExpiredAccounts(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAccountRepoForGemini) BindGroups(_ context.Context, _ int64, _ []int64) error { return nil }
func (m *mockAccountRepoForGemini) ListSchedulable(_ context.Context) ([]Account, error)   { return nil, nil }
func (m *mockAccountRepoForGemini) ListSchedulableByGroupID(_ context.Context, _ int64) ([]Account, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) ListSchedulableByPlatforms(_ context.Context, _ []string) ([]Account, error) {
	var result []Account
	for _, acc := range m.accounts {
		if acc.IsSchedulable() {
			result = append(result, acc)
		}
	}
	return result, nil
}
func (m *mockAccountRepoForGemini) ListSchedulableByGroupIDAndPlatforms(_ context.Context, _ int64, platforms []string) ([]Account, error) {
	return m.ListSchedulableByPlatforms(context.Background(), platforms)
}
func (m *mockAccountRepoForGemini) ListSchedulableUngrouped(_ context.Context) ([]Account, error) {
	return nil, nil
}
func (m *mockAccountRepoForGemini) ListSchedulableUngroupedByPlatform(_ context.Context, platform string) ([]Account, error) {
	return m.ListSchedulableByPlatform(context.Background(), platform)
}
func (m *mockAccountRepoForGemini) ListSchedulableUngroupedByPlatforms(_ context.Context, platforms []string) ([]Account, error) {
	return m.ListSchedulableByPlatforms(context.Background(), platforms)
}
func (m *mockAccountRepoForGemini) SetRateLimited(_ context.Context, _ int64, _ time.Time) error {
	return nil
}
func (m *mockAccountRepoForGemini) SetModelRateLimit(_ context.Context, _ int64, _ string, _ time.Time) error {
	return nil
}
func (m *mockAccountRepoForGemini) SetOverloaded(_ context.Context, _ int64, _ time.Time) error { return nil }
func (m *mockAccountRepoForGemini) SetTempUnschedulable(_ context.Context, _ int64, _ time.Time, _ string) error {
	return nil
}
func (m *mockAccountRepoForGemini) ClearTempUnschedulable(_ context.Context, _ int64) error { return nil }
func (m *mockAccountRepoForGemini) ClearRateLimit(_ context.Context, _ int64) error         { return nil }

func (m *mockAccountRepoForGemini) ClearModelRateLimits(_ context.Context, _ int64) error { return nil }
func (m *mockAccountRepoForGemini) UpdateSessionWindow(_ context.Context, _ int64, _, _ *time.Time, _ string) error {
	return nil
}
func (m *mockAccountRepoForGemini) UpdateExtra(_ context.Context, _ int64, _ map[string]any) error {
	return nil
}
func (m *mockAccountRepoForGemini) BulkUpdate(_ context.Context, _ []int64, _ AccountBulkUpdate) (int64, error) {
	return 0, nil
}
func (m *mockAccountRepoForGemini) IncrementQuotaUsed(_ context.Context, _ int64, _ float64) error {
	return nil
}
func (m *mockAccountRepoForGemini) ResetQuotaUsed(_ context.Context, _ int64) error { return nil }

var _ AccountRepository = (*mockAccountRepoForGemini)(nil)

// splitChain splits a digest chain string by '-' delimiter (test helper).
func splitChain(chain string) []string {
	if chain == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i < len(chain); i++ {
		if chain[i] == '-' {
			parts = append(parts, chain[start:i])
			start = i + 1
		}
	}
	if start < len(chain) {
		parts = append(parts, chain[start:])
	}
	return parts
}

// stubOpenAIAccountRepo is a no-op implementation for tests.
type stubOpenAIAccountRepo struct{}

// GetOpenAIClientTransport stub for tests
func GetOpenAIClientTransport(_ ...interface{}) string { return "http" }

// logSink stub for captureStructuredLog
type logSink struct {
	Entries []map[string]interface{}
}

func (s *logSink) Has(key, value string) bool                     { return false }
func (s *logSink) ContainsMessageAtLevel(level, msg string) bool { return false }

// captureStructuredLog stub for tests - captures log output
func captureStructuredLog(_ interface{}) (*logSink, func()) {
	return &logSink{}, func() {}
}

// resetGatewayHotpathStatsForTest stub
func resetGatewayHotpathStatsForTest() {}
