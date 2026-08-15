package taskapi

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	draftassistcontract "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// DraftAssistReady is the optional /draft-assist/ready probe. The sidecar
// supervisor and boot-time static probes implement it.
type DraftAssistReady interface {
	Ready() (ready bool, runner, reason string)
}

// DraftAssistHost is the already-started draft-assist runner the HTTP
// stack mounts. Production hosts pass draftsidecar.MustHost(); tests
// inject Fake so CI does not spawn hamix-draft-agent.
type DraftAssistHost struct {
	Runner draftassistcontract.Runner
	Ready  DraftAssistReady
	Close  func() error
}

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
//
// da must already own the runner lifetime. NewHTTPHandler never looks
// up hamix-draft-agent or constructs a Fake runner.
func NewHTTPHandler(s *composition.API, hub *realtime.SSEHub, rep *repo.Root, agent handler.AgentWorkerControl, drift postgres.SchemaDriftReport, da DraftAssistHost) (http.Handler, func() error) {
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
	if da.Runner != nil {
		opts = append(opts, handler.WithDraftAssist(draftassiststore.NewMemoryStore(), da.Runner))
		if da.Ready != nil {
			opts = append(opts, handler.WithDraftAssistReady(da.Ready))
		}
	}
	closer := da.Close
	if closer == nil {
		closer = func() error { return nil }
	}
	return middleware.Stack(handler.NewHandler(s, hub, rep, opts...), calltrace.Path), closer
}
