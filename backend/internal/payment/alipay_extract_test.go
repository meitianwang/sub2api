package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

// helper: create an AlipayProvider with a real key pair for sign/verify tests.
func newTestAlipayProvider(t *testing.T) (*AlipayProvider, *rsa.PrivateKey) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return &AlipayProvider{
		publicKey:  &privKey.PublicKey,
		privateKey: privKey,
	}, privKey
}

// signJSON signs a JSON substring the same way Alipay would.
func signJSON(t *testing.T, privKey *rsa.PrivateKey, data []byte) string {
	t.Helper()
	hashed := sha256.Sum256(data)
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}

func TestExtractAndVerifyResponse_Simple(t *testing.T) {
	p, privKey := newTestAlipayProvider(t)
	inner := `{"code":"10000","msg":"Success"}`
	sig := signJSON(t, privKey, []byte(inner))
	body := []byte(`{"alipay_trade_query_response":` + inner + `,"sign":"` + sig + `"}`)
	data, err := p.extractAndVerifyResponse(body, "alipay_trade_query_response")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}

func TestExtractAndVerifyResponse_NestedObjects(t *testing.T) {
	p, _ := newTestAlipayProvider(t)
	inner := `{"code":"10000","data":{"nested":{"deep":"value"}},"msg":"OK"}`
	body := []byte(`{"resp":` + inner + `}`)
	data, err := p.extractAndVerifyResponse(body, "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}

func TestExtractAndVerifyResponse_BracesInsideStrings(t *testing.T) {
	p, _ := newTestAlipayProvider(t)
	inner := `{"code":"10000","msg":"error {invalid} thing"}`
	body := []byte(`{"resp":` + inner + `}`)
	data, err := p.extractAndVerifyResponse(body, "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}

func TestExtractAndVerifyResponse_EscapedQuotesInsideStrings(t *testing.T) {
	p, _ := newTestAlipayProvider(t)
	// JSON with escaped quotes inside a string value that also contains braces
	inner := `{"code":"10000","msg":"say \"hello {world}\""}`
	body := []byte(`{"resp":` + inner + `}`)
	data, err := p.extractAndVerifyResponse(body, "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}

func TestExtractAndVerifyResponse_WithValidSignature(t *testing.T) {
	p, privKey := newTestAlipayProvider(t)
	inner := `{"code":"10000","trade_no":"2024010100001"}`
	sig := signJSON(t, privKey, []byte(inner))
	body := []byte(`{"alipay_trade_query_response":` + inner + `,"sign":"` + sig + `"}`)

	data, err := p.extractAndVerifyResponse(body, "alipay_trade_query_response")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}

func TestExtractAndVerifyResponse_WithInvalidSignature(t *testing.T) {
	p, privKey := newTestAlipayProvider(t)
	inner := `{"code":"10000","trade_no":"2024010100001"}`
	sig := signJSON(t, privKey, []byte(inner+"tampered"))
	body := []byte(`{"alipay_trade_query_response":` + inner + `,"sign":"` + sig + `"}`)

	_, err := p.extractAndVerifyResponse(body, "alipay_trade_query_response")
	if err == nil {
		t.Fatal("expected signature verification to fail")
	}
}

func TestExtractAndVerifyResponse_NoSign(t *testing.T) {
	p, _ := newTestAlipayProvider(t)
	inner := `{"code":"10000"}`
	body := []byte(`{"resp":` + inner + `}`)
	data, err := p.extractAndVerifyResponse(body, "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}

func TestExtractAndVerifyResponse_MissingKey(t *testing.T) {
	p, _ := newTestAlipayProvider(t)
	body := []byte(`{"other_response":{"code":"10000"}}`)
	_, err := p.extractAndVerifyResponse(body, "alipay_trade_query_response")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestExtractAndVerifyResponse_WhitespaceAroundColon(t *testing.T) {
	p, _ := newTestAlipayProvider(t)
	// Some JSON formatters add space after colon
	inner := `{"code":"10000"}`
	body := []byte(`{"resp": ` + inner + `}`)
	data, err := p.extractAndVerifyResponse(body, "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != inner {
		t.Errorf("got %q, want %q", data, inner)
	}
}
