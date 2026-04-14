package payment

import (
	"testing"
)

func TestEasyPaySign(t *testing.T) {
	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "order-001",
		"notify_url":   "https://example.com/notify",
		"return_url":   "https://example.com/return",
		"name":         "Test Product",
		"money":        "10.00",
		"clientip":     "127.0.0.1",
	}
	pkey := "test-merchant-secret-key"

	sig1 := easyPaySign(params, pkey)
	sig2 := easyPaySign(params, pkey)

	// Deterministic
	if sig1 != sig2 {
		t.Errorf("sign not deterministic: %s != %s", sig1, sig2)
	}

	// 32-char hex
	if len(sig1) != 32 {
		t.Errorf("expected 32-char hex, got %d chars", len(sig1))
	}

	// Verify
	if !easyPayVerifySign(params, pkey, sig1) {
		t.Error("verification failed for correct signature")
	}

	// Wrong pkey
	if easyPayVerifySign(params, "wrong-key", sig1) {
		t.Error("verification should fail with wrong key")
	}

	// Empty values and sign/sign_type should be excluded
	paramsWithExtra := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "order-001",
		"notify_url":   "https://example.com/notify",
		"return_url":   "https://example.com/return",
		"name":         "Test Product",
		"money":        "10.00",
		"clientip":     "127.0.0.1",
		"sign":         "should-be-excluded",
		"sign_type":    "MD5",
		"empty_field":  "",
	}
	sigExtra := easyPaySign(paramsWithExtra, pkey)
	if sigExtra != sig1 {
		t.Errorf("sign should ignore sign/sign_type/empty fields: %s != %s", sigExtra, sig1)
	}
}

func TestEasyPayVerifySign_ConstantTime(t *testing.T) {
	params := map[string]string{"key": "value"}
	pkey := "secret"
	sig := easyPaySign(params, pkey)

	// Wrong length should fail
	if easyPayVerifySign(params, pkey, sig+"x") {
		t.Error("should fail with wrong length")
	}

	// Tampered signature
	tampered := "0" + sig[1:]
	if tampered == sig {
		tampered = "1" + sig[1:]
	}
	if easyPayVerifySign(params, pkey, tampered) {
		t.Error("should fail with tampered signature")
	}
}
