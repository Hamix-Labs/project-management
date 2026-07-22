package execute

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// RunPlan is the cursor-resume plan for one execute runner.Run.
type RunPlan struct {
	Mode            string
	ResumeSessionID string
	Prompt          string
}

// PhasePorts injects harness-owned start/plan/invoke seams so execute stays
// free of package harness and processState.
type PhasePorts struct {
	StartPhase            func(ctx context.Context, cycle *cyclesdomain.TaskCycle) (*cyclesdomain.TaskCyclePhase, bool)
	PlanRun               func(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle) (RunPlan, error)
	PlanFallback          func(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle) RunPlan
	Invoke                func(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle, exec *cyclesdomain.TaskCyclePhase, plan RunPlan) (runner.Result, error)
	ConsumeOperatorCancel func() bool
	Publish               func(taskID, cycleID string)
	WorkingDir            string
	ReportDir             string
	RepoRoot              string
	IsFreshOrFallback     func(mode string) bool
}

// PhaseResult is the I/O outcome of one execute phase before Decide/Apply.
type PhaseResult struct {
	// FatalReason, when non-empty, means the harness should terminate the cycle
	// with this reason and stop the loop (pipeline aborted before Decide).
	FatalReason string

	ExecPhase         *cyclesdomain.TaskCyclePhase
	Result            runner.Result
	RunErr            error
	Snap              git.PhaseSnapshot
	IngestAttempted   bool
	IngestOutcome     git.ExecuteCommitIngestOutcome
	IngestErr         error
	OperatorCancelled bool
	CommitCount       int
	StaleRecovery     bool
	PostRunInput      orchestration.ExecutePostRunInput
}
