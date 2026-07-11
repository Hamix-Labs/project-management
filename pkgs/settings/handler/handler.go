// Package handler registers /settings* REST routes for taskapi.
package handler

import (
	"context"
	"net/http"
	"time"

	gitcontract "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// AgentWorkerControl is the narrow surface settings handlers use to drive
// the in-process agent worker without importing pkgs/tasks/handler.
type AgentWorkerControl interface {
	CancelCurrentRun() bool
	Reload(ctx context.Context) error
	ProbeRunner(ctx context.Context, runnerID, binaryPath string, timeout time.Duration) (version, resolvedBin string, err error)
}

// NotifyFunc publishes a scopeless SSE change event (settings_changed, agent_run_cancelled).
type NotifyFunc func(typ realtime.ChangeType)

// Deps wires settings HTTP handlers into the taskapi mux.
type Deps struct {
	Settings contract.SettingsStore
	GitRead  gitcontract.GitReadStore
	Agent    AgentWorkerControl
	Git      gitwork.Service
	Notify   NotifyFunc
}

// Handler serves /settings* REST routes.
type Handler struct {
	settings contract.SettingsStore
	gitRead  gitcontract.GitReadStore
	agent    AgentWorkerControl
	git      gitwork.Service
	notify   NotifyFunc
}

// Register mounts /settings* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	gitSvc := deps.Git
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	h := &Handler{
		settings: deps.Settings,
		gitRead:  deps.GitRead,
		agent:    deps.Agent,
		git:      gitSvc,
		notify:   deps.Notify,
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
	if h.git != nil {
		return h.git
	}
	return gitwork.New()
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyChange(typ realtime.ChangeType) {
	if h.notify == nil {
		return
	}
	h.notify(typ)
}
