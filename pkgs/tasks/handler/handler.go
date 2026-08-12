package handler

import (
	"log/slog"
	"net/http"
	"os/exec"

	draftassistcontract "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	composehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// Task routes: see README.md (handler_task_*.go). /repo: pkgs/repo/handler. SSE: sse.go.
// Settings routes: handler_settings.go (GET/PATCH /settings, POST /settings/probe-cursor,
// POST /settings/list-cursor-models, POST /settings/cancel-current-run).
// Runner routes: pkgs/runners/handler (GET /runners, GET /runners/{id}/config-schema,
// POST /runners/{id}/probe, POST /runners/{id}/list-models,
// POST /runners/{id}/validate-config).

// AgentWorkerControl aliases the shared settings/contract surface for
// composition and tests that still reference handler.AgentWorkerControl.
type AgentWorkerControl = settingscontract.AgentWorkerControl

// Handler carries dependencies for the mounted REST routes, SSE stream, repo
// helpers, and optional agent worker control. Use NewHandler; the zero value
// is not usable.
type Handler struct {
	store          HandlerStore
	hub            *realtime.SSEHub
	repoProv       RepoProvider
	agent          AgentWorkerControl
	systemHealthFn systemHealthSnapshotter
	git            gitwork.Service
	gitAvailable   bool
	schemaDrift    postgres.SchemaDriftReport

	enqueueInstantiate composehandler.EnqueueInstantiateFunc
	instantiateWorker  *instantiateWorker

	draftAssistStore  draftassistcontract.Store
	draftAssistRunner draftassistcontract.Runner
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
func NewHandler(s HandlerStore, hub *realtime.SSEHub, rep *repo.Root, opts ...HandlerOption) http.Handler {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.NewHandler")
	_, gitErr := exec.LookPath("git")
	h := &Handler{
		store:        s,
		hub:          hub,
		repoProv:     NewStaticRepoProvider(rep),
		git:          gitwork.New(),
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

// WithGitAvailable overrides the git binary probe (tests only).
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HandlerOption wiring; no production operation boundary."
func WithGitAvailable(ok bool) HandlerOption {
	return func(h *Handler) {
		h.gitAvailable = ok
	}
}

// WithDraftAssist wires the in-memory draft-assist store and runner
// (ADR-0101). When omitted, /draft-assist/* routes are not registered.
func WithDraftAssist(store draftassistcontract.Store, runner draftassistcontract.Runner) HandlerOption {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.WithDraftAssist")
	return func(h *Handler) {
		h.draftAssistStore = store
		h.draftAssistRunner = runner
	}
}
