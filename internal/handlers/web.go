package handlers

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var webFS embed.FS

func ServeWeb() http.Handler {
	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(webContent))
}
