// Package store implements GORM persistence for execution cycles, phases,
// stream events, verdict reports, and indexed commits.
package store

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/commits"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/cycles"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/reports"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for task execution cycles.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.NewStore")
	return &Store{db: db}
}

type (
	// StartCycleInput is the public re-export of the cycles subpackage input struct.
	StartCycleInput = contract.StartCycleInput
	// CompletePhaseInput is the public re-export of the phase completion input struct.
	CompletePhaseInput = contract.CompletePhaseInput
	// AppendCycleStreamEventInput is the durable per-attempt stream event input.
	AppendCycleStreamEventInput = contract.AppendCycleStreamEventInput
	// CycleCommitEntry is a commit upsert payload.
	CycleCommitEntry = contract.CycleCommitEntry
	// CriteriaReportEntry is the per-criterion criteria-report payload.
	CriteriaReportEntry = contract.CriteriaReportEntry
	// VerifyReportEntry is the verify-report counterpart of CriteriaReportEntry.
	VerifyReportEntry = contract.VerifyReportEntry
	// CommandRunEntry is one verify-phase shell command execution row.
	CommandRunEntry = contract.CommandRunEntry
)

// FailureSurfaceMessage returns operator-facing failure text for cycle_failed mirrors.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; cycles.FailureSurfaceMessage is a pure helper."
func FailureSurfaceMessage(hasPhase bool, cycleReason, phaseSummary string, phaseDetails map[string]any) string {
	return cycles.FailureSurfaceMessage(hasPhase, cycleReason, phaseSummary, phaseDetails)
}

func (s *Store) StartCycle(ctx context.Context, in StartCycleInput) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.StartCycle")
	return cycles.Start(ctx, s.db, in)
}

func (s *Store) TerminateCycle(ctx context.Context, cycleID string, status cyclesdomain.CycleStatus, reason string, by taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.TerminateCycle")
	return cycles.Terminate(ctx, s.db, cycleID, status, reason, by)
}

func (s *Store) GetCycle(ctx context.Context, cycleID string) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.GetCycle")
	return cycles.Get(ctx, s.db, cycleID)
}

func (s *Store) ListCyclesForTask(ctx context.Context, taskID string, limit int) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCyclesForTask")
	return cycles.ListForTask(ctx, s.db, taskID, limit)
}

func (s *Store) ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCyclesForTaskBefore")
	return cycles.ListForTaskBefore(ctx, s.db, taskID, beforeAttemptSeq, limit)
}

func (s *Store) StartPhase(ctx context.Context, cycleID string, phase cyclesdomain.Phase, by taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.StartPhase")
	return cycles.StartPhase(ctx, s.db, cycleID, phase, by)
}

func (s *Store) CompletePhase(ctx context.Context, in CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.CompletePhase")
	return cycles.CompletePhase(ctx, s.db, in)
}

func (s *Store) ListPhasesForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListPhasesForCycle")
	return cycles.ListPhasesForCycle(ctx, s.db, cycleID)
}

func (s *Store) LastSessionID(ctx context.Context, cycleID string, phase cyclesdomain.Phase) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.LastSessionID",
		"cycle_id", cycleID, "phase", string(phase))
	return cycles.LastSessionID(ctx, s.db, cycleID, phase)
}

func (s *Store) AppendCycleStreamEvent(ctx context.Context, in AppendCycleStreamEventInput) (*cyclesdomain.TaskCycleStreamEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.AppendCycleStreamEvent")
	return cycles.AppendStreamEvent(ctx, s.db, appendStreamIn(in))
}

func (s *Store) ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]cyclesdomain.TaskCycleStreamEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCycleStreamEvents")
	return cycles.ListStreamEvents(ctx, s.db, cycleID, afterSeq, limit)
}

func (s *Store) ListRunningCycles(ctx context.Context) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListRunningCycles")
	return cycles.ListRunning(ctx, s.db)
}

func (s *Store) ListRunningCyclePhases(ctx context.Context) ([]cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListRunningCyclePhases")
	return cycles.ListRunningPhases(ctx, s.db)
}

func (s *Store) UpsertCycleCommits(ctx context.Context, taskID, cycleID string, entries []CycleCommitEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.UpsertCycleCommits")
	return commits.UpsertCycleCommits(ctx, s.db, taskID, cycleID, cycleCommitEntries(entries))
}

func (s *Store) ListCommitsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCommitsForCycle")
	return commits.ListCommitsForCycle(ctx, s.db, cycleID)
}

func (s *Store) ListCommitsForTask(ctx context.Context, taskID string) ([]cyclesdomain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCommitsForTask")
	return commits.ListCommitsForTask(ctx, s.db, taskID)
}

func (s *Store) UpsertCriteriaReports(ctx context.Context, cycleID string, attemptSeq int64, entries []CriteriaReportEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.UpsertCriteriaReports")
	return reports.UpsertCriteriaReports(ctx, s.db, cycleID, attemptSeq, criteriaReportEntries(entries))
}

func (s *Store) UpsertVerifyReports(ctx context.Context, cycleID string, attemptSeq int64, entries []VerifyReportEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.UpsertVerifyReports")
	return reports.UpsertVerifyReports(ctx, s.db, cycleID, attemptSeq, verifyReportEntries(entries))
}

func (s *Store) ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCriteriaReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCriteriaReportsForCycle")
	return reports.ListCriteriaReportsForCycle(ctx, s.db, cycleID)
}

func (s *Store) ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleVerifyReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListVerifyReportsForCycle")
	return reports.ListVerifyReportsForCycle(ctx, s.db, cycleID)
}

func (s *Store) GetCriteriaReport(ctx context.Context, cycleID string, attemptSeq int64, criterionID string) (*cyclesdomain.TaskCycleCriteriaReport, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.GetCriteriaReport")
	return reports.GetCriteriaReport(ctx, s.db, cycleID, attemptSeq, criterionID)
}

func (s *Store) UpsertCommandRuns(ctx context.Context, cycleID string, attemptSeq int64, entries []CommandRunEntry) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.UpsertCommandRuns")
	return reports.UpsertCommandRuns(ctx, s.db, cycleID, attemptSeq, commandRunEntries(entries))
}

func (s *Store) ListCommandRunsForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCycleCommandRun, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.ListCommandRunsForCycle")
	return reports.ListCommandRunsForCycle(ctx, s.db, cycleID)
}
