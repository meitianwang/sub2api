package payment

import (
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"encoding/base64"

	"github.com/shopspring/decimal"
)

func TestAmountToFen(t *testing.T) {
	tests := []struct {
		cny string
		fen int64
	}{
		{"1.00", 100},
		{"0.01", 1},
		{"99.99", 9999},
		{"100.50", 10050},
	}
	for _, tt := range tests {
		d, _ := decimal.NewFromString(tt.cny)
		got := amountToFen(d)
		if got != tt.fen {
			t.Errorf("amountToFen(%s) = %d, want %d", tt.cny, got, tt.fen)
		}
	}
}

func TestFenToAmount(t *testing.T) {
	tests := []struct {
		fen  int64
		want string
	}{
		{100, "1"},
		{1, "0.01"},
		{9999, "99.99"},
		{10050, "100.5"},
	}
	for _, tt := range tests {
		got := fenToAmount(tt.fen)
		want, _ := decimal.NewFromString(tt.want)
		if !got.Equal(want) {
			t.Errorf("fenToAmount(%d) = %s, want %s", tt.fen, got, tt.want)
		}
	}
}

func TestDecryptAESGCM(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes
	plaintext := []byte(`{"out_trade_no":"123","trade_state":"SUCCESS"}`)
	nonceStr := "test-nonce12"
	aad := "transaction"

	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := aesGCM.Seal(nil, []byte(nonceStr), plaintext, []byte(aad))
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Decrypt
	result, err := decryptAESGCM(key, nonceStr, encoded, aad)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(result) != string(plaintext) {
		t.Errorf("got %q, want %q", result, plaintext)
	}

	// Wrong key
	wrongKey := []byte("00000000000000000000000000000000")
	_, err = decryptAESGCM(wrongKey, nonceStr, encoded, aad)
	if err == nil {
		t.Error("should fail with wrong key")
	}

	// Wrong AAD
	_, err = decryptAESGCM(key, nonceStr, encoded, "wrong-aad")
	if err == nil {
		t.Error("should fail with wrong AAD")
	}
}

func TestMapWxTradeState(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"SUCCESS", "paid"},
		{"REFUND", "refunded"},
		{"CLOSED", "failed"},
		{"PAYERROR", "failed"},
		{"NOTPAY", "pending"},
		{"UNKNOWN", "pending"},
	}
	for _, tt := range tests {
		got := mapWxTradeState(tt.input)
		if got != tt.want {
			t.Errorf("mapWxTradeState(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetHeaderCaseInsensitive(t *testing.T) {
	headers := map[string]string{
		"Wechatpay-Timestamp": "1234567890",
		"Content-Type":        "application/json",
	}
	if v := getHeaderCaseInsensitive(headers, "wechatpay-timestamp"); v != "1234567890" {
		t.Errorf("got %q, want %q", v, "1234567890")
	}
	if v := getHeaderCaseInsensitive(headers, "content-type"); v != "application/json" {
		t.Errorf("got %q, want %q", v, "application/json")
	}
	if v := getHeaderCaseInsensitive(headers, "missing"); v != "" {
		t.Errorf("got %q, want empty", v)
	}
}
