package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name       string
		amount     string
		feeRate    string
		wantFee    string
		wantPay    string
	}{
		{"zero rate", "100.00", "0", "0", "100"},
		{"2.5% rate", "100.00", "2.5", "2.50", "102.50"},
		{"round up", "10.00", "3", "0.30", "10.30"},
		{"round up fractional", "33.33", "1.5", "0.50", "33.83"},
		{"large amount", "10000.00", "5", "500.00", "10500.00"},
		{"small amount small rate", "1.00", "0.1", "0.01", "1.01"},
		{"negative rate treated as zero", "100.00", "-1", "0", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, _ := decimal.NewFromString(tt.amount)
			rate, _ := decimal.NewFromString(tt.feeRate)
			wantFee, _ := decimal.NewFromString(tt.wantFee)
			wantPay, _ := decimal.NewFromString(tt.wantPay)

			gotFee, gotPay := CalculateFee(amount, rate)
			if !gotFee.Equal(wantFee) {
				t.Errorf("fee = %s, want %s", gotFee, wantFee)
			}
			if !gotPay.Equal(wantPay) {
				t.Errorf("payAmount = %s, want %s", gotPay, wantPay)
			}
		})
	}
}
