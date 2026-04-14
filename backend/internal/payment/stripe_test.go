package payment

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v82"
)

func TestMapPaymentIntentStatus(t *testing.T) {
	tests := []struct {
		input stripe.PaymentIntentStatus
		want  string
	}{
		{stripe.PaymentIntentStatusSucceeded, "paid"},
		{stripe.PaymentIntentStatusCanceled, "failed"},
		{stripe.PaymentIntentStatusProcessing, "pending"},
		{stripe.PaymentIntentStatusRequiresAction, "pending"},
		{stripe.PaymentIntentStatusRequiresCapture, "pending"},
		{stripe.PaymentIntentStatusRequiresConfirmation, "pending"},
		{stripe.PaymentIntentStatusRequiresPaymentMethod, "pending"},
	}
	for _, tt := range tests {
		got := mapPaymentIntentStatus(tt.input)
		if got != tt.want {
			t.Errorf("mapPaymentIntentStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapRefundStatus(t *testing.T) {
	tests := []struct {
		input stripe.RefundStatus
		want  string
	}{
		{stripe.RefundStatusSucceeded, "success"},
		{stripe.RefundStatusFailed, "failed"},
		{stripe.RefundStatusPending, "pending"},
		{stripe.RefundStatusRequiresAction, "pending"},
	}
	for _, tt := range tests {
		got := mapRefundStatus(tt.input)
		if got != tt.want {
			t.Errorf("mapRefundStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToSmallestUnit(t *testing.T) {
	tests := []struct {
		amount   string
		currency string
		want     int64
	}{
		{"1.00", "cny", 100},
		{"0.01", "cny", 1},
		{"99.99", "cny", 9999},
		{"100.50", "usd", 10050},
		{"500", "jpy", 500},   // zero-decimal currency
		{"1000", "krw", 1000}, // zero-decimal currency
	}
	for _, tt := range tests {
		d, _ := decimal.NewFromString(tt.amount)
		got := toSmallestUnit(d, tt.currency)
		if got != tt.want {
			t.Errorf("toSmallestUnit(%s, %s) = %d, want %d", tt.amount, tt.currency, got, tt.want)
		}
	}
}

func TestFromSmallestUnit(t *testing.T) {
	tests := []struct {
		units    int64
		currency string
		want     string
	}{
		{100, "cny", "1"},
		{1, "cny", "0.01"},
		{9999, "cny", "99.99"},
		{10050, "usd", "100.5"},
		{500, "jpy", "500"},   // zero-decimal currency
		{1000, "krw", "1000"}, // zero-decimal currency
	}
	for _, tt := range tests {
		got := fromSmallestUnit(tt.units, tt.currency)
		want, _ := decimal.NewFromString(tt.want)
		if !got.Equal(want) {
			t.Errorf("fromSmallestUnit(%d, %s) = %s, want %s", tt.units, tt.currency, got, tt.want)
		}
	}
}
