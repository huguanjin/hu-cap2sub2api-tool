import type { ConvertResponse } from './types'

/**
 * 上传一批 JSON 文件到后端 /api/convert 进行双向转换。
 * 无状态：后端不落盘、不保留 session，直接返回内联的转换结果。
 */
export async function convertFiles(files: File[]): Promise<ConvertResponse> {
  const formData = new FormData()
  for (const file of files) {
    formData.append('files', file)
  }

  const res = await fetch('/api/convert', {
    method: 'POST',
    body: formData,
  })

  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new Error(text || `请求失败：HTTP ${res.status}`)
  }

  return (await res.json()) as ConvertResponse
}

/** 探测后端是否可用，可选调用（例如启动时的健康检查）。 */
export async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch('/api/health')
    if (!res.ok) return false
    const data = (await res.json()) as { status?: string }
    return data.status === 'ok'
  } catch {
    return false
  }
}
