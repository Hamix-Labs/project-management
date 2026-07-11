package domain

// Phase is one entry in a task execution cycle. The pipeline runs
// `execute → verify`, with verify → execute allowed as the corrective
// retry edge. The earlier `diagnose` (no-op skip) and `persist`
// (never reached) phases were removed once the V1 worker proved
// neither carried information the UI or audit trail needed.
type Phase string

const (
	PhaseExecute Phase = "execute"
	PhaseVerify  Phase = "verify"
)

// CycleStatus is the lifecycle state of a single task_cycles row.
type CycleStatus string

const (
	CycleStatusRunning   CycleStatus = "running"
	CycleStatusSucceeded CycleStatus = "succeeded"
	CycleStatusFailed    CycleStatus = "failed"
	CycleStatusAborted   CycleStatus = "aborted"
)

// PhaseStatus is the lifecycle state of a single task_cycle_phases row.
type PhaseStatus string

const (
	PhaseStatusRunning   PhaseStatus = "running"
	PhaseStatusSucceeded PhaseStatus = "succeeded"
	PhaseStatusFailed    PhaseStatus = "failed"
	PhaseStatusSkipped   PhaseStatus = "skipped"
)
