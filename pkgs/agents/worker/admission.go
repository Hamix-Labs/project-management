package worker

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"log/slog"
	"strings"
	"time"
)

const pickupPersistenceDefer = 2 * time.Minute

// reloadTask fetches the freshest task row from the store. ok==false
// means the caller should bail (and AckAfterRecv via the deferred path).
func (w *Worker) reloadTask(ctx context.Context, taskID string) (*taskcoredomain.Task, bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.worker.Worker.reloadTask",
		"task_id", taskID)
	fresh, err := w.store.Get(ctx, taskID)
	if err == nil {
		return fresh, true
	}
	if errors.Is(err, taskcoredomain.ErrNotFound) {
		slog.Info("task vanished before dequeue processing", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.reloadTask.not_found", "task_id", taskID)
		return nil, false
	}
	slog.Warn("agent worker reload failed", "cmd", calltrace.LogCmd,
		"operation", "agent.worker.Worker.reloadTask.err", "task_id", taskID, "err", err)
	return nil, false
}

func (w *Worker) deferTaskPickup(ctx context.Context, taskID string, delay time.Duration) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.worker.Worker.deferTaskPickup",
		"task_id", taskID, "delay", delay.String())
	at := w.clock().Add(delay).UTC()
	patch := taskcorestore.PickupNotBeforePatch{At: at}
	if _, err := w.store.Update(ctx, taskID, taskcorestore.UpdateTaskInput{PickupNotBefore: &patch}, taskcoredomain.ActorAgent); err != nil {
		slog.Warn("agent worker defer pickup failed", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.deferTaskPickup.err", "task_id", taskID, "err", err)
	}
}

// transitionTaskToRunning flips the task to running before the harness runs.
// Returns the post-pickup task row and any consumed retry intent.
func (w *Worker) transitionTaskToRunning(ctx context.Context, taskID string) (*taskcoredomain.Task, *taskcoredomain.PendingRetry, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.worker.Worker.transitionTaskToRunning",
		"task_id", taskID)
	res, err := w.store.AgentPickup(ctx, taskID, taskcoredomain.ActorAgent)
	if err != nil {
		level := slog.LevelWarn
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			level = slog.LevelInfo
		}
		slog.Log(ctx, level, "agent worker task pickup failed",
			"cmd", calltrace.LogCmd, "operation", "agent.worker.Worker.transitionTaskToRunning.err",
			"task_id", taskID, "err", err)
		return nil, nil, err
	}
	return res.Task, res.ConsumedRetry, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func pickupPersistenceFailure(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, taskcoredomain.ErrNotFound) || errors.Is(err, taskcoredomain.ErrInvalidInput) {
		return false
	}
	return true
}

