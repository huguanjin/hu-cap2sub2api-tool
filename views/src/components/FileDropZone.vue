<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ files: File[] }>()
const emit = defineEmits<{
  (e: 'update:files', files: File[]): void
}>()

const isDragOver = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

function keyOf(file: File): string {
  return `${file.name}::${file.size}::${file.lastModified}`
}

function addFiles(list: FileList | File[]) {
  const incoming = Array.from(list).filter((f) => f.name.toLowerCase().endsWith('.json'))
  if (incoming.length === 0) return

  const seen = new Set(props.files.map(keyOf))
  const merged = [...props.files]
  for (const file of incoming) {
    const key = keyOf(file)
    if (!seen.has(key)) {
      seen.add(key)
      merged.push(file)
    }
  }
  emit('update:files', merged)
}

function onDrop(e: DragEvent) {
  isDragOver.value = false
  if (e.dataTransfer?.files?.length) {
    addFiles(e.dataTransfer.files)
  }
}

function onInputChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.length) {
    addFiles(target.files)
  }
  target.value = ''
}

function openPicker() {
  inputRef.value?.click()
}

function removeAt(index: number) {
  const next = props.files.slice()
  next.splice(index, 1)
  emit('update:files', next)
}

function clearAll() {
  emit('update:files', [])
}
</script>

<template>
  <div class="drop-zone-wrap">
    <div
      class="drop-zone"
      :class="{ 'is-drag-over': isDragOver }"
      @click="openPicker"
      @dragover.prevent="isDragOver = true"
      @dragleave.prevent="isDragOver = false"
      @drop.prevent="onDrop"
    >
      <p class="hint">拖拽 JSON 文件到此处，或点击选择文件</p>
      <p class="sub-hint">支持 CPA 与 sub2api 格式，多文件混合上传</p>
      <input
        ref="inputRef"
        type="file"
        accept=".json,application/json"
        multiple
        class="hidden-input"
        @change="onInputChange"
      />
    </div>

    <ul v-if="files.length" class="file-list">
      <li v-for="(file, index) in files" :key="keyOf(file)" class="file-item">
        <span class="file-name">{{ file.name }}</span>
        <span class="file-size">{{ (file.size / 1024).toFixed(1) }} KB</span>
        <button type="button" class="remove-btn" @click.stop="removeAt(index)">移除</button>
      </li>
    </ul>

    <button v-if="files.length" type="button" class="clear-btn" @click="clearAll">清空全部</button>
  </div>
</template>

<style scoped>
.drop-zone-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.drop-zone {
  border: 2px dashed #555;
  border-radius: 8px;
  padding: 2rem 1rem;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.drop-zone.is-drag-over {
  border-color: #4d9fff;
  background-color: rgba(77, 159, 255, 0.08);
}

.hint {
  margin: 0;
  font-size: 1rem;
  color: #ddd;
}

.sub-hint {
  margin: 0.35rem 0 0;
  font-size: 0.8rem;
  color: #888;
}

.hidden-input {
  display: none;
}

.file-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  background: #1c1c1c;
  border: 1px solid #333;
  border-radius: 6px;
  padding: 0.4rem 0.6rem;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.85rem;
}

.file-size {
  font-size: 0.75rem;
  color: #888;
}

.remove-btn,
.clear-btn {
  background: transparent;
  border: 1px solid #555;
  color: #ccc;
  border-radius: 4px;
  padding: 0.2rem 0.5rem;
  cursor: pointer;
  font-size: 0.75rem;
}

.remove-btn:hover,
.clear-btn:hover {
  border-color: #4d9fff;
  color: #4d9fff;
}

.clear-btn {
  align-self: flex-start;
}
</style>
