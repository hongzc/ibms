// Package web 嵌入前端静态页面。
package web

import "embed"

//go:embed *.html
var FS embed.FS
