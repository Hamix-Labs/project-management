package verify

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// Hooks wires harness-owned side effects (SSE, metrics, progress persistence).
type Hooks struct {
	Publish         func(taskID, cycleID string)
	PersistProgress func(ctx context.Context, taskID, cycleID string, phaseSeq int64, ev runner.ProgressEvent)
	RecordVerdict   func(kind checklistdomain.VerifierKind, passed bool)
	ObserveDuration func(d time.Duration)
	// SetRunCancel registers or clears the in-flight verify cursor cancel func.
	SetRunCancel    func(cancel context.CancelFunc)
	StreamIdleStuck time.Duration
	OnStreamIdle    func(kind runner.StreamIdleKind)
	// PlanVerifyRun selects prompt + cursor resume fields before verify runner.Run.
	PlanVerifyRun func(ctx context.Context, in PlanVerifyRunInput) (VerifyRunPlan, error)
	// OnVerifyPhaseEnded is called after verify phase row closes (success or failure).
	OnVerifyPhaseEnded func(executePhaseSeq int64)
}

// PlanVerifyRunInput carries verify context into harness resume policy.
type PlanVerifyRunInput struct {
	Task             *taskcoredomain.Task
	Cycle            *cyclesdomain.TaskCycle
	Snap             Snapshot
	VerifyAttempt    int
	Feedback         string
	CmdEvidence      []CommandEvidence
	SelfReport       map[string]reports.CriteriaEntry
	PreviouslyPassed map[string]Verdict
}

// VerifyRunPlan is the cursor resume plan for one verify runner.Run.
type VerifyRunPlan struct {
	Prompt           string
	ResumeSessionID  string
	CursorResumeMode string
	RecoveryKind     string
}
