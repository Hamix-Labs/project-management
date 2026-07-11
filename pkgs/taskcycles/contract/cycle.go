package contract

import (
	"context"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// CycleStore covers execution cycles, phases, stream events, and indexed artifacts.
type CycleStore interface {
	StartCycle(ctx context.Context, in StartCycleInput) (*cyclesdomain.TaskCycle, error)
	GetCycle(ctx context.Context, cycleID string) (*cyclesdomain.TaskCycle, error)
	ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]cyclesdomain.TaskCycle, error)
	TerminateCycle(ctx context.Context, cycleID string, status cyclesdomain.CycleStatus, reason string, by taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error)
	StartPhase(ctx context.Context, cycleID string, phase cyclesdomain.Phase, by taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error)
	CompletePhase(ctx context.Context, in CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error)
	ListPhasesForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCyclePhase, error)
	ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]cyclesdomain.TaskCycleStreamEvent, error)
	ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCriteriaReport, error)
	ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleVerifyReport, error)
	ListCommandRunsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommandRun, error)
	ListCommitsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommit, error)
	ListCommitsForTask(ctx context.Context, taskID string) ([]cyclesdomain.TaskCycleCommit, error)
}
