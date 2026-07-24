package converter

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const isoLayout = "2006-01-02T15:04:05.000Z"

var dateParseLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	isoLayout,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02",
	time.RFC1123Z,
	time.RFC1123,
}

func formatISO(t time.Time) string {
	return t.UTC().Format(isoLayout)
}

// tryParseDate 尽量模拟 JS `new Date(string)` 的宽松解析行为。
func tryParseDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range dateParseLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// toFiniteNumber 对应 JS 版 toFiniteNumber：接受 float64 或可解析为数字的字符串。
func toFiniteNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return 0, false
		}
		return v, true
	case int:
		return float64(v), true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, false
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

// normalizeTimestamp 对应 JS 版 normalizeTimestamp。
func normalizeTimestamp(value interface{}) (string, bool) {
	switch v := value.(type) {
	case time.Time:
		if v.IsZero() {
			return "", false
		}
		return formatISO(v), true
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return "", false
		}
		return formatISO(time.UnixMilli(int64(v))), true
	case string:
		if strings.TrimSpace(v) == "" {
			return "", false
		}
		if t, ok := tryParseDate(v); ok {
			return formatISO(t), true
		}
		return "", false
	}
	return "", false
}

// timestampFromUnixLike 对应 JS 版 timestampFromUnixLike：兼容 unix 秒 / 毫秒。
func timestampFromUnixLike(value interface{}) (string, bool) {
	numeric, ok := toFiniteNumber(value)
	if !ok {
		return "", false
	}
	milliseconds := numeric
	if numeric <= 1e11 {
		milliseconds = numeric * 1000
	}
	return formatISO(time.UnixMilli(int64(milliseconds))), true
}

// normalizeFlexibleTimestamp 对应 JS 版 normalizeFlexibleTimestamp。
func normalizeFlexibleTimestamp(value interface{}) (string, bool) {
	if s, ok := normalizeTimestamp(value); ok {
		return s, true
	}
	return timestampFromUnixLike(value)
}

// normalizeUnixSecondsString 对应 JS 版 normalizeUnixSecondsString。
func normalizeUnixSecondsString(value interface{}) (string, bool) {
	normalized, ok := normalizeFlexibleTimestamp(value)
	if !ok {
		return "", false
	}
	t, ok := tryParseDate(normalized)
	if !ok {
		return "", false
	}
	return strconv.FormatInt(t.Unix(), 10), true
}

// timestampFromUnixSeconds 对应 JS 版 timestampFromUnixSeconds。
func timestampFromUnixSeconds(value interface{}) (string, bool) {
	numeric, ok := toFiniteNumber(value)
	if !ok {
		return "", false
	}
	return formatISO(time.Unix(int64(numeric), 0)), true
}

// timestampFromNowPlusSeconds 对应 JS 版 timestampFromNowPlusSeconds。
func timestampFromNowPlusSeconds(value interface{}, now time.Time) (string, bool) {
	seconds, ok := toFiniteNumber(value)
	if !ok {
		return "", false
	}
	if now.IsZero() {
		now = time.Now()
	}
	t := now.Add(time.Duration(seconds * float64(time.Second)))
	return formatISO(t), true
}

var emailKeyInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)
var emailKeyTrim = regexp.MustCompile(`^_+|_+$`)

// toEmailKey 对应 JS 版 toEmailKey。
func toEmailKey(email string) (string, bool) {
	if strings.TrimSpace(email) == "" {
		return "", false
	}
	lower := strings.ToLower(strings.TrimSpace(email))
	replaced := emailKeyInvalidChars.ReplaceAllString(lower, "_")
	trimmed := emailKeyTrim.ReplaceAllString(replaced, "")
	return trimmed, true
}

// joinScopes 对应 JS 版 joinScopes：字符串原样返回，数组则过滤空白后以空格拼接。
func joinScopes(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return "", false
		}
		return s, true
	case []interface{}:
		var parts []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					parts = append(parts, trimmed)
				}
			}
		}
		if len(parts) == 0 {
			return "", false
		}
		return strings.Join(parts, " "), true
	}
	return "", false
}

// getExpiresIn 对应 JS 版 getExpiresIn：expiresAt - now，单位秒，向下取整且不小于 0。
func getExpiresIn(expiresAt string, now time.Time) (float64, bool) {
	if strings.TrimSpace(expiresAt) == "" {
		return 0, false
	}
	t, ok := tryParseDate(expiresAt)
	if !ok {
		return 0, false
	}
	if now.IsZero() {
		now = time.Now()
	}
	diff := math.Floor(t.Sub(now).Seconds())
	if diff < 0 {
		diff = 0
	}
	return diff, true
}

// debugFormat 仅用于错误信息中安全展示不可预期类型的值。
func debugFormat(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
