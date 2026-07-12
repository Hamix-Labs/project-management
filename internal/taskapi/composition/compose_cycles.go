package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

// StartCycle creates a new TaskCycle row with status=running for the given task.
func (a *API) StartCycle(ctx context.Context, in cyclescontract.StartCycleInput) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.StartCycle")
	return a.cycles.StartCycle(ctx, in)
}

// TerminateCycle moves a running cycle into a terminal state.
func (a *API) TerminateCycle(ctx context.Context, cycleID string, status cyclesdomain.CycleStatus, reason string, by taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.TerminateCycle")
	return a.cycles.TerminateCycle(ctx, cycleID, status, reason, by)
}

// GetCycle returns one cycle by id; ErrNotFound when missing.
func (a *API) GetCycle(ctx context.Context, cycleID string) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetCycle")
	return a.cycles.GetCycle(ctx, cycleID)
}

// ListCyclesForTask returns cycles for a task ordered by attempt_seq DESC (newest first).
func (a *API) ListCyclesForTask(ctx context.Context, taskID string, limit int) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCyclesForTask")
	return a.cycles.ListCyclesForTask(ctx, taskID, limit)
}

// ListCyclesForTaskBefore is the keyset-paginated form of ListCyclesForTask.
func (a *API) ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCyclesForTaskBefore")
	return a.cycles.ListCyclesForTaskBefore(ctx, taskID, beforeAttemptSeq, limit)
}

// StartPhase appends a new phase row to a running cycle.
func (a *API) StartPhase(ctx context.Context, cycleID string, phase cyclesdomain.Phase, by taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.StartPhase")
	return a.cycles.StartPhase(ctx, cycleID, phase, by)
}

// CompletePhase moves a running phase to a terminal status.
func (a *API) CompletePhase(ctx context.Context, in cyclescontract.CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CompletePhase")
	return a.cycles.CompletePhase(ctx, in)
}

// ListPhasesForCycle returns phases for cycleID in execution order (phase_seq ASC).
func (a *API) ListPhasesForCycle(ctx context.Context, cycleID string) ([]cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListPhasesForCycle")
	return a.cycles.ListPhasesForCycle(ctx, cycleID)
}

// LastSessionID returns the Cursor session_id from the latest completed phase row.
func (a *API) LastSessionID(ctx context.Context, cycleID string, phase cyclesdomain.Phase) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.LastSessionID",
		"cycle_id", cycleID, "phase", string(phase))
	return a.cycles.LastSessionID(ctx, cycleID, phase)
}

// AppendCycleStreamEvent persists one normalized runner progress event for a cycle.
func (a *API) AppendCycleStreamEvent(ctx context.Context, in cyclesstore.AppendCycleStreamEventInput) (*cyclesdomain.TaskCycleStreamEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AppendCycleStreamEvent")
	return a.cycles.AppendCycleStreamEvent(ctx, in)
}

// ListCycleStreamEvents returns persisted stream events for cycleID ordered by stream_seq ASC.
func (a *API) ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]cyclesdomain.TaskCycleStreamEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCycleStreamEvents")
	return a.cycles.ListCycleStreamEvents(ctx, cycleID, afterSeq, limit)
}

// ListRunningCycles returns every cycle currently in CycleStatusRunning across all tasks.
func (a *API) ListRunningCycles(ctx context.Context) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListRunningCycles")
	return a.cycles.ListRunningCycles(ctx)
}

// ListRunningCyclePhases returns every phase row currently in PhaseStatusRunning.
func (a *API) ListRunningCyclePhases(ctx context.Context) ([]cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListRunningCyclePhases")
	return a.cycles.ListRunningCyclePhases(ctx)
}
