// Package handler registers /tasks/{id}/events* and /tasks/activity REST routes for taskapi.
package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
)

// NotifyTaskEventChangedFunc publishes a task_event_changed SSE frame after thread append.
type NotifyTaskEventChangedFunc func(taskID string, eventSeq int64)

// Deps wires task event HTTP handlers into the taskapi mux.
type Deps struct {
	Events                 contract.TaskEventStore
	Activity               contract.TaskActivityStore
	Tasks                  contract.TaskGetter
	NotifyTaskEventChanged NotifyTaskEventChangedFunc
}

// Handler serves task audit event REST routes.
type Handler struct {
	events                 contract.TaskEventStore
	activity               contract.TaskActivityStore
	tasks                  contract.TaskGetter
	notifyTaskEventChanged NotifyTaskEventChangedFunc
}

// Register mounts /tasks/{id}/events* and /tasks/activity routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{
		events:                 deps.Events,
		activity:               deps.Activity,
		tasks:                  deps.Tasks,
		notifyTaskEventChanged: deps.NotifyTaskEventChanged,
	}
	m.Handle("GET /tasks/activity", http.HandlerFunc(h.taskActivity))
	m.Handle("GET /tasks/{id}/events/{seq}", http.HandlerFunc(h.taskEvent))
	m.Handle("PATCH /tasks/{id}/events/{seq}", http.HandlerFunc(h.patchTaskEventUserResponse))
	m.Handle("GET /tasks/{id}/events", http.HandlerFunc(h.taskEvents))
}
