package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
)

func TestFormatPEMKey_BareBase64(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	derBytes := x509.MarshalPKCS1PrivateKey(key)
	bare := base64.StdEncoding.EncodeToString(derBytes)

	result := formatPEMKey(bare, "RSA PRIVATE KEY")

	if !strings.HasPrefix(result, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Error("should have PEM header")
	}
	if !strings.HasSuffix(strings.TrimSpace(result), "-----END RSA PRIVATE KEY-----") {
		t.Error("should have PEM footer")
	}

	block, _ := pem.Decode([]byte(result))
	if block == nil {
		t.Fatal("pem.Decode returned nil")
	}
	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse key from formatted PEM: %v", err)
	}
}

func TestFormatPEMKey_AlreadyPEM(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	derBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: derBytes})
	pemStr := string(pemBlock)

	result := formatPEMKey(pemStr, "RSA PRIVATE KEY")
	// Should not modify already-PEM content
	if !strings.HasPrefix(result, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Error("should preserve PEM header")
	}
}

func TestBuildAlipaySignString_SortAndFilter(t *testing.T) {
	params := map[string]string{
		"b_key":     "b_value",
		"a_key":     "a_value",
		"sign":      "should_be_excluded",
		"sign_type": "RSA2",
		"c_key":     "c_value",
		"empty":     "",
	}

	// Notify sign string: excludes both sign and sign_type
	notifyResult := buildAlipayNotifySignString(params)
	expectedNotify := "a_key=a_value&b_key=b_value&c_key=c_value"
	if notifyResult != expectedNotify {
		t.Errorf("notify sign string: got %q, want %q", notifyResult, expectedNotify)
	}

	// Request sign string: excludes sign but keeps sign_type
	requestResult := buildAlipayRequestSignString(params)
	expectedRequest := "a_key=a_value&b_key=b_value&c_key=c_value&sign_type=RSA2"
	if requestResult != expectedRequest {
		t.Errorf("request sign string: got %q, want %q", requestResult, expectedRequest)
	}
}

func TestAlipaySignAndVerifyRoundTrip(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubKey := &privKey.PublicKey

	data := "a_key=a_value&b_key=b_value"

	// Sign
	hashed := sha256.Sum256([]byte(data))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		t.Fatal(err)
	}
	sig := base64.StdEncoding.EncodeToString(sigBytes)

	// Verify using the helper from alipay.go
	provider := &AlipayProvider{publicKey: pubKey}
	if err := provider.verifySignature(data, sig); err != nil {
		t.Errorf("verification failed: %v", err)
	}

	// Tampered data should fail
	if err := provider.verifySignature(data+"x", sig); err == nil {
		t.Error("should fail with tampered data")
	}
}

func TestMapAlipayTradeStatus(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TRADE_SUCCESS", "paid"},
		{"TRADE_FINISHED", "paid"},
		{"TRADE_CLOSED", "failed"},
		{"WAIT_BUYER_PAY", "pending"},
		{"UNKNOWN", "pending"},
	}
	for _, tt := range tests {
		got := mapAlipayTradeStatus(tt.input)
		if got != tt.want {
			t.Errorf("mapAlipayTradeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
