package contract

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"

// StartCycleInput captures everything needed to begin a new execution attempt.
type StartCycleInput struct {
	TaskID        string
	TriggeredBy   domain.Actor
	ParentCycleID *string
	Meta          []byte
}

// CompletePhaseInput captures the terminal transition for a phase row.
type CompletePhaseInput struct {
	CycleID  string
	PhaseSeq int64
	Status   domain.PhaseStatus
	Summary  *string
	Details  []byte
	By       domain.Actor
}
