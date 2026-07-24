package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/huguanjin/hu-cap2sub2api-tool/server/internal/converter"
)

// maxUploadMemory 是 multipart 表单解析时允许驻留内存的最大字节数（超出部分落临时文件），
// 该工具面向小体积的凭证 JSON 文件，32MB 足够宽裕。
const maxUploadMemory = 32 << 20

// WrittenFile 对应一份可供前端下载的转换产物。
type WrittenFile struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	Direction string `json:"direction"` // "cpa-to-sub2api" | "sub2api-to-cpa"
	Label     string `json:"label,omitempty"`
}

// SkippedItem 表示被跳过（无法转换）的条目。
type SkippedItem struct {
	SourceName string `json:"sourceName"`
	EntryLabel string `json:"entryLabel,omitempty"`
	Reason     string `json:"reason"`
}

// ParseErrorItem 表示上传文件本身无法解析为 JSON。
type ParseErrorItem struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// ConvertResponse 是 POST /api/convert 的响应体。
type ConvertResponse struct {
	Written     []WrittenFile    `json:"written"`
	Skipped     []SkippedItem    `json:"skipped"`
	ParseErrors []ParseErrorItem `json:"parseErrors"`
}

// HealthHandler 处理 GET /api/health。
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ConvertHandler 处理 POST /api/convert：接收一个或多个 JSON 文件（multipart 字段名 "files"），
// 自动识别每个文件是 CPA 还是 sub2api 格式并双向转换。整个请求无状态，不落盘、不保留 session，
// 转换结果以内联 JSON 字符串的形式返回，由前端触发浏览器下载。
func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
		http.Error(w, "无法解析上传的表单数据", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "未上传任何文件", http.StatusBadRequest)
		return
	}

	resp := ConvertResponse{
		Written:     []WrittenFile{},
		Skipped:     []SkippedItem{},
		ParseErrors: []ParseErrorItem{},
	}
	var cpaResults []*converter.CPAConversionResult

	for _, fh := range files {
		name := fh.Filename

		content, err := readUploadedFile(fh)
		if err != nil {
			resp.ParseErrors = append(resp.ParseErrors, ParseErrorItem{Name: name, Reason: "无法读取上传的文件"})
			continue
		}

		var doc interface{}
		if err := json.Unmarshal(content, &doc); err != nil {
			resp.ParseErrors = append(resp.ParseErrors, ParseErrorItem{Name: name, Reason: "不是合法的 JSON 文件"})
			continue
		}

		opts := converter.ConvertOptions{SourceName: name}

		switch converter.DetectFormat(doc) {
		case "cpa":
			record, ok := doc.(map[string]interface{})
			if !ok {
				resp.Skipped = append(resp.Skipped, SkippedItem{SourceName: name, Reason: "文件不是 JSON 对象"})
				continue
			}
			result, err := converter.ConvertCPARecord(record, opts)
			if err != nil {
				resp.Skipped = append(resp.Skipped, SkippedItem{SourceName: name, Reason: err.Error()})
				continue
			}
			cpaResults = append(cpaResults, result)

		case "sub2api":
			batch, err := converter.ConvertSub2ApiDocument(doc, opts)
			if err != nil {
				resp.Skipped = append(resp.Skipped, SkippedItem{SourceName: name, Reason: err.Error()})
				continue
			}
			for _, item := range batch.Converted {
				prettyContent, mErr := marshalPretty(item.Document)
				if mErr != nil {
					resp.Skipped = append(resp.Skipped, SkippedItem{SourceName: name, EntryLabel: item.EntryLabel, Reason: "生成输出内容失败"})
					continue
				}
				resp.Written = append(resp.Written, WrittenFile{
					Name:      item.OutputFileName,
					Content:   prettyContent,
					Direction: "sub2api-to-cpa",
					Label:     item.EntryLabel,
				})
			}
			for _, s := range batch.Skipped {
				resp.Skipped = append(resp.Skipped, SkippedItem{SourceName: s.SourceName, EntryLabel: s.EntryLabel, Reason: s.Reason})
			}

		default:
			resp.Skipped = append(resp.Skipped, SkippedItem{SourceName: name, Reason: "无法识别文件格式（既不是 CPA 也不是 sub2api）"})
		}
	}

	if len(cpaResults) > 0 {
		merged := converter.BuildMergedSub2ApiDocument(cpaResults, converter.ConvertOptions{})
		prettyContent, err := marshalPretty(merged.Document)
		if err != nil {
			resp.Skipped = append(resp.Skipped, SkippedItem{Reason: "生成合并输出内容失败"})
		} else {
			resp.Written = append(resp.Written, WrittenFile{
				Name:      merged.OutputFileName,
				Content:   prettyContent,
				Direction: "cpa-to-sub2api",
				Label:     fmt.Sprintf("合并 %d 个账号", merged.Count),
			})
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func readUploadedFile(fh *multipart.FileHeader) ([]byte, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	// 去除可能存在的 UTF-8 BOM，避免 json.Unmarshal 解析失败。
	return bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}), nil
}

func marshalPretty(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
