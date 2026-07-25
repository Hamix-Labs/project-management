package domain

import (
	"time"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

// TaskCycleCriteriaReport is the per-criterion durable record of what
// the execute agent claimed about a single criterion in one retry
// attempt. Mirrors the `criteria-report.json` side-channel file the
// agent CLI writes (see docs/data-model.md "Report file contracts")
// into the database so verdict evidence survives the cycle. The file
// is still produced and parsed by the worker — it is the agent ↔
// worker wire format — but the file is GC'd at cycle terminate; this
// row is the audit trail.
//
// (CycleID, AttemptSeq, CriterionID) is the natural read key and the
// idempotency key for the worker's bulk upsert: re-parsing the same
// report after a transient store error is safe.
//
// Cascade semantics:
//   - cycle_id: ON DELETE CASCADE — verdicts disappear with their cycle.
//   - criterion_id: ON DELETE NO ACTION — when an operator deletes a
//     checklist item, prior verdicts for it stay so historical cycles
//     remain readable. The handler returns the row even if the FK is
//     stale; the SPA renders the criterion id verbatim in that case.
type TaskCycleCriteriaReport struct {
	ID          string    `json:"id"`
	CycleID     string    `json:"cycle_id"`
	AttemptSeq  int64     `json:"attempt_seq"`
	CriterionID string    `json:"criterion_id"`
	ClaimedDone bool      `json:"claimed_done"`
	Evidence    string    `json:"evidence"`
	WrittenAt   time.Time `json:"written_at"`
}

// TaskCycleVerifyReport is the per-criterion durable record of the
// verify agent's verdict for a single criterion in one retry attempt.
// See TaskCycleCriteriaReport for cascade and idempotency rationale.
//
// VerifierKind is recorded so the SPA can distinguish a deterministic
// check pass (`deterministic_check`) from an LLM verifier pass
// (`execute_agent`) without re-parsing the workflow — same field as
// task_checklist_completions.VerifiedBy.
type TaskCycleVerifyReport struct {
	ID           string                       `json:"id"`
	CycleID      string                       `json:"cycle_id"`
	AttemptSeq   int64                        `json:"attempt_seq"`
	CriterionID  string                       `json:"criterion_id"`
	Verified     bool                         `json:"verified"`
	VerifierKind checklistdomain.VerifierKind `json:"verifier_kind"`
	Reasoning    string                       `json:"reasoning"`
	WrittenAt    time.Time                    `json:"written_at"`
}

// TaskCycleCommandRun mirrors one verify-phase command execution for a
// criterion attempt. Output bytes live in temp files referenced by
// MetaPath; this row is the durable audit trail for the SPA timeline.
type TaskCycleCommandRun struct {
	ID          string    `json:"id"`
	CycleID     string    `json:"cycle_id"`
	AttemptSeq  int64     `json:"attempt_seq"`
	CriterionID string    `json:"criterion_id"`
	CommandSeq  int64     `json:"command_seq"`
	ExitCode    int       `json:"exit_code"`
	MetaPath    string    `json:"meta_path"`
	WrittenAt   time.Time `json:"written_at"`
}

// ExecuteCriteriaReportAttemptSeq is the attempt_seq used when mirroring
// criteria-report.json at execute phase end. Verify attempts use 1..N;
// this sentinel avoids colliding with the verify retry budget.
const ExecuteCriteriaReportAttemptSeq int64 = 1_000_000
