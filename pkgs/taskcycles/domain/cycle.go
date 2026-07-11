package domain

import (
	"encoding/json"
	"time"
)

// TaskCycle is one execution attempt for a task. The (TaskID, AttemptSeq) pair
// gives a stable monotonic ordering of attempts. A cycle's lifecycle is enforced
// at the store boundary: at most one Running cycle per task at any time, and
// terminal statuses (Succeeded / Failed / Aborted) are immutable. See
// docs/data-model.md.
type TaskCycle struct {
	ID            string          `json:"id"`
	TaskID        string          `json:"task_id"`
	AttemptSeq    int64           `json:"attempt_seq"`
	Status        CycleStatus     `json:"status"`
	StartedAt     time.Time       `json:"started_at"`
	EndedAt       *time.Time      `json:"ended_at,omitempty"`
	TriggeredBy   string          `json:"triggered_by"` // Actor wire value: "user" | "agent"
	ParentCycleID *string         `json:"parent_cycle_id,omitempty"`
	MetaJSON      json.RawMessage `json:"meta_json"`
}

// TaskCyclePhase is one phase entry within a cycle. A single cycle can have
// multiple rows for the same Phase value (for example a corrective Verify after
// a second Execute), so PhaseSeq is the monotonic entry-order identity within
// a cycle, while Phase is the phase kind. Lifecycle invariants (one Running
// phase per cycle, terminal status immutable, transitions validated by
// ValidPhaseTransition) live at the store boundary.
type TaskCyclePhase struct {
	ID          string          `json:"id"`
	CycleID     string          `json:"cycle_id"`
	Phase       Phase           `json:"phase"`
	PhaseSeq    int64           `json:"phase_seq"`
	Status      PhaseStatus     `json:"status"`
	StartedAt   time.Time       `json:"started_at"`
	EndedAt     *time.Time      `json:"ended_at,omitempty"`
	Summary     *string         `json:"summary,omitempty"`
	DetailsJSON json.RawMessage `json:"details_json"`
	// EventSeq points at the task_events row that mirrors the most recent
	// transition for this phase (set in the same SQL transaction as the mirror
	// insert). Nullable because it is filled in by the store, not by the caller.
	EventSeq *int64 `json:"event_seq,omitempty"`
}

// TaskCycleStreamEvent is a durable, per-attempt record of normalized runner
// progress. It is intentionally separate from TaskEvent so high-volume tool
// streams do not pollute the human-scale task audit timeline.
type TaskCycleStreamEvent struct {
	ID          string          `json:"id"`
	TaskID      string          `json:"task_id"`
	CycleID     string          `json:"cycle_id"`
	PhaseSeq    int64           `json:"phase_seq"`
	StreamSeq   int64           `json:"stream_seq"`
	At          time.Time       `json:"at"`
	Source      string          `json:"source"`
	Kind        string          `json:"kind"`
	Subtype     string          `json:"subtype"`
	Message     string          `json:"message"`
	Tool        string          `json:"tool"`
	PayloadJSON json.RawMessage `json:"payload_json"`
}
