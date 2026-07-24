# CPA ⇄ sub2api 转换工具

Web 应用：在 **CPA（CLIProxyApi）认证 JSON** 与 **sub2api 配置格式**之间自动双向转换，支持 `codex` / `claude` / `antigravity` / `gemini` 四种平台。

后端 **Go**（零第三方依赖，仅 stdlib）+ 前端 **Vue3 + Vite + TypeScript**（无 UI 框架依赖），前端构建产物通过 `go:embed` 打包进单一二进制 / 单一 Docker 镜像，部署即单进程、无外部依赖、无需数据库。所有转换均在单次 HTTP 请求内完成，服务端不落盘、不保留任何用户数据。

## 目录结构

```text
.
├── server/            # Go 后端（go.mod 在此目录下）
│   ├── cmd/server/    # 程序入口
│   └── internal/
│       ├── converter/ # 核心转换逻辑（CPA ⇄ sub2api）
│       ├── api/       # HTTP 路由与 handler
│       └── web/       # go:embed 前端产物
├── views/             # Vue3 前端（Vite + TypeScript）
├── Dockerfile         # 多阶段构建：views → server → distroless 运行时
└── .github/workflows/ # CI：push 后自动构建并推送镜像到 GHCR
```

## 本地开发

需要 Go ≥ 1.23、Node.js ≥ 22。

**后端**（默认监听 `:8080`，可用 `PORT` 环境变量覆盖）：

```bash
cd server
go run ./cmd/server
```

**前端**（Vite dev server，`/api/*` 会自动代理到 `http://localhost:8080`）：

```bash
cd views
npm install
npm run dev
```

开发时两个进程分开跑即可，无需手动处理跨域。

## 构建与测试

```bash
# 后端
cd server
go build ./... && go vet ./... && go test ./...

# 前端（含 vue-tsc 类型检查）
cd views
npm install
npm run build
```

## 生产部署（Docker）

镜像多阶段构建：Node 构建前端 → Go 构建后端（`go:embed` 内嵌前端产物）→ `distroless/static-debian12:nonroot` 运行时，最终镜像不含 shell、包管理器等多余内容。

```bash
docker build -t cpa2sub2api-tool:local .
docker run --rm -p 8080:8080 cpa2sub2api-tool:local
```

访问 `http://localhost:8080` 即可使用。

### 从 GHCR 拉取

每次 push 到 `main` 分支（或打 `v*` tag）时，GitHub Actions 会自动构建镜像并推送到 GHCR：

```bash
docker pull ghcr.io/huguanjin/hu-cap2sub2api-tool:latest
docker run --rm -p 8080:8080 ghcr.io/huguanjin/hu-cap2sub2api-tool:latest
```

## API

- `GET /api/health` → `{"status":"ok"}`
- `POST /api/convert`（`multipart/form-data`，字段名 `files`，可多文件）

  响应：

  ```json
  {
    "written": [{ "name": "...", "content": "<格式化 JSON 字符串>", "direction": "cpa-to-sub2api|sub2api-to-cpa", "label": "..." }],
    "skipped": [{ "sourceName": "...", "entryLabel": "...", "reason": "..." }],
    "parseErrors": [{ "name": "...", "reason": "..." }]
  }
  ```

  转换结果以内联 JSON 字符串形式返回，由浏览器端触发下载；服务端不生成压缩包、不持久化任何文件。

## 转换约定

| 项 | 行为 |
|----|------|
| 方向 | 自动识别 CPA / sub2api，转换为另一种格式 |
| 输出样式 | 始终美化 JSON（缩进 2） |
| CPA → sub2api | 同一批上传合并为 1 个文件（多个账号进 `accounts` 数组） |
| sub2api → CPA | 1 个输入文件按账号拆分为 N 个 CPA 文件 |
| 数据落盘 | 全程无服务端持久化，转换完全在单次请求内完成 |

## 支持平台

`codex` / `claude` / `antigravity` / `gemini`
