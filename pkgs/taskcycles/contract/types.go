package contract

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// StartCycleInput captures everything needed to begin a new execution attempt.
type StartCycleInput struct {
	TaskID        string
	TriggeredBy   taskcoredomain.Actor
	ParentCycleID *string
	Meta          []byte
}

// CompletePhaseInput captures the terminal transition for a phase row.
type CompletePhaseInput struct {
	CycleID  string
	PhaseSeq int64
	Status   cyclesdomain.PhaseStatus
	Summary  *string
	Details  []byte
	By       taskcoredomain.Actor
}