func (w *Worker) recordPickupPersistenceFailure(ctx context.Context, taskID string, pickupErr error) {
	slog.Warn("agent worker pickup persistence failure; deferring retry",
		"cmd", calltrace.LogCmd, "operation", "agent.worker.Worker.recordPickupPersistenceFailure",
		"task_id", taskID, "err", pickupErr)
	w.deferTaskPickup(ctx, taskID, pickupPersistenceDefer)
	payload, err := json.Marshal(map[string]string{"reason": "persistence"})
	if err != nil {
		slog.Warn("agent worker pickup failure event marshal failed", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.recordPickupPersistenceFailure.marshal", "task_id", taskID, "err", err)
		return
	}
	if err := w.store.AppendTaskEvent(ctx, taskID, taskeventsdomain.EventTaskPickupFailed, taskcoredomain.ActorAgent, payload); err != nil {
		slog.Warn("agent worker pickup failure event append failed", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.recordPickupPersistenceFailure.append", "task_id", taskID, "err", err)
	}
}

func (w *Worker) openRunningCycle(ctx context.Context, taskID string) (*cyclesdomain.TaskCycle, bool) {
	cycles, err := w.store.ListCyclesForTask(ctx, taskID, 0)
	if err != nil {
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			return nil, false
		}
		slog.Warn("agent worker list cycles failed", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.openRunningCycle.err", "task_id", taskID, "err", err)
		return nil, false
	}
	for i := len(cycles) - 1; i >= 0; i-- {
		if cycles[i].Status == cyclesdomain.CycleStatusRunning {
			cycle := cycles[i]
			return &cycle, true
		}
	}
	return nil, false
}

// processOne runs queue admission then delegates the cycle body to the harness.
func (w *Worker) processOne(parentCtx context.Context, task taskcoredomain.Task) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.worker.Worker.processOne",
		"task_id", task.ID)
	defer w.queue.AckAfterRecv(task.ID)
	defer w.recoverAdmissionPanic(task.ID)

	fresh, ok := w.reloadTask(parentCtx, task.ID)
	if !ok {
		return
	}

	switch fresh.Status {
	case taskcoredomain.StatusRunning:
		cycle, ok := w.openRunningCycle(parentCtx, fresh.ID)
		if !ok {
			slog.Warn("running task without open cycle at dequeue", "cmd", calltrace.LogCmd,
				"operation", "agent.worker.Worker.processOne.no_open_cycle", "task_id", task.ID)
			return
		}
		if !taskHasBinding(fresh) {
			slog.Warn("running task missing git binding", "cmd", calltrace.LogCmd,
				"operation", "agent.worker.Worker.processOne.missing_binding", "task_id", task.ID)
			return
		}
		unlock := w.gate.Lock(strings.TrimSpace(*fresh.WorktreeID))
		defer unlock()
		w.runWithGitPrep(parentCtx, fresh, func() {
			w.harness.Resume(parentCtx, fresh, cycle)
		})
		return
	case taskcoredomain.StatusReady:
		// continue below
	default:
		slog.Warn("stale task at dequeue", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.processOne.stale", "task_id", task.ID,
			"status", string(fresh.Status))
		return
	}

	now := w.clock()
	ready, failedPredicate, err := w.store.ReadyForAgentPickup(parentCtx, fresh, now)
	if err != nil {
		slog.Warn("agent worker readiness check failed", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.processOne.readiness", "task_id", task.ID, "err", err)
		return
	}
	if !ready {
		slog.Debug("agent worker admission deferred", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.processOne.defer",
			"task_id", task.ID, "failed_predicate", string(failedPredicate))
		w.deferTaskPickup(parentCtx, task.ID, 60*time.Second)
		return
	}
	if !taskHasBinding(fresh) {
		slog.Warn("agent worker task missing git binding; deferring pickup", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.processOne.missing_binding", "task_id", task.ID)
		w.deferTaskPickup(parentCtx, task.ID, 60*time.Second)
		return
	}
	wtID := strings.TrimSpace(*fresh.WorktreeID)
	unlock, acquired := w.gate.TryLock(wtID)
	if !acquired {
		slog.Debug("agent worker worktree busy; deferring pickup", "cmd", calltrace.LogCmd,
			"operation", "agent.worker.Worker.processOne.worktree_busy",
			"task_id", task.ID, "worktree_id", wtID)
		w.deferTaskPickup(parentCtx, task.ID, 5*time.Second)
		return
	}
	defer unlock()
	picked, consumedRetry, err := w.transitionTaskToRunning(parentCtx, task.ID)
	if err != nil {
		if pickupPersistenceFailure(err) {
			w.recordPickupPersistenceFailure(parentCtx, task.ID, err)
		}
		return
	}
	w.runWithGitPrep(parentCtx, picked, func() {
		w.harness.RunWithRetry(parentCtx, picked, consumedRetry)
	})
}

func (w *Worker) recoverAdmissionPanic(taskID string) {
	if recover() == nil {
		return
	}
	slog.Error("agent worker admission panic", "cmd", calltrace.LogCmd,
		"operation", "agent.worker.Worker.recoverAdmissionPanic", "task_id", taskID)
	bg, cancel := context.WithTimeout(context.Background(), DefaultShutdownAbortTimeout)
	defer cancel()
	failed := taskcoredomain.StatusFailed
	if _, err := w.store.Update(bg, taskID, taskcorestore.UpdateTaskInput{Status: &failed}, taskcoredomain.ActorAgent); err != nil {
		if !errors.Is(err, taskcoredomain.ErrNotFound) {
			slog.Warn("agent worker admission panic task transition failed", "cmd", calltrace.LogCmd,
				"operation", "agent.worker.Worker.recoverAdmissionPanic.err",
				"task_id", taskID, "err", err)
		}
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *Worker) clock() time.Time {
	if w.opts.Clock != nil {
		return w.opts.Clock()
	}
	return time.Now().UTC()
}
