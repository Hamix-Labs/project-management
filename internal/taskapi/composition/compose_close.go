package composition

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// CancelRunForTaskFunc cancels an in-flight agent run for taskID when present.
type CancelRunForTaskFunc func(taskID string) bool

// QueueDropFunc removes a task id from the in-memory agent queue pending set.
type QueueDropFunc func(taskID string)

// SetCancelRunForTask registers the harness cancel hook used on close.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook."
func (a *API) SetCancelRunForTask(fn CancelRunForTaskFunc) {
	if a == nil {
		return
	}
	a.cancelRunForTask = fn
}

// SetQueueDrop registers the memory-queue drop hook used on close.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook."
func (a *API) SetQueueDrop(fn QueueDropFunc) {
	if a == nil {
		return
	}
	a.queueDrop = fn
}

// Close is the sole entry point for terminating a task's lifecycle.
//
// Composition order (matches Plan 2 in HARNESS_IMPROVEMENTS.md):
//  1. Load task (404 if missing).
//  2. If already closed, return the current task unchanged (idempotent 200,
//     no second audit and no cancel/drop side effects).
//  3. Cancel any in-flight agent run for this task (CancelRunForTask —
//     never CancelCurrentRun; see docs/domain/agent-queue.md).
//  4. Drop the task from the in-memory agent queue pending set.
//  5. Cancel any scheduled pickup wake.
//  6. Persist the terminal status transition (clears pending_retry,
//     appends a status_changed event).
//
// Callers (handler) publish SSE task_updated on success; nothing is
// published on the idempotent path — the caller already knows the task
// is closed.
func (a *API) Close(ctx context.Context, id string, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Close", "task_id", id)
	cur, err := a.taskcore.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur.Status == taskcoredomain.StatusClosed {
		return cur, nil
	}
	if a.cancelRunForTask != nil {
		a.cancelRunForTask(id)
	}
	if a.queueDrop != nil {
		a.queueDrop(id)
	}
	a.cancelPickupWake(id)
	return a.taskcore.Close(ctx, id, by)
}

// Reopen transitions a closed task back to ready (409 if not currently closed).
func (a *API) Reopen(ctx context.Context, id string, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Reopen", "task_id", id)
	return a.taskcore.Reopen(ctx, id, by)
}
