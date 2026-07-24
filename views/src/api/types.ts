// 与后端 server/internal/api/handlers.go 中的 ConvertResponse 一一对应。

export interface WrittenFile {
  name: string
  content: string
  direction: 'cpa-to-sub2api' | 'sub2api-to-cpa'
  label?: string
}

export interface SkippedItem {
  sourceName: string
  entryLabel?: string
  reason: string
}

export interface ParseErrorItem {
  name: string
  reason: string
}

export interface ConvertResponse {
  written: WrittenFile[]
  skipped: SkippedItem[]
  parseErrors: ParseErrorItem[]
}
