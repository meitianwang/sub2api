package payment

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestDecodeAlipayBody_UTF8(t *testing.T) {
	raw := []byte("trade_status=TRADE_SUCCESS&total_amount=10.00")
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=utf-8"}
	result := decodeAlipayBody(raw, headers)
	if result != string(raw) {
		t.Errorf("got %q, want %q", result, string(raw))
	}
}

func TestDecodeAlipayBody_GBK(t *testing.T) {
	// Encode Chinese text "测试" in GBK
	gbkEncoder := simplifiedchinese.GBK.NewEncoder()
	gbkBytes, err := gbkEncoder.Bytes([]byte("name=测试&status=OK"))
	if err != nil {
		t.Fatal(err)
	}

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=gbk"}
	result := decodeAlipayBody(gbkBytes, headers)
	if result != "name=测试&status=OK" {
		t.Errorf("got %q, want %q", result, "name=测试&status=OK")
	}
}

func TestDecodeAlipayBody_FallbackFromBodyParam(t *testing.T) {
	// charset specified in body, not header
	gbkEncoder := simplifiedchinese.GBK.NewEncoder()
	gbkBytes, err := gbkEncoder.Bytes([]byte("charset=gbk&name=测试"))
	if err != nil {
		t.Fatal(err)
	}

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	result := decodeAlipayBody(gbkBytes, headers)
	if result != "charset=gbk&name=测试" {
		t.Errorf("got %q, want %q", result, "charset=gbk&name=测试")
	}
}

func TestDetectAlipayCharset_ContentType(t *testing.T) {
	tests := []struct {
		ct   string
		want string
	}{
		{"application/x-www-form-urlencoded; charset=gbk", "gbk"},
		{"application/x-www-form-urlencoded; charset=UTF-8", "utf-8"},
		{"application/x-www-form-urlencoded; charset=gb2312", "gbk"},
		{"text/html", "utf-8"},
		{"", "utf-8"},
	}
	for _, tt := range tests {
		headers := map[string]string{"Content-Type": tt.ct}
		got := detectAlipayCharset(nil, headers)
		if got != tt.want {
			t.Errorf("detectAlipayCharset(ct=%q) = %q, want %q", tt.ct, got, tt.want)
		}
	}
}

func TestNormalizeCharset(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"UTF-8", "utf-8"},
		{"utf8", "utf-8"},
		{"GBK", "gbk"},
		{"gb2312", "gbk"},
		{"gb_2312-80", "gbk"},
		{"GB18030", "gb18030"},
	}
	for _, tt := range tests {
		got := normalizeCharset(tt.input)
		if got != tt.want {
			t.Errorf("normalizeCharset(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"utf-8", "gbk", "utf-8", "gb18030", "gbk"}
	result := uniqueStrings(input)
	expected := []string{"utf-8", "gbk", "gb18030"}
	if len(result) != len(expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, result[i], expected[i])
		}
	}
}
