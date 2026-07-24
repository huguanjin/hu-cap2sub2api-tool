package converter

import "testing"

func TestParseJWTPayload_Valid(t *testing.T) {
	token := makeJWT(t, anyMap{"sub": "123", "exp": float64(1700000000)})
	payload, err := ParseJWTPayload(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload["sub"] != "123" {
		t.Errorf("sub = %v, want 123", payload["sub"])
	}
}

func TestParseJWTPayload_Empty(t *testing.T) {
	if _, err := ParseJWTPayload(""); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestParseJWTPayload_TooFewSegments(t *testing.T) {
	if _, err := ParseJWTPayload("onlyonepart"); err == nil {
		t.Fatal("expected error for malformed JWT (too few segments)")
	}
}

func TestParseJWTPayload_InvalidBase64(t *testing.T) {
	if _, err := ParseJWTPayload("aaa.!!!not-base64!!!.bbb"); err == nil {
		t.Fatal("expected error for invalid base64 payload")
	}
}

func TestParseJWTPayload_InvalidJSON(t *testing.T) {
	bad := "bm90IGpzb24" // base64url("not json") without padding
	if _, err := ParseJWTPayload("h." + bad + ".s"); err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestParseJWTPayload_PayloadNotObject(t *testing.T) {
	// payload 段是合法 JSON，但不是对象（例如数组）。
	arrayPayload := "WzEsMiwzXQ" // base64url(`[1,2,3]`)
	if _, err := ParseJWTPayload("h." + arrayPayload + ".s"); err == nil {
		t.Fatal("expected error when payload is not a JSON object")
	}
}
