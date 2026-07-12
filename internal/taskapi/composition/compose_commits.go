package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

// UpsertCycleCommits persists worker-indexed git commits for one cycle.
func (a *API) UpsertCycleCommits(ctx context.Context, taskID, cycleID string, entries []cyclesstore.CycleCommitEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertCycleCommits")
	return a.cycles.UpsertCycleCommits(ctx, taskID, cycleID, entries)
}

// ListCommitsForCycle returns commits for a cycle ordered by ancestry seq.
func (a *API) ListCommitsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCommitsForCycle")
	return a.cycles.ListCommitsForCycle(ctx, cycleID)
}

// ListCommitsForTask returns distinct commits indexed for a task across all cycles.
func (a *API) ListCommitsForTask(ctx context.Context, taskID string) ([]cyclesdomain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCommitsForTask")
	return a.cycles.ListCommitsForTask(ctx, taskID)
}
