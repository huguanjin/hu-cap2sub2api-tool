// Package web 通过 go:embed 打包 Vue3 前端的构建产物（views/dist），
// 供 Go 后端在生产环境下以单进程同时提供 API 与静态资源。
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// FS 返回嵌入的前端构建产物（dist 子目录）。
func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
