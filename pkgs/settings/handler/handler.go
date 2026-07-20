// Package handler registers /settings* REST routes for taskapi.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gitcontract "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// NotifyFunc publishes a scopeless SSE change event (settings_changed, agent_run_cancelled).
type NotifyFunc func(typ realtime.ChangeType)

// Deps wires settings HTTP handlers into the taskapi mux.
type Deps struct {
	Settings     contract.SettingsStore
	GitInventory gitcontract.GitInventoryStore
	Agent        contract.AgentWorkerControl
	Git          gitwork.Service
	RunnerModels RunnerModelLister
	Notify       NotifyFunc
}

// RunnerModel is a runner-agnostic model choice for deprecated settings list-models.
type RunnerModel struct {
	ID    string
	Label string
}

// RunnerModelLister lists models for a runner without importing agents/runner.
type RunnerModelLister interface {
	ListModels(ctx context.Context, runnerID, binaryPath string, timeout time.Duration) (models []RunnerModel, resolvedBinary string, err error)
}

// Handler serves /settings* REST routes.
type Handler struct {
	settings     contract.SettingsStore
	gitInventory gitcontract.GitInventoryStore
	agent        contract.AgentWorkerControl
	git          gitwork.Service
	runnerModels RunnerModelLister
	notify       NotifyFunc
}

// Register mounts /settings* routes on m.
// Deps.Git must be non-nil; Register panics if it is missing so tests and
// wiring cannot accidentally hit a silent gitwork.New() default.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	if deps.Git == nil {
		panic(fmt.Errorf("settings handler: Deps.Git is required"))
	}
	h := &Handler{
		settings:     deps.Settings,
		gitInventory: deps.GitInventory,
		agent:        deps.Agent,
		git:          deps.Git,
		runnerModels: deps.RunnerModels,
		notify:       deps.Notify,
	}
	m.Handle("GET /settings", http.HandlerFunc(h.getSettings))
	m.Handle("GET /settings/workspace-roots", http.HandlerFunc(h.workspaceRoots))
	m.Handle("GET /settings/browse-dirs", http.HandlerFunc(h.browseDirs))
	m.Handle("GET /settings/git-probe", http.HandlerFunc(h.gitRepositoryProbe))
	m.Handle("PATCH /settings", http.HandlerFunc(h.patchSettings))
	m.Handle("POST /settings/probe-cursor", http.HandlerFunc(h.probeCursor))
	m.Handle("POST /settings/list-cursor-models", http.HandlerFunc(h.listCursorModels))
	m.Handle("POST /settings/cancel-current-run", http.HandlerFunc(h.cancelCurrentRun))
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitService() gitwork.Service {
	return h.git
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyChange(typ realtime.ChangeType) {
	if h.notify == nil {
		return
	}
	h.notify(typ)
}
