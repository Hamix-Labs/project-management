// Package handler registers /repo/* REST routes for taskapi.
package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
)

// Deps wires workspace repo HTTP handlers into the taskapi mux.
type Deps struct {
	Provider repo.RepoProvider
}

// Handler serves /repo/* REST routes.
type Handler struct {
	provider repo.RepoProvider
}

// Register mounts /repo/* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{provider: deps.Provider}
	m.Handle("GET /repo/search", http.HandlerFunc(h.repoSearch))
	m.Handle("GET /repo/file", http.HandlerFunc(h.repoFile))
	m.Handle("GET /repo/validate-range", http.HandlerFunc(h.repoValidateRange))
	m.Handle("GET /repo/diff", http.HandlerFunc(h.repoDiff))
}
