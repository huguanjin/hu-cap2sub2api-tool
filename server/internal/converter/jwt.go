package converter

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// ParseJWTPayload 解析 JWT 的 payload 段（不校验签名），对应 JS 版 parseJwtPayload。
// token 为空、格式不合法或 payload 无法解析为 JSON 对象时返回 nil, err。
func ParseJWTPayload(token string) (anyMap, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("空 token")
	}

	segments := strings.Split(token, ".")
	if len(segments) < 2 {
		return nil, errors.New("token 段数不足")
	}

	normalized := strings.NewReplacer("-", "+", "_", "/").Replace(segments[1])
	// Go 的 StdEncoding 需要正确的 padding；RawURLEncoding 则不需要 padding，
	// 但这里已经把 URL-safe 字符替换成标准字符集，因此改用 StdEncoding 并手动补 "="。
	if rem := len(normalized) % 4; rem != 0 {
		normalized += strings.Repeat("=", 4-rem)
	}

	decoded, err := base64.StdEncoding.DecodeString(normalized)
	if err != nil {
		return nil, err
	}

	var payload interface{}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}

	m, ok := asMap(payload)
	if !ok {
		return nil, errors.New("payload 不是 JSON 对象")
	}
	return m, nil
}
