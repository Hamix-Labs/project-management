package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// CycleCommitEntry is the public re-export of a commit upsert payload.
type CycleCommitEntry = cyclesstore.CycleCommitEntry

// UpsertCycleCommits persists worker-indexed git commits for one cycle.
func (s *Store) UpsertCycleCommits(ctx context.Context, taskID, cycleID string, entries []CycleCommitEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertCycleCommits")
	return s.cycles.UpsertCycleCommits(ctx, taskID, cycleID, entries)
}

// ListCommitsForCycle returns commits for a cycle ordered by ancestry seq.
func (s *Store) ListCommitsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCommitsForCycle")
	return s.cycles.ListCommitsForCycle(ctx, cycleID)
}

// ListCommitsForTask returns distinct commits indexed for a task across all cycles.
func (s *Store) ListCommitsForTask(ctx context.Context, taskID string) ([]domain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCommitsForTask")
	return s.cycles.ListCommitsForTask(ctx, taskID)
}
