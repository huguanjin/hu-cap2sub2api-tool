package converter

import (
	"regexp"
	"strings"
)

var (
	extSuffixRe     = regexp.MustCompile(`\.[^.]+$`)
	unsafeCharsRe   = regexp.MustCompile(`[\\/:*?"<>|]+`)
	whitespaceRe    = regexp.MustCompile(`\s+`)
	dashCollapseRe  = regexp.MustCompile(`-+`)
	dashTrimRe      = regexp.MustCompile(`^-+|-+$`)
)

// SanitizeBaseName 对应 JS 版 sanitizeBaseName：去扩展名、替换路径不安全字符与空白为
// "-"，合并连续 "-"，去首尾 "-"，转小写。
func SanitizeBaseName(name string) string {
	s := extSuffixRe.ReplaceAllString(name, "")
	s = unsafeCharsRe.ReplaceAllString(s, "-")
	s = whitespaceRe.ReplaceAllString(s, "-")
	s = dashCollapseRe.ReplaceAllString(s, "-")
	s = dashTrimRe.ReplaceAllString(s, "")
	return strings.ToLower(s)
}

// BuildSub2ApiOutputFileName 对应 JS 版 buildSub2ApiOutputFileName。
func BuildSub2ApiOutputFileName(sourceName, email string) string {
	sourceBase := ""
	if strings.TrimSpace(sourceName) != "" {
		parts := strings.Split(sourceName, "/")
		last := parts[len(parts)-1]
		sourceBase = SanitizeBaseName(last)
	}

	emailBase := ""
	if strings.TrimSpace(email) != "" {
		emailBase = SanitizeBaseName(email)
	}

	base := sourceBase
	if base == "" {
		base = emailBase
	}
	if base == "" {
		base = "converted-account"
	}
	return base + ".sub2api.json"
}

// BuildMergedSub2ApiFileName 为合并后的 sub2api 文档生成输出文件名：
// 单条记录时复用其自身的输出文件名；多条记录合并为一份文档时使用统一文件名。
func BuildMergedSub2ApiFileName(results []*CPAConversionResult) string {
	if len(results) == 1 {
		return results[0].OutputFileName
	}
	return "merged-accounts.sub2api.json"
}

// BuildCPAOutputFileName 对应 JS 版 buildCPAOutputFileName。
func BuildCPAOutputFileName(accountName, email, providerType string) string {
	seed := firstNonEmptyOr(email, accountName, providerType, "converted-account")
	base := SanitizeBaseName(seed)

	typeSeed := providerType
	if strings.TrimSpace(typeSeed) == "" {
		typeSeed = "account"
	}
	typeBase := SanitizeBaseName(typeSeed)

	if base == "" || base == typeBase {
		return typeBase + ".cpa.json"
	}
	return base + "." + typeBase + ".cpa.json"
}
