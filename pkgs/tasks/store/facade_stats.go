package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

type (
	TaskStats               = taskcorestore.TaskStats
	CycleStats              = taskcorestore.CycleStats
	PhaseStats              = taskcorestore.PhaseStats
	RunnerStats             = taskcorestore.RunnerStats
	RunnerBucket            = taskcorestore.RunnerBucket
	RecentFailure           = taskcorestore.RecentFailure
	ListCycleFailuresInput  = taskcorestore.ListCycleFailuresInput
	ListCycleFailuresResult = taskcorestore.ListCycleFailuresResult
	PreFeatureCycleCounts   = taskcorestore.PreFeatureCycleCounts
)

const (
	CycleFailureSortAtDesc     = taskcorestore.CycleFailureSortAtDesc
	CycleFailureSortAtAsc      = taskcorestore.CycleFailureSortAtAsc
	CycleFailureSortReasonAsc  = taskcorestore.CycleFailureSortReasonAsc
	CycleFailureSortReasonDesc = taskcorestore.CycleFailureSortReasonDesc
	RunnerUnknownKey           = taskcorestore.RunnerUnknownKey
)

func (s *Store) TaskStats(ctx context.Context) (TaskStats, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.TaskStats")
	return s.taskcore.TaskStats(ctx)
}

func (s *Store) ListCycleFailures(ctx context.Context, in ListCycleFailuresInput) (ListCycleFailuresResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCycleFailures")
	return s.taskcore.ListCycleFailures(ctx, in)
}

func (s *Store) CountPreFeatureCycles(ctx context.Context) (PreFeatureCycleCounts, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CountPreFeatureCycles")
	return s.taskcore.CountPreFeatureCycles(ctx)
}
