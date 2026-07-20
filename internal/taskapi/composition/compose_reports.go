package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

// UpsertCriteriaReports persists one batch of per-criterion criteria-report rows.
func (a *API) UpsertCriteriaReports(ctx context.Context, cycleID string, attemptSeq int64, entries []cyclesstore.CriteriaReportEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertCriteriaReports")
	return a.cycles.UpsertCriteriaReports(ctx, cycleID, attemptSeq, entries)
}

// UpsertVerifyReports is the verify-report counterpart of UpsertCriteriaReports.
func (a *API) UpsertVerifyReports(ctx context.Context, cycleID string, attemptSeq int64, entries []cyclesstore.VerifyReportEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertVerifyReports")
	return a.cycles.UpsertVerifyReports(ctx, cycleID, attemptSeq, entries)
}

// ListCriteriaReportsForCycle returns every persisted criteria-report row for cycleID.
func (a *API) ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCriteriaReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCriteriaReportsForCycle")
	return a.cycles.ListCriteriaReportsForCycle(ctx, cycleID)
}

// ListVerifyReportsForCycle is the verify counterpart of ListCriteriaReportsForCycle.
func (a *API) ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleVerifyReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListVerifyReportsForCycle")
	return a.cycles.ListVerifyReportsForCycle(ctx, cycleID)
}

// GetCriteriaReport returns the criteria-report row for (cycleID, attemptSeq, criterionID).
func (a *API) GetCriteriaReport(ctx context.Context, cycleID string, attemptSeq int64, criterionID string) (*cyclesdomain.TaskCycleCriteriaReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetCriteriaReport")
	return a.cycles.GetCriteriaReport(ctx, cycleID, attemptSeq, criterionID)
}

// UpsertCommandRuns persists command run metadata for one verify attempt.
func (a *API) UpsertCommandRuns(ctx context.Context, cycleID string, attemptSeq int64, entries []cyclesstore.CommandRunEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpsertCommandRuns")
	return a.cycles.UpsertCommandRuns(ctx, cycleID, attemptSeq, entries)
}

// ListCommandRunsForCycle returns command run rows for cycleID.
func (a *API) ListCommandRunsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommandRun, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCommandRunsForCycle")
	return a.cycles.ListCommandRunsForCycle(ctx, cycleID)
}
