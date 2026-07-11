package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os/exec"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
)

// Task routes: see README.md (handler_task_*.go). /repo: pkgs/repo/handler. SSE: sse.go.
// Settings routes: handler_settings.go (GET/PATCH /settings, POST /settings/probe-cursor,
// POST /settings/list-cursor-models, POST /settings/cancel-current-run).
// Runner routes: handler_runners.go (GET /runners, GET /runners/{id}/config-schema,
// POST /runners/{id}/probe, POST /runners/{id}/list-models,
// POST /runners/{id}/validate-config).

// AgentWorkerControl is the narrow surface the /settings handlers use
// to drive the in-process agent worker. The cmd/taskapi supervisor
// implements it; tests can stub it out (or pass nil to disable the
// supervisor-aware endpoints — they then return 503).
//
// Reload is invoked after PATCH /settings persists so the worker
// picks up the new config without a process restart. CancelCurrentRun
// is the explicit "stop the runaway run" knob exposed at
// POST /settings/cancel-current-run; it returns true when there was
// an in-flight run to cancel. ProbeRunner is invoked from POST
// /settings/probe-cursor so the SPA can validate a binary path
// against the configured runner before saving.
type AgentWorkerControl interface {
	CancelCurrentRun() bool
	Reload(ctx context.Context) error
	ProbeRunner(ctx context.Context, runnerID, binaryPath string, timeout time.Duration) (version, resolvedBin string, err error)
}

// Handler carries dependencies for the mounted REST routes, SSE stream, repo
// helpers, and optional agent worker control. Use NewHandler; the zero value
// is not usable.
type Handler struct {
	store          contract.HandlerStore
	hub            *SSEHub
	repoProv       RepoProvider
	agent          AgentWorkerControl
	systemHealthFn systemHealthSnapshotter
	git            gitwork.Service
	pathMap        *PathMap
	gitAvailable   bool
	schemaDrift    postgres.SchemaDriftReport
}

// NewHandler returns the task REST API and GET /events (SSE) when hub is non-nil.
//
// rep is the legacy static workspace root: pass nil to disable /repo
// routes (they return 409 repo_root_not_configured) or pre-open one
// for tests that want a fixed tmpdir. The production wiring should
// instead pass nil here and call WithRepoProvider with a settings-
// backed provider so the repo follows AppSettings.RepoRoot live.
//
// agent is optional: when nil, settings-control endpoints (PATCH /settings,
// POST /settings/probe-cursor, POST /settings/cancel-current-run) respond 503.
// GET /settings still works without it (read-only).
func NewHandler(s contract.HandlerStore, hub *SSEHub, rep *repo.Root, opts ...HandlerOption) http.Handler {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.NewHandler")
	_, gitErr := exec.LookPath("git")
	h := &Handler{
		store:        s,
		hub:          hub,
		repoProv:     NewStaticRepoProvider(rep),
		git:          gitwork.New(),
		pathMap:      &PathMap{},
		gitAvailable: gitErr == nil,
	}
	for _, opt := range opts {
		opt(h)
	}
	m := http.NewServeMux()
	h.registerRoutes(m)
	return m
}

// HandlerOption configures the Handler at construction time. Optional
// because most callers (tests, embedding) only need the core surface.
type HandlerOption func(*Handler)

// WithAgentWorkerControl wires the supervisor that owns the in-process
// agent worker so PATCH /settings can hot-reload, POST
// /settings/probe-cursor can probe the runner, and POST
// /settings/cancel-current-run can cancel an in-flight run. Pass nil
// (or omit the option) to disable those endpoints — they then return
// 503 service_unavailable and GET /settings still works.
func WithAgentWorkerControl(c AgentWorkerControl) HandlerOption {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.WithAgentWorkerControl")
	return func(h *Handler) {
		h.agent = c
	}
}

// WithRepoProvider replaces the default static repo wiring with a
// dynamic provider. cmd/taskapi passes a NewSettingsRepoProvider so
// /repo/* and prompt-mention validation always look at the current
// AppSettings.RepoRoot; tests rarely need this option (the rep
// argument to NewHandler covers the static case).
func WithRepoProvider(p RepoProvider) HandlerOption {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.WithRepoProvider")
	return func(h *Handler) {
		if p != nil {
			h.repoProv = p
		}
	}
}

// WithGitService replaces the default gitwork.New() wiring so handler
// tests can inject a stub without touching the real git binary.
func WithGitService(s gitwork.Service) HandlerOption {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.WithGitService")
	return func(h *Handler) {
		if s != nil {
			h.git = s
		}
	}
}

// WithSchemaDriftReport wires startup schema revision drift for GET /health/ready.
func WithSchemaDriftReport(r postgres.SchemaDriftReport) HandlerOption {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.WithSchemaDriftReport")
	return func(h *Handler) {
		h.schemaDrift = r
	}
}
