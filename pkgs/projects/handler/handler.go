// Package handler registers /projects* REST routes for taskapi.
package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// NotifyFunc aliases the shared hint-with-id SSE notify type.
type NotifyFunc = realtime.NotifyFunc

// Deps wires project HTTP handlers into the taskapi mux.
type Deps struct {
	Store  contract.ProjectStore
	Notify NotifyFunc
}

// Handler serves project REST routes.
type Handler struct {
	store  contract.ProjectStore
	notify NotifyFunc
}

// Register mounts /projects* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{store: deps.Store, notify: deps.Notify}
	m.Handle("GET /projects", http.HandlerFunc(h.listProjects))
	m.Handle("POST /projects", http.HandlerFunc(h.createProject))
	m.Handle("GET /projects/{id}", http.HandlerFunc(h.getProject))
	m.Handle("PATCH /projects/{id}", http.HandlerFunc(h.patchProject))
	m.Handle("DELETE /projects/{id}", http.HandlerFunc(h.deleteProject))
	m.Handle("GET /projects/{id}/context", http.HandlerFunc(h.listProjectContext))
	m.Handle("POST /projects/{id}/context", http.HandlerFunc(h.createProjectContext))
	m.Handle("POST /projects/{id}/context/edges", http.HandlerFunc(h.createProjectContextEdge))
	m.Handle("PATCH /projects/{id}/context/edges/{edgeId}", http.HandlerFunc(h.patchProjectContextEdge))
	m.Handle("DELETE /projects/{id}/context/edges/{edgeId}", http.HandlerFunc(h.deleteProjectContextEdge))
	m.Handle("PATCH /projects/{id}/context/{contextId}", http.HandlerFunc(h.patchProjectContext))
	m.Handle("DELETE /projects/{id}/context/{contextId}", http.HandlerFunc(h.deleteProjectContext))
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) notifyChange(typ realtime.ChangeType, id string) {
	if h.notify == nil || id == "" {
		return
	}
	h.notify(typ, id)
}
