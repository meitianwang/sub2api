//go:build unit

package service

import "testing"

func TestModelPricingMap_GetPricing(t *testing.T) {
	tests := []struct {
		name     string
		m        ModelPricingMap
		model    string
		wantNil  bool
		wantSell float64
	}{
		{
			name:    "nil map returns nil",
			m:       nil,
			model:   "claude-sonnet-4-6",
			wantNil: true,
		},
		{
			name:    "empty map returns nil",
			m:       ModelPricingMap{},
			model:   "claude-sonnet-4-6",
			wantNil: true,
		},
		{
			name: "exact match",
			m: ModelPricingMap{
				"claude-sonnet-4-6": {SellInputPrice: 1.5, SellOutputPrice: 7.5},
			},
			model:    "claude-sonnet-4-6",
			wantNil:  false,
			wantSell: 1.5,
		},
		{
			name: "case insensitive match",
			m: ModelPricingMap{
				"claude-sonnet-4-6": {SellInputPrice: 1.5},
			},
			model:    "Claude-Sonnet-4-6",
			wantNil:  false,
			wantSell: 1.5,
		},
		{
			name: "model not found",
			m: ModelPricingMap{
				"claude-sonnet-4-6": {SellInputPrice: 1.5},
			},
			model:   "gpt-5.4",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.GetPricing(tt.model)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil, got nil")
			}
			if got.SellInputPrice != tt.wantSell {
				t.Errorf("SellInputPrice = %v, want %v", got.SellInputPrice, tt.wantSell)
			}
		})
	}
}

func TestParseModelPricingMap(t *testing.T) {
	tests := []struct {
		name    string
		raw     map[string]map[string]float64
		wantNil bool
		wantLen int
	}{
		{
			name:    "nil input",
			raw:     nil,
			wantNil: true,
		},
		{
			name:    "empty input",
			raw:     map[string]map[string]float64{},
			wantNil: true,
		},
		{
			name: "parses correctly and lowercases keys",
			raw: map[string]map[string]float64{
				"Claude-Sonnet-4-6": {
					"sell_input_price":  1.5,
					"sell_output_price": 7.5,
					"cost_input_price":  0.5,
					"cost_output_price": 3.0,
				},
			},
			wantNil: false,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseModelPricingMap(tt.raw)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil")
			}
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
			// verify lowercase key
			if _, ok := got["claude-sonnet-4-6"]; !ok {
				t.Error("expected lowercase key 'claude-sonnet-4-6'")
			}
		})
	}
}
