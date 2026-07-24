package converter

import "testing"

func TestNormalizeSub2ApiPlatform(t *testing.T) {
	cases := map[string]string{
		"openai":      "codex",
		"Codex":       "codex",
		"anthropic":   "claude",
		"Claude":      "claude",
		"antigravity": "antigravity",
		"gemini":      "gemini",
		"unknown":     "",
		"":            "",
	}
	for input, want := range cases {
		if got := NormalizeSub2ApiPlatform(input); got != want {
			t.Errorf("NormalizeSub2ApiPlatform(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExtractSub2ApiAccounts_Array(t *testing.T) {
	doc := []interface{}{anyMap{"platform": "codex"}}
	accounts, err := ExtractSub2ApiAccounts(doc)
	if err != nil || len(accounts) != 1 {
		t.Fatalf("accounts = %v, err = %v", accounts, err)
	}
}

func TestExtractSub2ApiAccounts_AccountsField(t *testing.T) {
	doc := anyMap{"accounts": []interface{}{anyMap{"platform": "claude"}}}
	accounts, err := ExtractSub2ApiAccounts(doc)
	if err != nil || len(accounts) != 1 {
		t.Fatalf("accounts = %v, err = %v", accounts, err)
	}
}

func TestExtractSub2ApiAccounts_BareSingleAccount(t *testing.T) {
	doc := anyMap{"platform": "claude", "credentials": anyMap{"access_token": "x"}}
	accounts, err := ExtractSub2ApiAccounts(doc)
	if err != nil || len(accounts) != 1 {
		t.Fatalf("accounts = %v, err = %v", accounts, err)
	}
}

func TestExtractSub2ApiAccounts_Invalid(t *testing.T) {
	if _, err := ExtractSub2ApiAccounts(anyMap{"foo": "bar"}); err == nil {
		t.Fatal("expected error for invalid document")
	}
}

func TestConvertSub2ApiAccount_OpenAI_MissingIDToken(t *testing.T) {
	account := anyMap{
		"platform":    "openai",
		"credentials": anyMap{"access_token": "at"},
	}
	_, err := ConvertSub2ApiAccount(account, ConvertOptions{})
	if err == nil || err.Error() != "credentials.id_token 为空，无法生成 Codex CPA 文件" {
		t.Fatalf("err = %v", err)
	}
}

func TestConvertSub2ApiAccount_OpenAI_HappyPath(t *testing.T) {
	account := anyMap{
		"name":     "my-openai",
		"platform": "openai",
		"type":     "oauth",
		"credentials": anyMap{
			"access_token":       "at",
			"id_token":           "idt",
			"refresh_token":      "rt",
			"chatgpt_account_id": "acc-1",
		},
		"extra": anyMap{"email": "openai@example.com"},
	}
	result, err := ConvertSub2ApiAccount(account, ConvertOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Document["type"] != "codex" {
		t.Errorf("type = %v", result.Document["type"])
	}
	if result.Document["refresh_token"] != "rt" {
		t.Errorf("refresh_token = %v", result.Document["refresh_token"])
	}
	if result.OutputFileName == "" {
		t.Error("expected non-empty OutputFileName")
	}
}

func TestConvertSub2ApiAccount_OpenAI_RefreshTokenAlwaysPresent(t *testing.T) {
	account := anyMap{
		"platform": "openai",
		"credentials": anyMap{
			"access_token": "at",
			"id_token":     "idt",
			// 无 refresh_token
		},
	}
	result, err := ConvertSub2ApiAccount(account, ConvertOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, exists := result.Document["refresh_token"]
	if !exists {
		t.Fatal("refresh_token key should always be present for codex output, even when empty")
	}
	if v != "" {
		t.Errorf("refresh_token = %v, want empty string", v)
	}
}

func TestConvertSub2ApiAccount_Claude_MissingAccessToken(t *testing.T) {
	account := anyMap{"platform": "claude", "credentials": anyMap{}}
	if _, err := ConvertSub2ApiAccount(account, ConvertOptions{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestConvertSub2ApiAccount_Antigravity_HappyPath(t *testing.T) {
	account := anyMap{
		"platform": "antigravity",
		"credentials": anyMap{
			"access_token": "at",
			"expires_in":   float64(120),
			"project_id":   "proj",
		},
		"extra": anyMap{"email": "ag@example.com"},
	}
	result, err := ConvertSub2ApiAccount(account, ConvertOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Document["type"] != "antigravity" {
		t.Errorf("type = %v", result.Document["type"])
	}
}

func TestConvertSub2ApiAccount_Gemini_HappyPath(t *testing.T) {
	account := anyMap{
		"platform": "gemini",
		"credentials": anyMap{
			"access_token": "at",
			"project_id":   "proj-1",
		},
		"extra": anyMap{"email": "gem@example.com", "checked": true},
	}
	result, err := ConvertSub2ApiAccount(account, ConvertOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token, ok := result.Document["token"].(anyMap)
	if !ok {
		t.Fatalf("token missing: %#v", result.Document)
	}
	if token["access_token"] != "at" {
		t.Errorf("token.access_token = %v", token["access_token"])
	}
	if result.Document["checked"] != true {
		t.Errorf("checked = %v", result.Document["checked"])
	}
}

func TestConvertSub2ApiAccount_Gemini_MissingAccessToken(t *testing.T) {
	account := anyMap{"platform": "gemini", "credentials": anyMap{}}
	if _, err := ConvertSub2ApiAccount(account, ConvertOptions{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestConvertSub2ApiAccount_RejectsNonOAuthType(t *testing.T) {
	account := anyMap{"platform": "claude", "type": "apikey", "credentials": anyMap{"access_token": "x"}}
	if _, err := ConvertSub2ApiAccount(account, ConvertOptions{}); err == nil {
		t.Fatal("expected error for non-oauth type")
	}
}

func TestConvertSub2ApiAccount_UnsupportedPlatform(t *testing.T) {
	account := anyMap{"platform": "unknown-platform", "credentials": anyMap{"access_token": "x"}}
	if _, err := ConvertSub2ApiAccount(account, ConvertOptions{}); err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

func TestConvertSub2ApiAccount_NotAnObject(t *testing.T) {
	if _, err := ConvertSub2ApiAccount("not-a-map", ConvertOptions{}); err == nil {
		t.Fatal("expected error for non-object account")
	}
}

func TestBuildSub2ApiEntryLabel(t *testing.T) {
	if got := BuildSub2ApiEntryLabel(anyMap{"name": "Foo"}, 0); got != "Foo" {
		t.Errorf("got %q, want Foo", got)
	}
	if got := BuildSub2ApiEntryLabel(anyMap{}, 2); got != "accounts[2]" {
		t.Errorf("got %q, want accounts[2]", got)
	}
	if got := BuildSub2ApiEntryLabel("not-a-map", 5); got != "accounts[5]" {
		t.Errorf("got %q, want accounts[5]", got)
	}
	if got := BuildSub2ApiEntryLabel(anyMap{"extra": anyMap{"email": "a@b.c"}}, 1); got != "a@b.c" {
		t.Errorf("got %q, want a@b.c (extra.email fallback)", got)
	}
}

func TestConvertSub2ApiDocument_BatchWithSkipped(t *testing.T) {
	doc := anyMap{
		"accounts": []interface{}{
			anyMap{"platform": "claude", "credentials": anyMap{"access_token": "at1"}},
			anyMap{"platform": "unknown", "credentials": anyMap{}},
		},
	}
	batch, err := ConvertSub2ApiDocument(doc, ConvertOptions{SourceName: "src.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batch.Converted) != 1 {
		t.Errorf("Converted len = %d, want 1", len(batch.Converted))
	}
	if len(batch.Skipped) != 1 {
		t.Errorf("Skipped len = %d, want 1", len(batch.Skipped))
	}
}

func TestConvertSub2ApiDocument_EmptyAccounts(t *testing.T) {
	if _, err := ConvertSub2ApiDocument(anyMap{"accounts": []interface{}{}}, ConvertOptions{}); err == nil {
		t.Fatal("expected error for empty accounts")
	}
}
