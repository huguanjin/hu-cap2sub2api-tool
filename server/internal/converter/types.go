// Package converter 在 CPA（CLIProxyApi 认证 JSON）与 sub2api 配置格式之间
// 进行双向转换。逻辑移植自原 Node.js 实现（lib/converter.mjs、lib/detect.mjs），
// 字段回退顺序、错误文案与输出格式均保持一致。
package converter

import (
	"strings"
	"time"
)

// anyMap 是贯穿本包的"鸭子类型"文档表示，对应 JS 中的 plain object。
type anyMap = map[string]interface{}

func isPlainObject(v interface{}) bool {
	if v == nil {
		return false
	}
	_, ok := v.(anyMap)
	return ok
}

func asMap(v interface{}) (anyMap, bool) {
	m, ok := v.(anyMap)
	return m, ok
}

func asSlice(v interface{}) ([]interface{}, bool) {
	s, ok := v.([]interface{})
	return s, ok
}

// getString 返回 m[key] 作为字符串（非字符串类型返回 "", false）。
func getString(m anyMap, key string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getMap(m anyMap, key string) (anyMap, bool) {
	if m == nil {
		return nil, false
	}
	return asMap(m[key])
}

func getSlice(m anyMap, key string) ([]interface{}, bool) {
	if m == nil {
		return nil, false
	}
	return asSlice(m[key])
}

func getBool(m anyMap, key string) (bool, bool) {
	if m == nil {
		return false, false
	}
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func getFloat(m anyMap, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}

// firstNonEmpty 返回第一个非空白的字符串（对应 JS 版 firstNonEmpty，仅支持字符串场景，
// 与原实现中调用方始终传字符串参数的用法一致）。
func firstNonEmpty(values ...string) (string, bool) {
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			return trimmed, true
		}
	}
	return "", false
}

// firstNonEmptyOr 是 firstNonEmpty 的便捷版本，找不到时返回 ""。
func firstNonEmptyOr(values ...string) string {
	s, _ := firstNonEmpty(values...)
	return s
}

// ConvertOptions 对应 JS 版 options 参数（sourceName / now）。
type ConvertOptions struct {
	SourceName string
	Now        time.Time // 零值表示使用 time.Now()
}

func (o ConvertOptions) now() time.Time {
	if o.Now.IsZero() {
		return time.Now()
	}
	return o.Now
}

// SkippedEntry 对应批量转换中被跳过的条目。
type SkippedEntry struct {
	SourceName string `json:"sourceName"`
	EntryLabel string `json:"entryLabel,omitempty"`
	Reason     string `json:"reason"`
}
