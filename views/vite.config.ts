import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 开发模式下通过 Vite 代理把 /api/* 转发给本地 Go 后端（默认 :8080），
// 前端始终只请求同源路径，无需在 Go 后端处理 CORS。
// 生产环境下前后端同源（Go 单进程通过 go:embed 同时提供 API 与静态资源），同样无跨域问题。
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
