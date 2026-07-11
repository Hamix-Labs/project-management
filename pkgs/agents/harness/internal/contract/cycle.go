package contract

import (
	"context"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

// CycleStore covers harness execution cycles, phases, stream events, and indexed artifacts.
type CycleStore interface {
	StartCycle(ctx context.Context, in cyclescontract.StartCycleInput) (*cyclesdomain.TaskCycle, error)
	GetCycle(ctx context.Context, cycleID string) (*cyclesdomain.TaskCycle, error)
	ListCyclesForTask(ctx context.Context, taskID string, limit int) ([]cyclesdomain.TaskCycle, error)
	TerminateCycle(ctx context.Context, cycleID string, status cyclesdomain.CycleStatus, reason string, by taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error)

	StartPhase(ctx context.Context, cycleID string, phase cyclesdomain.Phase, by taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error)
	CompletePhase(ctx context.Context, in cyclescontract.CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error)
	ListPhasesForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCyclePhase, error)
	LastSessionID(ctx context.Context, cycleID string, phase cyclesdomain.Phase) (string, error)

	ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCriteriaReport, error)
	UpsertCriteriaReports(ctx context.Context, cycleID string, attemptSeq int64, entries []cyclesstore.CriteriaReportEntry) error
	ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleVerifyReport, error)
	UpsertVerifyReports(ctx context.Context, cycleID string, attemptSeq int64, entries []cyclesstore.VerifyReportEntry) error
	UpsertCommandRuns(ctx context.Context, cycleID string, attemptSeq int64, entries []cyclesstore.CommandRunEntry) error

	ListCommitsForTask(ctx context.Context, taskID string) ([]cyclesdomain.TaskCycleCommit, error)
	ListCommitsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommit, error)
	UpsertCycleCommits(ctx context.Context, taskID, cycleID string, entries []cyclesstore.CycleCommitEntry) error

	AppendCycleStreamEvent(ctx context.Context, in cyclesstore.AppendCycleStreamEventInput) (*cyclesdomain.TaskCycleStreamEvent, error)
	ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]cyclesdomain.TaskCycleStreamEvent, error)
}
