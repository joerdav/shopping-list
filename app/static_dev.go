//+build dev
//go:build dev

package app

import (
	"os"
	"net/http"
)

// This package allows for tailwind to rebuild the css without rebuilding the entire binary.

func public() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}
