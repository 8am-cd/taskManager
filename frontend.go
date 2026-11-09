package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web
var frontendFiles embed.FS

func registerFrontend(mux *http.ServeMux) {
	web, err := fs.Sub(frontendFiles, "web")
	if err != nil {
		panic(err)
	}

	files := http.FileServerFS(web)
	mux.Handle("GET /{$}", files)
	mux.Handle("GET /assets/", files)
}
