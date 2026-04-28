package jwt

import (
	"strings"
	"testing"
)

const testSecret = "abcdefghijklmnopqrstuvwxyz"

// ===================== Create =====================

func TestCreate_Success(t *testing.T) {
	claims := map[string]any{
		"login_user_key": "test-uuid-123",
		"user_id":        float64(1),
	}
	token, err := Create(testSecret, claims)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if token == "" {
		t.Fatal("Create returned empty token")
	}
	// JWT 由三段 base64 组成，用 . 分隔
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 token parts, got %d", len(parts))
	}
}

func TestCreate_EmptyClaims(t *testing.T) {
	token, err := Create(testSecret, map[string]any{})
	if err != nil {
		t.Fatalf("Create with empty claims should succeed, got: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty even with empty claims")
	}
}

func TestCreate_HS512Algorithm(t *testing.T) {
	// 项目要求使用 HS512（与 Java 兼容）
	// 通过验证 token header 包含 alg: HS512 来确认
	token, err := Create(testSecret, map[string]any{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	headerB64 := strings.Split(token, ".")[0]
	// HS512 的 base64 编码 header 通常以特定前缀开始
	// {"alg":"HS512","typ":"JWT"} → eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9
	if headerB64 != "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9" {
		t.Errorf("expected HS512 header, got header b64: %s", headerB64)
	}
}

// ===================== Parse =====================

func TestParse_Success(t *testing.T) {
	original := map[string]any{
		"login_user_key": "uuid-abc",
		"user_id":        float64(42),
	}
	token, err := Create(testSecret, original)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(testSecret, token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed["login_user_key"] != "uuid-abc" {
		t.Errorf("login_user_key mismatch: got %v", parsed["login_user_key"])
	}
	if parsed["user_id"] != float64(42) {
		t.Errorf("user_id mismatch: got %v", parsed["user_id"])
	}
}

func TestParse_WrongSecret(t *testing.T) {
	token, _ := Create(testSecret, map[string]any{"k": "v"})
	_, err := Parse("wrong-secret", token)
	if err == nil {
		t.Fatal("Parse should fail with wrong secret, but succeeded")
	}
}

func TestParse_MalformedToken(t *testing.T) {
	cases := []string{
		"",
		"not.a.jwt",
		"only-one-segment",
		"two.segments",
		"invalid-base64...invalid...invalid",
	}
	for _, c := range cases {
		_, err := Parse(testSecret, c)
		if err == nil {
			t.Errorf("expected error for malformed token %q, got nil", c)
		}
	}
}

func TestParse_TamperedPayload(t *testing.T) {
	token, _ := Create(testSecret, map[string]any{"user_id": float64(1)})
	parts := strings.Split(token, ".")
	// 篡改 payload（使签名不匹配）
	tampered := parts[0] + ".eyJ1c2VyX2lkIjo5OTl9." + parts[2]
	_, err := Parse(testSecret, tampered)
	if err == nil {
		t.Fatal("Parse should reject tampered token")
	}
}

func TestParse_RejectNoneAlgorithm(t *testing.T) {
	// 安全：必须拒绝 alg=none 攻击
	// {"alg":"none","typ":"JWT"} = eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoxfQ."
	_, err := Parse(testSecret, noneToken)
	if err == nil {
		t.Fatal("Parse must reject alg=none token (security risk)")
	}
}

// ===================== Round-trip 一致性 =====================

func TestRoundTrip_PreservesClaims(t *testing.T) {
	tests := []map[string]any{
		{"k1": "string-value"},
		{"k2": float64(123)},
		{"k3": true},
		{"a": "x", "b": float64(1), "c": false},
	}
	for i, c := range tests {
		token, err := Create(testSecret, c)
		if err != nil {
			t.Fatalf("[%d] Create failed: %v", i, err)
		}
		parsed, err := Parse(testSecret, token)
		if err != nil {
			t.Fatalf("[%d] Parse failed: %v", i, err)
		}
		for k, v := range c {
			if parsed[k] != v {
				t.Errorf("[%d] key %s: want %v, got %v", i, k, v, parsed[k])
			}
		}
	}
}
