package desktop

import (
	"io"
	"io/fs"
	"net/http"
	"strings"
)

// API path prefixes that must never receive SPA index.html fallback.
var apiPrefixes = []string{
	"/tasks",
	"/task-drafts",
	"/task-templates",
	"/events",
	"/projects",
	"/git",
	"/settings",
	"/repo",
	"/runners",
	"/health",
	"/system",
	"/metrics",
	"/v1",
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func isAPIPath(path string) bool {
	if path == "/" {
		return false
	}
	for _, p := range apiPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func wantsHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return true
	}
	// Prefer HTML document navigations over */* API clients.
	if strings.Contains(accept, "text/html") {
		return true
	}
	if strings.Contains(accept, "application/json") {
		return false
	}
	return strings.Contains(accept, "*/*")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func serveIndexHTML(w http.ResponseWriter, assets fs.FS) bool {
	if assets == nil {
		return false
	}
	f, err := assets.Open("index.html")
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
	return true
}
