package taskapi

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"

	"context"
	"log/slog"
	"net/http"
	"os"
	"os/exec"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/draftsidecar"
	draftassistcontract "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	draftassistmetrics "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/metrics"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/runner"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
)

// staticReadyProbe implements handler.ReadyProbe for boot-time fallback
// paths that cannot recover without operator action (missing binary or
// missing API key). It reports ready=false with a stable reason so the
// SPA banner can show actionable copy.
type staticReadyProbe struct {
	runner string
	reason string
}

// Ready implements handler.ReadyProbe (and draftassist/handler.ReadyProbe).
//
//funclogmeasure:skip category=hot-path reason="Static ready-probe accessor; /draft-assist/ready already traces."
func (s staticReadyProbe) Ready() (bool, string, string) {
	return false, s.runner, s.reason
}

// draftAssistRunnerSelection returns the runner + ready probe pair to
// wire into the handler. The chosen path depends on:
//
//   - hamix-draft-agent on PATH: no → fake + no_runner.
//   - hamix-draft-agent on PATH, CURSOR_API_KEY unset: → fake + missing_key.
//   - Both present: → sidecar SDK runner + supervisor.Ready() probe.
//
// The supervisor spawn error also degrades to fake + sidecar_down; the
// SPA banner then invites the operator to retry.
//
//funclogmeasure:skip category=tool-required-noop reason="Boot-time runner picker; each branch emits a decision log."
func draftAssistRunnerSelection() (draftassistcontract.Runner, handler.HandlerOption, func() error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.draftAssistRunnerSelection")
	fake := runner.NewFake(runner.FakeOptions{})
	noopCloser := func() error { return nil }

	binPath, lookErr := exec.LookPath(draftsidecar.BinaryName)
	if lookErr != nil {
		slog.Info("draft-assist runner=fake reason=no_runner",
			"cmd", calltrace.LogCmd,
			"operation", "taskapi.draftAssistRunnerSelection",
			"binary", draftsidecar.BinaryName,
		)
		return fake, handler.WithDraftAssistReady(staticReadyProbe{
			runner: "missing",
			reason: draftassistmetrics.ReasonNoRunner,
		}), noopCloser
	}

	if os.Getenv(draftsidecar.APIKeyEnv) == "" {
		slog.Info("draft-assist runner=fake reason=missing_key",
			"cmd", calltrace.LogCmd,
			"operation", "taskapi.draftAssistRunnerSelection",
			"binary", binPath,
		)
		return fake, handler.WithDraftAssistReady(staticReadyProbe{
			runner: "sdk",
			reason: draftassistmetrics.ReasonMissingKey,
		}), noopCloser
	}

	sup := draftsidecar.NewSupervisor(draftsidecar.Options{
		BinaryPath: binPath,
		Stderr:     os.Stderr,
	})
	if err := sup.Start(context.Background()); err != nil {
		slog.Warn("draft-assist supervisor failed to start; falling back to fake",
			"cmd", calltrace.LogCmd,
			"operation", "taskapi.draftAssistRunnerSelection",
			"err", err,
		)
		return fake, handler.WithDraftAssistReady(staticReadyProbe{
			runner: "sdk",
			reason: draftassistmetrics.ReasonSidecarDown,
		}), noopCloser
	}
	slog.Info("draft-assist runner=sdk",
		"cmd", calltrace.LogCmd,
		"operation", "taskapi.draftAssistRunnerSelection",
		"binary", binPath,
		"port", sup.Port(),
	)
	sdk := draftsidecar.NewRunner(draftsidecar.RunnerOptions{
		PortSource: sup,
	})
	return sdk, handler.WithDraftAssistReady(sup), sup.Close
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
func NewHTTPHandler(s *composition.API, hub *realtime.SSEHub, rep *repo.Root, agent handler.AgentWorkerControl, drift postgres.SchemaDriftReport) (http.Handler, func() error) {
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
	// Draft-assist (ADR-0101): pick the SDK-backed sidecar when the
	// binary + CURSOR_API_KEY are present; else keep the in-process fake
	// runner so CI and offline dev keep working.
	draftRunner, readyOpt, closer := draftAssistRunnerSelection()
	opts = append(opts,
		handler.WithDraftAssist(draftassiststore.NewMemoryStore(), draftRunner),
		readyOpt,
	)
	return middleware.Stack(handler.NewHandler(s, hub, rep, opts...), calltrace.Path), closer
}
