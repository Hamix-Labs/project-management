package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

type (
	// StartCycleInput is the public re-export of the cycles subpackage input struct.
	StartCycleInput = cyclescontract.StartCycleInput
	// CompletePhaseInput is the public re-export of the phase completion input struct.
	CompletePhaseInput = cyclescontract.CompletePhaseInput
	// AppendCycleStreamEventInput is the durable per-attempt stream event input.
	AppendCycleStreamEventInput = cyclesstore.AppendCycleStreamEventInput
)

// StartCycle creates a new TaskCycle row with status=running for the given task.
func (s *Store) StartCycle(ctx context.Context, in StartCycleInput) (*domain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.StartCycle")
	return s.cycles.StartCycle(ctx, in)
}

// TerminateCycle moves a running cycle into a terminal state.
func (s *Store) TerminateCycle(ctx context.Context, cycleID string, status domain.CycleStatus, reason string, by domain.Actor) (*domain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.TerminateCycle")
	return s.cycles.TerminateCycle(ctx, cycleID, status, reason, by)
}

// GetCycle returns one cycle by id; ErrNotFound when missing.
func (s *Store) GetCycle(ctx context.Context, cycleID string) (*domain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetCycle")
	return s.cycles.GetCycle(ctx, cycleID)
}

// ListCyclesForTask returns cycles for a task ordered by attempt_seq DESC (newest first).
func (s *Store) ListCyclesForTask(ctx context.Context, taskID string, limit int) ([]domain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCyclesForTask")
	return s.cycles.ListCyclesForTask(ctx, taskID, limit)
}

// ListCyclesForTaskBefore is the keyset-paginated form of ListCyclesForTask.
func (s *Store) ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]domain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCyclesForTaskBefore")
	return s.cycles.ListCyclesForTaskBefore(ctx, taskID, beforeAttemptSeq, limit)
}

// StartPhase appends a new phase row to a running cycle.
func (s *Store) StartPhase(ctx context.Context, cycleID string, phase domain.Phase, by domain.Actor) (*domain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.StartPhase")
	return s.cycles.StartPhase(ctx, cycleID, phase, by)
}

// CompletePhase moves a running phase to a terminal status.
func (s *Store) CompletePhase(ctx context.Context, in CompletePhaseInput) (*domain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CompletePhase")
	return s.cycles.CompletePhase(ctx, in)
}

// ListPhasesForCycle returns phases for cycleID in execution order (phase_seq ASC).
func (s *Store) ListPhasesForCycle(ctx context.Context, cycleID string) ([]domain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListPhasesForCycle")
	return s.cycles.ListPhasesForCycle(ctx, cycleID)
}

// LastSessionID returns the Cursor session_id from the latest completed phase row.
func (s *Store) LastSessionID(ctx context.Context, cycleID string, phase domain.Phase) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.LastSessionID",
		"cycle_id", cycleID, "phase", string(phase))
	return s.cycles.LastSessionID(ctx, cycleID, phase)
}

// AppendCycleStreamEvent persists one normalized runner progress event for a cycle.
func (s *Store) AppendCycleStreamEvent(ctx context.Context, in AppendCycleStreamEventInput) (*domain.TaskCycleStreamEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AppendCycleStreamEvent")
	return s.cycles.AppendCycleStreamEvent(ctx, in)
}

// ListCycleStreamEvents returns persisted stream events for cycleID ordered by stream_seq ASC.
func (s *Store) ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]domain.TaskCycleStreamEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCycleStreamEvents")
	return s.cycles.ListCycleStreamEvents(ctx, cycleID, afterSeq, limit)
}

// ListRunningCycles returns every cycle currently in CycleStatusRunning across all tasks.
func (s *Store) ListRunningCycles(ctx context.Context) ([]domain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListRunningCycles")
	return s.cycles.ListRunningCycles(ctx)
}

// ListRunningCyclePhases returns every phase row currently in PhaseStatusRunning.
func (s *Store) ListRunningCyclePhases(ctx context.Context) ([]domain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListRunningCyclePhases")
	return s.cycles.ListRunningCyclePhases(ctx)
}
