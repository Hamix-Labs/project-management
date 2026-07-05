package cycles

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
)

// StartCycleInput captures everything needed to begin a new execution
// attempt for a task. The store decides AttemptSeq; callers cannot
// supply it. Re-aliased by the public store facade as
// store.StartCycleInput so handler code stays unchanged.
type StartCycleInput = contract.StartCycleInput

// CompletePhaseInput captures the terminal transition for a phase row,
// keyed by (cycleID, phaseSeq) so the URL-level identifier from
// /cycles/{cycleId}/phases/{phaseSeq} is also the natural store key.
// Re-aliased by the public store facade as store.CompletePhaseInput.
type CompletePhaseInput = contract.CompletePhaseInput

// AppendStreamEventInput captures one durable normalized progress event for
// a cycle attempt. The store assigns StreamSeq per cycle.
type AppendStreamEventInput struct {
	TaskID   string
	CycleID  string
	PhaseSeq int64
	At       time.Time
	Source   string
	Kind     string
	Subtype  string
	Message  string
	Tool     string
	Payload  []byte
}
