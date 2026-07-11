// Package handler registers /tasks/{id}/cycles*, commits, and cycle-failures REST routes for taskapi.
package handler

import (
	"context"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	taskscontract "github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// NotifyCycleChangedFunc publishes a cycle_changed SSE frame after cycle mutations.
// When data is non-nil it is the enriched cycle detail payload; nil means hint-only.
type NotifyCycleChangedFunc func(ctx context.Context, taskID, cycleID string, data any)

// TaskReader is the narrow task lookup surface commits handlers use to preflight task existence.
type TaskReader interface {
	Get(ctx context.Context, id string) (*domain.Task, error)
}

// CycleFailuresStore lists paginated cycle failure mirror rows for GET /tasks/cycle-failures.
type CycleFailuresStore interface {
	ListCycleFailures(ctx context.Context, in taskscontract.ListCycleFailuresInput) (taskscontract.ListCycleFailuresResult, error)
}

// Deps wires cycle HTTP handlers into the taskapi mux.
type Deps struct {
	Cycles             contract.CycleStore
	Tasks              TaskReader
	CycleFailures      CycleFailuresStore
	NotifyCycleChanged NotifyCycleChangedFunc
}

// Handler serves execution cycle, commit, and cycle-failure REST routes.
type Handler struct {
	cycles             contract.CycleStore
	tasks              TaskReader
	failures           CycleFailuresStore
	notifyCycleChanged NotifyCycleChangedFunc
}

// Register mounts cycle, commit, and cycle-failure routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{
		cycles:             deps.Cycles,
		tasks:              deps.Tasks,
		failures:           deps.CycleFailures,
		notifyCycleChanged: deps.NotifyCycleChanged,
	}
	m.Handle("GET /tasks/cycle-failures", http.HandlerFunc(h.cycleFailures))
	m.Handle("POST /tasks/{id}/cycles", http.HandlerFunc(h.postTaskCycle))
	m.Handle("GET /tasks/{id}/cycles", http.HandlerFunc(h.getTaskCycles))
	m.Handle("GET /tasks/{id}/cycles/{cycleId}/stream", http.HandlerFunc(h.getTaskCycleStream))
	m.Handle("GET /tasks/{id}/commits", http.HandlerFunc(h.getTaskCommits))
	m.Handle("GET /tasks/{id}/cycles/{cycleId}/verdicts", http.HandlerFunc(h.getTaskCycleVerdicts))
	m.Handle("GET /tasks/{id}/cycles/{cycleId}", http.HandlerFunc(h.getTaskCycle))
	m.Handle("PATCH /tasks/{id}/cycles/{cycleId}", http.HandlerFunc(h.patchTaskCycle))
	m.Handle("POST /tasks/{id}/cycles/{cycleId}/phases", http.HandlerFunc(h.postTaskCyclePhase))
	m.Handle("PATCH /tasks/{id}/cycles/{cycleId}/phases/{phaseSeq}", http.HandlerFunc(h.patchTaskCyclePhase))
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyCycleChangedFromStore(ctx context.Context, taskID, cycleID string) {
	if h.notifyCycleChanged == nil || taskID == "" || cycleID == "" {
		return
	}
	cycle, err := h.cycles.GetCycle(ctx, cycleID)
	if err != nil {
		h.notifyCycleChanged(ctx, taskID, cycleID, nil)
		return
	}
	phases, err := h.cycles.ListPhasesForCycle(ctx, cycleID)
	if err != nil {
		h.notifyCycleChanged(ctx, taskID, cycleID, nil)
		return
	}
	h.notifyCycleChanged(ctx, taskID, cycleID, taskCycleDetailFromDomain(cycle, phases))
}
