package converter

import (
	"testing"
	"time"
)

func fixedOpts(sourceName string) ConvertOptions {
	return ConvertOptions{
		SourceName: sourceName,
		Now:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestConvertCPARecord_Codex_HappyPath(t *testing.T) {
	accessToken := makeJWT(t, anyMap{
		"email": "user@example.com",
		"exp":   float64(2000000000),
		"https://api.openai.com/auth": anyMap{
			"chatgpt_account_id": "acc-123",
			"chatgpt_plan_type":  "pro",
			"chatgpt_user_id":    "user-1",
			"organizations": []interface{}{
				anyMap{"id": "org-1", "is_default": true},
			},
		},
	})
	idToken := makeJWT(t, anyMap{"email": "user@example.com"})

	record := anyMap{
		"type":          "codex",
		"access_token":  accessToken,
		"id_token":      idToken,
		"refresh_token": "refresh-xyz",
	}

	result, err := ConvertCPARecord(record, fixedOpts("codex-source.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourceType != "codex" {
		t.Errorf("SourceType = %q, want codex", result.SourceType)
	}
	if result.Email != "user@example.com" {
		t.Errorf("Email = %q", result.Email)
	}
	if result.Account["platform"] != "openai" {
		t.Errorf("platform = %v, want openai (保留原实现的特意写法)", result.Account["platform"])
	}
	creds, ok := result.Account["credentials"].(anyMap)
	if !ok {
		t.Fatalf("credentials missing or wrong type: %#v", result.Account["credentials"])
	}
	if creds["organization_id"] != "org-1" {
		t.Errorf("organization_id = %v, want org-1", creds["organization_id"])
	}
	if creds["refresh_token"] != "refresh-xyz" {
		t.Errorf("refresh_token = %v, want refresh-xyz", creds["refresh_token"])
	}
}

func TestConvertCPARecord_Codex_MissingAccessToken(t *testing.T) {
	record := anyMap{"type": "codex", "id_token": "x"}
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil || err.Error() != "缺少 access_token" {
		t.Fatalf("err = %v, want 缺少 access_token", err)
	}
}

func TestConvertCPARecord_Codex_MissingIDToken(t *testing.T) {
	accessToken := makeJWT(t, anyMap{})
	record := anyMap{"type": "codex", "access_token": accessToken}
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil || err.Error() != "缺少 id_token" {
		t.Fatalf("err = %v, want 缺少 id_token", err)
	}
}

func TestConvertCPARecord_Codex_InvalidJWT(t *testing.T) {
	record := anyMap{"type": "codex", "access_token": "not-a-jwt", "id_token": "also-not-a-jwt"}
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil {
		t.Fatal("expected error for invalid access_token JWT")
	}
}

func TestConvertCPARecord_Claude_HappyPath(t *testing.T) {
	record := anyMap{
		"type":          "claude",
		"access_token":  "at-1",
		"email":         "claude@example.com",
		"refresh_token": "rt-1",
		"expired":       "2030-01-01T00:00:00.000Z",
	}
	result, err := ConvertCPARecord(record, fixedOpts(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourceType != "claude" {
		t.Errorf("SourceType = %q", result.SourceType)
	}
	if result.Account["platform"] != "anthropic" {
		t.Errorf("platform = %v, want anthropic", result.Account["platform"])
	}
}

func TestConvertCPARecord_Claude_MissingAccessToken(t *testing.T) {
	record := anyMap{"type": "claude"}
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil || err.Error() != "缺少 access_token" {
		t.Fatalf("err = %v, want 缺少 access_token", err)
	}
}

func TestConvertCPARecord_Antigravity_DerivedExpiry(t *testing.T) {
	record := anyMap{
		"type":         "antigravity",
		"access_token": "at-ag",
		"timestamp":    float64(1700000000000),
		"expires_in":   float64(3600),
		"email":        "ag@example.com",
	}
	result, err := ConvertCPARecord(record, fixedOpts(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExpiresAt == "" {
		t.Error("expected derived expiresAt from timestamp+expires_in, got empty")
	}
}

func TestConvertCPARecord_Gemini_CamelCaseFallback(t *testing.T) {
	record := anyMap{
		"type": "gemini",
		"token": anyMap{
			"accessToken":  "at-gemini",
			"refreshToken": "rt-gemini",
			"tokenType":    "Bearer",
		},
		"project_id": "proj-1",
		"email":      "gem@example.com",
	}
	result, err := ConvertCPARecord(record, fixedOpts(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	creds := result.Account["credentials"].(anyMap)
	if creds["access_token"] != "at-gemini" {
		t.Errorf("access_token = %v", creds["access_token"])
	}
	if creds["oauth_type"] != "code_assist" {
		t.Errorf("oauth_type = %v, want code_assist", creds["oauth_type"])
	}
}

func TestConvertCPARecord_Gemini_MissingToken(t *testing.T) {
	record := anyMap{"type": "gemini"}
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil {
		t.Fatal("expected error for missing token object")
	}
}

func TestConvertCPARecord_DefaultsToCodex(t *testing.T) {
	record := anyMap{"access_token": "x"} // 无 type 字段，应默认按 codex 处理并因缺少 id_token 报错
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil || err.Error() != "缺少 id_token" {
		t.Fatalf("err = %v, want 缺少 id_token (默认按 codex 处理)", err)
	}
}

func TestConvertCPARecord_UnsupportedType(t *testing.T) {
	record := anyMap{"type": "unknown-thing"}
	_, err := ConvertCPARecord(record, fixedOpts(""))
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
