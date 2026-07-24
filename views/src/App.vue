<script setup lang="ts">
import { ref } from 'vue'
import FileDropZone from './components/FileDropZone.vue'
import ResultsList from './components/ResultsList.vue'
import { convertFiles } from './api/client'
import type { ConvertResponse } from './api/types'

const files = ref<File[]>([])
const result = ref<ConvertResponse | null>(null)
const loading = ref(false)
const error = ref('')

async function handleConvert() {
  if (files.value.length === 0 || loading.value) return
  loading.value = true
  error.value = ''
  try {
    result.value = await convertFiles(files.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="app">
    <header class="header">
      <h1>CPA ⇄ sub2api 转换工具</h1>
      <p class="desc">
        在浏览器中双向转换 CLIProxyApi（CPA）认证文件与 sub2api 配置文件，支持 codex / claude / antigravity /
        gemini 四种平台。文件仅在本次请求中处理，不会被服务端保存。
      </p>
    </header>

    <section class="panel">
      <FileDropZone v-model:files="files" />

      <button type="button" class="convert-btn" :disabled="files.length === 0 || loading" @click="handleConvert">
        {{ loading ? '转换中…' : '开始转换' }}
      </button>

      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <section v-if="result" class="panel">
      <ResultsList :result="result" />
    </section>
  </main>
</template>

<style scoped>
.app {
  max-width: 720px;
  margin: 0 auto;
  padding: 2.5rem 1.25rem 4rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

.header h1 {
  margin: 0 0 0.5rem;
  font-size: 1.5rem;
}

.desc {
  margin: 0;
  color: #999;
  font-size: 0.85rem;
  line-height: 1.5;
}

.panel {
  background: #181818;
  border: 1px solid #2a2a2a;
  border-radius: 10px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.convert-btn {
  align-self: flex-start;
  background: #2d6cdf;
  border: none;
  color: #fff;
  border-radius: 6px;
  padding: 0.5rem 1.25rem;
  cursor: pointer;
  font-size: 0.9rem;
}

.convert-btn:disabled {
  background: #3a3a3a;
  color: #777;
  cursor: not-allowed;
}

.convert-btn:not(:disabled):hover {
  background: #4d84e8;
}

.error {
  color: #e08c7e;
  font-size: 0.85rem;
  margin: 0;
}
</style>
