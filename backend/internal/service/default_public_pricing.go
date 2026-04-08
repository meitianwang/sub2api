package service

import "strings"

// DefaultPublicPrice 公开展示的默认模型定价（USD/M tokens）
type DefaultPublicPrice struct {
	InputPrice  float64 // USD per million input tokens
	OutputPrice float64 // USD per million output tokens
}

// defaultPublicPricing 默认公开定价表（来源：billing_service fallback pricing）
// 键为小写模型系列前缀，用于前缀匹配
var defaultPublicPricing = map[string]DefaultPublicPrice{
	// Claude
	"claude-opus-4-5":    {InputPrice: 5, OutputPrice: 25},
	"claude-opus-4-6":    {InputPrice: 5, OutputPrice: 25},
	"claude-sonnet-4-5":  {InputPrice: 3, OutputPrice: 15},
	"claude-sonnet-4-6":  {InputPrice: 3, OutputPrice: 15},

	// Gemini
	"gemini-2.5-flash":      {InputPrice: 0.15, OutputPrice: 0.60},
	"gemini-2.5-pro":        {InputPrice: 1.25, OutputPrice: 10},
	"gemini-3-flash":        {InputPrice: 0.15, OutputPrice: 0.60},
	"gemini-3-pro":          {InputPrice: 2, OutputPrice: 12},
	"gemini-3.1-pro":        {InputPrice: 2, OutputPrice: 12},
	"gemini-3.1-flash":      {InputPrice: 0.15, OutputPrice: 0.60},

	// OpenAI
	"gpt-5.4":      {InputPrice: 2.5, OutputPrice: 15},
	"gpt-5.4-mini": {InputPrice: 0.75, OutputPrice: 4.5},
	"gpt-5.4-nano": {InputPrice: 0.20, OutputPrice: 1.25},
	"gpt-5.3-codex": {InputPrice: 1.5, OutputPrice: 12},
	"gpt-5.2":       {InputPrice: 1.75, OutputPrice: 14},
	"gpt-5.2-codex": {InputPrice: 1.75, OutputPrice: 14},
	"gpt-5.1":       {InputPrice: 1.25, OutputPrice: 10},
	"gpt-5.1-codex": {InputPrice: 1.5, OutputPrice: 12},
	"gpt-5":         {InputPrice: 1.25, OutputPrice: 10},
}

// GetDefaultPublicPricing 获取模型的默认公开定价
// 优先精确匹配，然后前缀匹配
func GetDefaultPublicPricing(modelID string) *DefaultPublicPrice {
	modelLower := strings.ToLower(modelID)

	// 精确匹配
	if p, ok := defaultPublicPricing[modelLower]; ok {
		return &p
	}

	// 前缀匹配（从最长前缀开始）
	bestMatch := ""
	for prefix := range defaultPublicPricing {
		if strings.HasPrefix(modelLower, prefix) && len(prefix) > len(bestMatch) {
			bestMatch = prefix
		}
	}
	if bestMatch != "" {
		p := defaultPublicPricing[bestMatch]
		return &p
	}

	return nil
}
