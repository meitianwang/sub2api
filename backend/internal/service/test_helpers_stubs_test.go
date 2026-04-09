package service

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
