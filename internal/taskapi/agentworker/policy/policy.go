package policy

import (
	"context"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

// SchedulingIdleHintReason is the diagnostic idle reason emitted when
// the worker is fully configured and could run, but the ready queue
// is empty only because every ready task is deferred via
// pickup_not_before > now. Not returned by DecideIdle — see
// docs/domain/agent-supervisor.md.
const SchedulingIdleHintReason = "awaiting_scheduled_task"

// GitRegistrationChecker validates git repository registration before the worker runs.
type GitRegistrationChecker func(ctx context.Context) (idle bool, reason string, err error)

// DecideSchedulingIdleHint reports the diagnostic hint when the queue
// is empty but scheduled tasks exist. Errors from probes degrade to "".
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideSchedulingIdleHint(queueEmpty bool, scheduledCount int64) string {
	if queueEmpty && scheduledCount > 0 {
		return SchedulingIdleHintReason
	}
	return ""
}

// DecideIdle reports whether the worker should stay idle given settings.
// checkGit inspects registered git repositories and worktree paths.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideIdle(ctx context.Context, cfg settingsdomain.AppSettings, checkGit GitRegistrationChecker) (idle bool, reason string) {
	if cfg.AgentPaused {
		return true, "paused_by_operator"
	}
	if checkGit == nil {
		return false, ""
	}
	idle, reason, err := checkGit(ctx)
	if err != nil {
		return true, "git_registration_check_failed"
	}
	return idle, reason
}

// InstanceSnapshot captures the running worker state needed for
// material-change comparison without importing supervisor types.
type InstanceSnapshot struct {
	Settings      settingsdomain.AppSettings
	RunnerVersion string
}

// InstanceMatchesSettings reports whether the running worker already
// matches desired settings and probed runner version.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func InstanceMatchesSettings(inst *InstanceSnapshot, cfg settingsdomain.AppSettings, version string) bool {
	if inst == nil {
		return false
	}
	if inst.Settings.Runner != cfg.Runner {
		return false
	}
	if inst.Settings.CursorBin != cfg.CursorBin {
		return false
	}
	if inst.Settings.CursorModel != cfg.CursorModel {
		return false
	}
	if inst.Settings.MaxRunDurationSeconds != cfg.MaxRunDurationSeconds {
		return false
	}
	if inst.Settings.AgentPaused != cfg.AgentPaused {
		return false
	}
	if inst.RunnerVersion != "" && inst.RunnerVersion != version {
		return false
	}
	return true
}
