package converter

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// makeJWT 构造一个不校验签名的最小 JWT 字符串，供测试使用。
func makeJWT(t *testing.T, payload anyMap) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal jwt payload: %v", err)
	}
	body := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return header + "." + body + ".sig"
}
