package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
)

// streamIdleRunnerFields returns Request.StreamIdleStuck and OnStreamIdle when
// Options.StreamIdleStuck is configured. Zero disables the watchdog.
func (h *Harness) streamIdleRunnerFields(baseOnProgress func(runner.ProgressEvent)) (time.Duration, func(runner.StreamIdleKind)) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.streamIdleRunnerFields",
		"stuck_ns", int64(h.opts.StreamIdleStuck))
	stuck := h.opts.StreamIdleStuck
	if stuck <= 0 {
		return 0, nil
	}
	return stuck, func(kind runner.StreamIdleKind) {
		ev := runner.StreamIdleProgressEvent(kind, stuck)
		if baseOnProgress != nil {
			baseOnProgress(ev)
		}
	}
}
