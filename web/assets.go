package web

import (
	"embed"
	"net/http"
)

// Files are embedded so the workbench runs as one binary without a frontend build step.
//
//go:embed index.html styles.css app.js
var files embed.FS

func Handler() http.Handler { return http.FileServer(http.FS(files)) }
