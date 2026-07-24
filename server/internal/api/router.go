// Package api 提供 HTTP 路由与请求处理器：/api/convert、/api/health，
// 以及非 API 路径下的前端静态资源 + SPA 回退服务。
package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewRouter 注册全部路由：API 路径由具体 handler 处理，其余路径回退到
// 嵌入的前端静态资源（SPA fallback：找不到对应静态文件时返回 index.html）。
func NewRouter(staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", HealthHandler)
	mux.HandleFunc("POST /api/convert", ConvertHandler)
	mux.Handle("/", spaHandler(staticFS))
	return mux
}

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleaned := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if cleaned == "." || cleaned == "" {
			cleaned = "index.html"
		}

		if _, err := fs.Stat(staticFS, cleaned); err != nil {
			// 静态资源中不存在该路径（例如前端路由 /some/route）→ 回退到 index.html，
			// 交由前端路由处理。
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r2)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}
