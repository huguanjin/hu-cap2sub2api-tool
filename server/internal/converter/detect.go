package converter

import "strings"

var cpaTypes = map[string]bool{
	"codex":       true,
	"claude":      true,
	"antigravity": true,
	"gemini":      true,
}

// DetectFormat 自动识别 JSON 文档是 CPA 还是 sub2api 格式。
// 识别策略偏保守，避免把无关 JSON 误判为 CPA。
// 对应 JS 版 lib/detect.mjs 的 detectFormat。
func DetectFormat(document interface{}) string {
	if _, ok := document.([]interface{}); ok {
		return "sub2api"
	}

	doc, ok := asMap(document)
	if !ok {
		return "unknown"
	}

	// sub2api：合并配置或单账号形态
	if _, ok := getSlice(doc, "accounts"); ok {
		return "sub2api"
	}

	if platform, ok := getString(doc, "platform"); ok && strings.TrimSpace(platform) != "" {
		if _, ok := getMap(doc, "credentials"); ok {
			return "sub2api"
		}
	}

	// CPA：显式 type
	if typ, ok := getString(doc, "type"); ok {
		t := strings.ToLower(strings.TrimSpace(typ))
		if cpaTypes[t] {
			return "cpa"
		}
		// 未知 type 不当作 CPA，避免误伤
		return "unknown"
	}

	// CPA：无 type 时仅在有明显凭证特征时认作 codex / gemini 形态
	if accessToken, ok := getString(doc, "access_token"); ok && strings.TrimSpace(accessToken) != "" {
		if idToken, ok := getString(doc, "id_token"); ok && strings.TrimSpace(idToken) != "" {
			return "cpa"
		}
	}

	if tok, ok := getMap(doc, "token"); ok {
		if at, ok := getString(tok, "access_token"); ok && strings.TrimSpace(at) != "" {
			return "cpa"
		}
		if at, ok := getString(tok, "accessToken"); ok && strings.TrimSpace(at) != "" {
			return "cpa"
		}
	}

	return "unknown"
}
