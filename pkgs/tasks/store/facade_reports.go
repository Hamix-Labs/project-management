package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

type (
	// CriteriaReportEntry is the per-criterion criteria-report payload.
	CriteriaReportEntry = cyclesstore.CriteriaReportEntry
	// VerifyReportEntry is the verify-report counterpart of CriteriaReportEntry.
	VerifyReportEntry = cyclesstore.VerifyReportEntry
	// CommandRunEntry is one verify-phase shell command execution row.
	CommandRunEntry = cyclesstore.CommandRunEntry
)

// UpsertCriteriaReports persists one batch of per-criterion criteria-report rows.
func (s *Store) UpsertCriteriaReports(ctx context.Context, cycleID string, attemptSeq int64, entries []CriteriaReportEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertCriteriaReports")
	return s.cycles.UpsertCriteriaReports(ctx, cycleID, attemptSeq, entries)
}

// UpsertVerifyReports is the verify-report counterpart of UpsertCriteriaReports.
func (s *Store) UpsertVerifyReports(ctx context.Context, cycleID string, attemptSeq int64, entries []VerifyReportEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertVerifyReports")
	return s.cycles.UpsertVerifyReports(ctx, cycleID, attemptSeq, entries)
}

// ListCriteriaReportsForCycle returns every persisted criteria-report row for cycleID.
func (s *Store) ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCriteriaReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCriteriaReportsForCycle")
	return s.cycles.ListCriteriaReportsForCycle(ctx, cycleID)
}

// ListVerifyReportsForCycle is the verify counterpart of ListCriteriaReportsForCycle.
func (s *Store) ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleVerifyReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListVerifyReportsForCycle")
	return s.cycles.ListVerifyReportsForCycle(ctx, cycleID)
}

// GetCriteriaReport returns the criteria-report row for (cycleID, attemptSeq, criterionID).
func (s *Store) GetCriteriaReport(ctx context.Context, cycleID string, attemptSeq int64, criterionID string) (*domain.TaskCycleCriteriaReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetCriteriaReport")
	return s.cycles.GetCriteriaReport(ctx, cycleID, attemptSeq, criterionID)
}

// UpsertCommandRuns persists command run metadata for one verify attempt.
func (s *Store) UpsertCommandRuns(ctx context.Context, cycleID string, attemptSeq int64, entries []CommandRunEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertCommandRuns")
	return s.cycles.UpsertCommandRuns(ctx, cycleID, attemptSeq, entries)
}

// ListCommandRunsForCycle returns command run rows for cycleID.
func (s *Store) ListCommandRunsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommandRun, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCommandRunsForCycle")
	return s.cycles.ListCommandRunsForCycle(ctx, cycleID)
}
