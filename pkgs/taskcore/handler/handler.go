// Package handler registers core /tasks* REST routes for taskapi.
package handler

import (
	"context"
	"net/http"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// SettingsReader loads app settings for task create validation.
type SettingsReader interface {
	GetSettings(ctx context.Context) (settingsdomain.AppSettings, error)
}

// ComposeGitValidator validates git/worktree bindings during task create.
type ComposeGitValidator interface {
	ValidateTaskGitBindingV2(ctx context.Context, projectID *string, worktreeID *string) error
	ValidateComposeGitBinding(ctx context.Context, repositoryID, projectID, worktreeID *string) error
	ValidatePromptMentionsForWorktree(ctx context.Context, worktreeID *string, prompt string) error
}

// NotifyChangeFunc publishes a hint-only SSE frame after task mutations.
type NotifyChangeFunc func(typ realtime.ChangeType, id string)

// NotifyTaskChangedFunc publishes an enriched task SSE frame.
type NotifyTaskChangedFunc func(typ realtime.ChangeType, id string, data any)

// Deps wires taskcore HTTP handlers into the taskapi mux.
type Deps struct {
	Tasks             contract.TaskCRUDStore
	Settings          SettingsReader
	GitCompose        ComposeGitValidator
	NotifyChange      NotifyChangeFunc
	NotifyTaskChanged NotifyTaskChangedFunc
}

// Handler serves core task CRUD, stats, gate, dependency, and retry routes.
type Handler struct {
	tasks             contract.TaskCRUDStore
	settings          SettingsReader
	gitCompose        ComposeGitValidator
	notifyChange      NotifyChangeFunc
	notifyTaskChanged NotifyTaskChangedFunc
}

// New returns a Handler wired from deps (for compose callbacks from pkgs/tasks/handler).
//
//funclogmeasure:skip category=hot-path reason="Constructor wiring only; operation trace is emitted by registered handlers."
func New(deps Deps) *Handler {
	return &Handler{
		tasks:             deps.Tasks,
		settings:          deps.Settings,
		gitCompose:        deps.GitCompose,
		notifyChange:      deps.NotifyChange,
		notifyTaskChanged: deps.NotifyTaskChanged,
	}
}

// Register mounts core /tasks* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{
		tasks:             deps.Tasks,
		settings:          deps.Settings,
		gitCompose:        deps.GitCompose,
		notifyChange:      deps.NotifyChange,
		notifyTaskChanged: deps.NotifyTaskChanged,
	}
	m.Handle("POST /tasks", http.HandlerFunc(h.create))
	m.Handle("GET /tasks", http.HandlerFunc(h.list))
	m.Handle("GET /tasks/stats", http.HandlerFunc(h.stats))
	m.Handle("GET /tasks/{id}/dependencies", http.HandlerFunc(h.listTaskDependencies))
	m.Handle("POST /tasks/{id}/dependencies", http.HandlerFunc(h.addTaskDependency))
	m.Handle("DELETE /tasks/{id}/dependencies/{depId}", http.HandlerFunc(h.removeTaskDependency))
	m.Handle("PATCH /tasks/{id}/gate", http.HandlerFunc(h.patchTaskGate))
	m.Handle("POST /tasks/{id}/retry", http.HandlerFunc(h.postTaskRetry))
	m.Handle("GET /tasks/{id}", http.HandlerFunc(h.get))
	m.Handle("PATCH /tasks/{id}", http.HandlerFunc(h.patch))
	m.Handle("DELETE /tasks/{id}", http.HandlerFunc(h.delete))
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyChangeSafe(typ realtime.ChangeType, id string) {
	if h.notifyChange != nil {
		h.notifyChange(typ, id)
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyTaskChangedSafe(typ realtime.ChangeType, id string, data any) {
	if h.notifyTaskChanged != nil {
		h.notifyTaskChanged(typ, id, data)
	}
}
