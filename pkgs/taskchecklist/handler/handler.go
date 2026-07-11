// Package handler registers /tasks/{id}/checklist* REST routes for taskapi.
package handler

import (
	"context"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
)

// NotifyTaskUpdatedFunc publishes an enriched task_updated SSE frame after checklist mutations.
type NotifyTaskUpdatedFunc func(ctx context.Context, taskID string) error

// Deps wires checklist HTTP handlers into the taskapi mux.
type Deps struct {
	Checklist         contract.ChecklistStore
	NotifyTaskUpdated NotifyTaskUpdatedFunc
}

// Handler serves task checklist REST routes.
type Handler struct {
	checklist         contract.ChecklistStore
	notifyTaskUpdated NotifyTaskUpdatedFunc
}

// Register mounts /tasks/{id}/checklist* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{
		checklist:         deps.Checklist,
		notifyTaskUpdated: deps.NotifyTaskUpdated,
	}
	m.Handle("GET /tasks/{id}/checklist", http.HandlerFunc(h.getChecklist))
	m.Handle("POST /tasks/{id}/checklist/items", http.HandlerFunc(h.postChecklistItem))
	m.Handle("PATCH /tasks/{id}/checklist/items/{itemId}", http.HandlerFunc(h.patchChecklistItem))
	m.Handle("DELETE /tasks/{id}/checklist/items/{itemId}", http.HandlerFunc(h.deleteChecklistItem))
}
