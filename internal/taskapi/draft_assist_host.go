package taskapi

import (
	"context"
	"log/slog"
	"os"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/draftsidecar"
	draftassistmetrics "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/metrics"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// staticReadyProbe implements DraftAssistReady for boot-time fallback
// paths that cannot recover without operator action (missing binary or
// missing API key). It reports ready=false with a stable reason so the
// SPA banner can show actionable copy.
type staticReadyProbe struct {
	runner string
	reason string
}

// Ready implements DraftAssistReady.
//
//funclogmeasure:skip category=hot-path reason="Static ready-probe accessor; /draft-assist/ready already traces."
func (s staticReadyProbe) Ready() (bool, string, string) {
	return false, s.runner, s.reason
}

// DefaultDraftAssistHost picks the SDK sidecar when ResolveBinary finds
// hamix-draft-agent and CURSOR_API_KEY is set; otherwise it keeps the
// in-process fake runner so CI and offline hosts still serve.
//
//funclogmeasure:skip category=tool-required-noop reason="Boot-time runner picker; each branch emits a decision log."
func DefaultDraftAssistHost() DraftAssistHost {
	return draftAssistRunnerSelection()
}

// draftAssistRunnerSelection returns the runner + ready probe pair to
// wire into the handler. The chosen path depends on:
//
//   - ResolveBinary miss: fake + no_runner.
//   - Binary found, CURSOR_API_KEY unset: fake + missing_key.
//   - Both present: sidecar SDK runner + supervisor.Ready() probe.
//
// The supervisor spawn error also degrades to fake + sidecar_down; the
// SPA banner then invites the operator to retry.
//
//funclogmeasure:skip category=tool-required-noop reason="Boot-time runner picker; each branch emits a decision log."
func draftAssistRunnerSelection() DraftAssistHost {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.draftAssistRunnerSelection")
	fake := runner.NewFake(runner.FakeOptions{})
	noopCloser := func() error { return nil }

	binPath, lookErr := draftsidecar.ResolveBinary()
	if lookErr != nil {
		slog.Info("draft-assist runner=fake reason=no_runner",
			"cmd", calltrace.LogCmd,
			"operation", "taskapi.draftAssistRunnerSelection",
			"binary", draftsidecar.BinaryName,
			"err", lookErr,
		)
		return DraftAssistHost{
			Runner: fake,
			Ready: staticReadyProbe{
				runner: "missing",
				reason: draftassistmetrics.ReasonNoRunner,
			},
			Close: noopCloser,
		}
	}

	if os.Getenv(draftsidecar.APIKeyEnv) == "" {
		slog.Info("draft-assist runner=fake reason=missing_key",
			"cmd", calltrace.LogCmd,
			"operation", "taskapi.draftAssistRunnerSelection",
			"binary", binPath,
		)
		return DraftAssistHost{
			Runner: fake,
			Ready: staticReadyProbe{
				runner: "sdk",
				reason: draftassistmetrics.ReasonMissingKey,
			},
			Close: noopCloser,
		}
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
		return DraftAssistHost{
			Runner: fake,
			Ready: staticReadyProbe{
				runner: "sdk",
				reason: draftassistmetrics.ReasonSidecarDown,
			},
			Close: noopCloser,
		}
	}
	slog.Info("draft-assist runner=sdk",
		"cmd", calltrace.LogCmd,
		"operation", "taskapi.draftAssistRunnerSelection",
		"binary", binPath,
		"port", sup.Port(),
	)
	return DraftAssistHost{
		Runner: draftsidecar.NewRunner(draftsidecar.RunnerOptions{
			PortSource: sup,
		}),
		Ready: sup,
		Close: sup.Close,
	}
}
