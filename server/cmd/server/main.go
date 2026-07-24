// Command server 启动 cpa2sub2api 工具的 Web 后端：提供 /api/convert、/api/health，
// 并在同一进程内通过 go:embed 提供 Vue3 前端的静态资源。
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/huguanjin/hu-cap2sub2api-tool/server/internal/api"
	"github.com/huguanjin/hu-cap2sub2api-tool/server/internal/web"
)

func main() {
	staticFS, err := web.FS()
	if err != nil {
		log.Fatalf("加载前端静态资源失败: %v", err)
	}

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	handler := api.NewRouter(staticFS)

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
