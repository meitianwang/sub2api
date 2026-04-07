package service

import "strings"

// ModelPricingEntry 单个模型的定价配置
type ModelPricingEntry struct {
	SellInputPrice  float64 `json:"sell_input_price"`  // 卖给用户的输入价格 (USD/M tokens)
	SellOutputPrice float64 `json:"sell_output_price"` // 卖给用户的输出价格 (USD/M tokens)
	CostInputPrice  float64 `json:"cost_input_price"`  // 上游进货输入价格 (USD/M tokens)
	CostOutputPrice float64 `json:"cost_output_price"` // 上游进货输出价格 (USD/M tokens)
}

// ModelPricingMap 分组的模型定价映射
type ModelPricingMap map[string]ModelPricingEntry

// GetPricing 获取指定模型的定价，支持精确匹配（大小写不敏感）
// 返回 nil 表示该模型未配置定价
func (m ModelPricingMap) GetPricing(model string) *ModelPricingEntry {
	if m == nil || len(m) == 0 {
		return nil
	}
	model = strings.ToLower(model)
	if entry, ok := m[model]; ok {
		return &entry
	}
	return nil
}

// ParseModelPricingMap 从 map[string]map[string]float64 解析为 ModelPricingMap
// 用于从 Ent 的 JSON 字段转换
func ParseModelPricingMap(raw map[string]map[string]float64) ModelPricingMap {
	if raw == nil || len(raw) == 0 {
		return nil
	}
	result := make(ModelPricingMap, len(raw))
	for model, prices := range raw {
		result[strings.ToLower(model)] = ModelPricingEntry{
			SellInputPrice:  prices["sell_input_price"],
			SellOutputPrice: prices["sell_output_price"],
			CostInputPrice:  prices["cost_input_price"],
			CostOutputPrice: prices["cost_output_price"],
		}
	}
	return result
}
