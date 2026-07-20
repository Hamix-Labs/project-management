package taskapi

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"

	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
)

// NewHTTPHandler returns the REST + SSE task API with the standard middleware stack
// (see pkgs/tasks/middleware.Stack) wrapping handler.NewHandler.
//
// rep is the legacy static workspace; pass nil in production wiring
// to delegate to the settings-backed RepoProvider built inside
// (which makes /repo/* + prompt mention validation follow
// AppSettings.RepoRoot live, as required by docs/configuration.md).
// Tests that need a fixed tmpdir can still pass a non-nil rep.
//
// Pass a nil agent control to opt out of the supervisor-aware
// /settings sub-routes (PATCH /settings, POST /settings/probe-cursor,
// POST /settings/cancel-current-run); GET /settings still works.
func NewHTTPHandler(s *composition.API, hub *realtime.SSEHub, rep *repo.Root, agent handler.AgentWorkerControl, drift postgres.SchemaDriftReport) http.Handler {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "internal.taskapi.NewHTTPHandler")
	opts := []handler.HandlerOption{
		handler.WithSchemaDriftReport(drift),
	}
	if agent != nil {
		opts = append(opts, handler.WithAgentWorkerControl(agent))
	}
	if rep == nil {
		opts = append(opts, handler.WithRepoProvider(handler.NewSettingsRepoProvider(s)))
	}
	return middleware.Stack(handler.NewHandler(s, hub, rep, opts...), calltrace.Path)
}
