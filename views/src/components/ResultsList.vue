<script setup lang="ts">
import type { ConvertResponse, WrittenFile } from '../api/types'
import { buildZip } from '../utils/zip'

const props = defineProps<{ result: ConvertResponse | null }>()

function directionLabel(direction: WrittenFile['direction']): string {
  return direction === 'cpa-to-sub2api' ? 'CPA → sub2api' : 'sub2api → CPA'
}

function downloadOne(file: WrittenFile) {
  const blob = new Blob([file.content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = file.name
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

function downloadAll() {
  const files = props.result?.written ?? []
  if (!files.length) return

  const usedNames = new Set<string>()
  const entries = files.map((file) => {
    let name = file.name
    let suffix = 1
    while (usedNames.has(name)) {
      const dotIndex = file.name.lastIndexOf('.')
      name =
        dotIndex === -1
          ? `${file.name}-${suffix}`
          : `${file.name.slice(0, dotIndex)}-${suffix}${file.name.slice(dotIndex)}`
      suffix += 1
    }
    usedNames.add(name)
    return { name, content: file.content }
  })

  const blob = buildZip(entries)
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `converted-files-${Date.now()}.zip`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div v-if="result" class="results">
    <div v-if="result.written.length" class="section">
      <div class="section-header">
        <h3>转换成功（{{ result.written.length }}）</h3>
        <button type="button" class="download-all-btn" @click="downloadAll">全部下载（打包 zip）</button>
      </div>
      <ul class="written-list">
        <li v-for="file in result.written" :key="file.name" class="written-item">
          <span class="badge" :class="file.direction">{{ directionLabel(file.direction) }}</span>
          <span class="file-name">{{ file.name }}</span>
          <span v-if="file.label" class="file-label">{{ file.label }}</span>
          <button type="button" class="download-btn" @click="downloadOne(file)">下载</button>
        </li>
      </ul>
    </div>

    <div v-if="result.skipped.length" class="section">
      <h3>已跳过（{{ result.skipped.length }}）</h3>
      <ul class="issue-list">
        <li v-for="(item, idx) in result.skipped" :key="`${item.sourceName}-${idx}`" class="issue-item">
          <span class="file-name">{{ item.sourceName }}</span>
          <span v-if="item.entryLabel" class="entry-label">{{ item.entryLabel }}</span>
          <span class="reason">{{ item.reason }}</span>
        </li>
      </ul>
    </div>

    <div v-if="result.parseErrors.length" class="section">
      <h3>解析失败（{{ result.parseErrors.length }}）</h3>
      <ul class="issue-list">
        <li v-for="item in result.parseErrors" :key="item.name" class="issue-item">
          <span class="file-name">{{ item.name }}</span>
          <span class="reason">{{ item.reason }}</span>
        </li>
      </ul>
    </div>

    <p v-if="!result.written.length && !result.skipped.length && !result.parseErrors.length" class="empty">
      没有可展示的结果。
    </p>
  </div>
</template>

<style scoped>
.results {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

h3 {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: #ddd;
}

.download-all-btn {
  background: #2d6cdf;
  border: none;
  color: #fff;
  border-radius: 6px;
  padding: 0.35rem 0.75rem;
  cursor: pointer;
  font-size: 0.8rem;
}

.download-all-btn:hover {
  background: #4d84e8;
}

.written-list,
.issue-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.written-item,
.issue-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  background: #1c1c1c;
  border: 1px solid #333;
  border-radius: 6px;
  padding: 0.4rem 0.6rem;
  font-size: 0.85rem;
}

.badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  white-space: nowrap;
}

.badge.cpa-to-sub2api {
  background: #23412b;
  color: #7ee08c;
}

.badge.sub2api-to-cpa {
  background: #2b3542;
  color: #7eb8e0;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-label,
.entry-label {
  color: #999;
  font-size: 0.75rem;
}

.reason {
  color: #e08c7e;
  font-size: 0.8rem;
}

.download-btn {
  background: transparent;
  border: 1px solid #555;
  color: #ccc;
  border-radius: 4px;
  padding: 0.2rem 0.5rem;
  cursor: pointer;
  font-size: 0.75rem;
}

.download-btn:hover {
  border-color: #4d9fff;
  color: #4d9fff;
}

.empty {
  color: #888;
  font-size: 0.85rem;
}
</style>
