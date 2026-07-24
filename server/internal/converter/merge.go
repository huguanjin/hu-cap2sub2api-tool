package converter

// MergedSub2ApiResult 表示多条 CPA 记录合并后的单个 sub2api 文档，对应 Web 版批量转换场景中
// 一次请求上传的多个 CPA 文件被合并为一个 sub2api 输出文件的结果。
type MergedSub2ApiResult struct {
	Document       anyMap
	OutputFileName string
	Count          int
}

// BuildMergedSub2ApiDocument 将多条 CPA→sub2api 转换结果的 account 合并进同一份 sub2api
// 文档（accounts 数组），对应 JS 版 buildMergedSub2ApiDocument。
func BuildMergedSub2ApiDocument(results []*CPAConversionResult, opts ConvertOptions) *MergedSub2ApiResult {
	if len(results) == 0 {
		return nil
	}

	accounts := make([]interface{}, 0, len(results))
	for _, r := range results {
		accounts = append(accounts, r.Account)
	}

	exportedAt, _ := normalizeTimestamp(opts.now())
	document := anyMap{
		"exported_at": exportedAt,
		"proxies":     []interface{}{},
		"accounts":    accounts,
	}

	return &MergedSub2ApiResult{
		Document:       document,
		OutputFileName: BuildMergedSub2ApiFileName(results),
		Count:          len(results),
	}
}
