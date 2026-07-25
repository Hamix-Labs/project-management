package contract

import (
	"context"
	"time"
)

// AgentWorkerControl is the narrow surface settings and tasks handlers use
// to drive the in-process agent worker. The taskapi supervisor implements
// it; tests stub it (or pass nil to disable supervisor-aware endpoints —
// they then return 503).
//
// Reload runs after PATCH /settings persists so the worker picks up config
// without a process restart. CancelCurrentRun is the explicit stop knob at
// POST /settings/cancel-current-run (true when an in-flight run was
// cancelled). ProbeRunner backs POST /settings/probe-cursor so the SPA can
// validate a binary path against the configured runner before saving.
type AgentWorkerControl interface {
	CancelCurrentRun() bool
	CancelRunForTask(taskID string) bool
	Reload(ctx context.Context) error
	ProbeRunner(ctx context.Context, runnerID, binaryPath string, timeout time.Duration) (version, resolvedBin string, err error)
}
