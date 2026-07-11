package domain

type Status string

const (
	StatusReady   Status = "ready"
	StatusRunning Status = "running"
	StatusBlocked Status = "blocked"
	StatusReview  Status = "review"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
	// StatusOnHold flags a task that the operator wants to keep out of
	// the agent worker's queue. Pickup is gated on Status == StatusReady
	// (see pkgs/tasks/store/internal/tasks/readiness.go), so on_hold
	// rows simply never become eligible. The user resumes the task by
	// flipping it back to StatusReady from the detail page.
	StatusOnHold Status = "on_hold"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// GateStatus is the lifecycle for a task release gate.
type GateStatus string

const (
	GateStatusLocked         GateStatus = "locked"
	GateStatusActive         GateStatus = "active"
	GateStatusPendingRelease GateStatus = "pending_release"
	GateStatusReleased       GateStatus = "released"
)
