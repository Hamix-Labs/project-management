package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// CycleStore covers execution cycles, phases, stream events, and indexed artifacts.
type CycleStore interface {
	StartCycle(ctx context.Context, in StartCycleInput) (*domain.TaskCycle, error)
	GetCycle(ctx context.Context, cycleID string) (*domain.TaskCycle, error)
	ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]domain.TaskCycle, error)
	TerminateCycle(ctx context.Context, cycleID string, status domain.CycleStatus, reason string, by domain.Actor) (*domain.TaskCycle, error)
	StartPhase(ctx context.Context, cycleID string, phase domain.Phase, by domain.Actor) (*domain.TaskCyclePhase, error)
	CompletePhase(ctx context.Context, in CompletePhaseInput) (*domain.TaskCyclePhase, error)
	ListPhasesForCycle(ctx context.Context, cycleID string) ([]domain.TaskCyclePhase, error)
	ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]domain.TaskCycleStreamEvent, error)
	ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCriteriaReport, error)
	ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleVerifyReport, error)
	ListCommandRunsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommandRun, error)
	ListCommitsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommit, error)
	ListCommitsForTask(ctx context.Context, taskID string) ([]domain.TaskCycleCommit, error)
}
