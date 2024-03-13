//go:build dev

package app

import (
	"net/http"
	"os"
)

// This package allows for tailwind to rebuild the css without rebuilding the entire binary.

func public() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("app/public")))
}
