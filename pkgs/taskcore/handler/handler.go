// Package handler registers core /tasks* REST routes for taskapi.
package handler

import (
	"context"
	"net/http"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
)

// SettingsReader loads app settings for task create validation.
type SettingsReader interface {
	GetSettings(ctx context.Context) (settingsdomain.AppSettings, error)
}

// ComposeGitValidator validates git/worktree bindings during task create.
type ComposeGitValidator interface {
	ValidateTaskGitBindingV2(ctx context.Context, projectID *string, worktreeID *string) error
	ValidateComposeGitBinding(ctx context.Context, repositoryID, projectID, worktreeID *string) error
	ValidateTaskRepositoryBinding(ctx context.Context, projectID, repositoryID *string) error
	AllocateTaskWorktree(ctx context.Context, repositoryID, taskID string) (worktreeID string, err error)
	ValidatePromptMentionsForWorktree(ctx context.Context, worktreeID *string, prompt string) error
	ValidatePromptMentionsForRepository(ctx context.Context, repositoryID, prompt string) error
	ValidatePromptMentionsForProject(ctx context.Context, projectID *string, prompt string) error
}

// RunnerValidator checks that a runner id is known at the composition edge.
// Implemented outside this BC (wrapping agents/runner/registry) so taskcore
// does not import agents.
type RunnerValidator interface {
	ValidateRunner(id string) error
}

// NotifyChangeFunc aliases the contract hint-with-id SSE notify type.
type NotifyChangeFunc = contract.NotifyChangeFunc

// NotifyTaskChangedFunc aliases the contract enriched-task SSE notify type.
type NotifyTaskChangedFunc = contract.NotifyTaskChangedFunc

// Deps wires taskcore HTTP handlers into the taskapi mux.
type Deps struct {
	Tasks             contract.TaskCRUDStore
	Settings          SettingsReader
	GitCompose        ComposeGitValidator
	HTTP              contract.HTTPHelpers
	Runners           RunnerValidator
	NotifyChange      NotifyChangeFunc
	NotifyTaskChanged NotifyTaskChangedFunc
}

// Handler serves core task CRUD, stats, gate, dependency, and retry routes.
type Handler struct {
	tasks             contract.TaskCRUDStore
	settings          SettingsReader
	gitCompose        ComposeGitValidator
	httpPort          contract.HTTPHelpers
	runners           RunnerValidator
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
		httpPort:          deps.HTTP,
		runners:           deps.Runners,
		notifyChange:      deps.NotifyChange,
		notifyTaskChanged: deps.NotifyTaskChanged,
	}
}

// Register mounts core /tasks* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := New(deps)
	m.Handle("POST /tasks", http.HandlerFunc(h.create))
	m.Handle("GET /tasks", http.HandlerFunc(h.list))
	m.Handle("GET /tasks/stats", http.HandlerFunc(h.stats))
	m.Handle("GET /tasks/{id}/dependencies", http.HandlerFunc(h.listTaskDependencies))
	m.Handle("POST /tasks/{id}/dependencies", http.HandlerFunc(h.addTaskDependency))
	m.Handle("DELETE /tasks/{id}/dependencies/{depId}", http.HandlerFunc(h.removeTaskDependency))
	m.Handle("PATCH /tasks/{id}/gate", http.HandlerFunc(h.patchTaskGate))
	m.Handle("POST /tasks/{id}/retry", http.HandlerFunc(h.postTaskRetry))
	m.Handle("POST /tasks/{id}/approve", http.HandlerFunc(h.postTaskApprove))
	m.Handle("POST /tasks/{id}/polish", http.HandlerFunc(h.postTaskPolish))
	m.Handle("POST /tasks/{id}/close", http.HandlerFunc(h.postTaskClose))
	m.Handle("POST /tasks/{id}/reopen", http.HandlerFunc(h.postTaskReopen))
	m.Handle("GET /tasks/{id}", http.HandlerFunc(h.get))
	m.Handle("PATCH /tasks/{id}", http.HandlerFunc(h.patch))
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyChangeSafe(typ contract.ChangeType, id string) {
	if h.notifyChange != nil {
		h.notifyChange(typ, id)
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyTaskChangedSafe(typ contract.ChangeType, id string, data any) {
	if h.notifyTaskChanged != nil {
		h.notifyTaskChanged(typ, id, data)
	}
}
