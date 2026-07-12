package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"
	"time"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/scheduling"
)

//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcore/store; operation trace is emitted by the BC chokepoint."
func ShouldNotifyReadyNow(pickupNotBefore *time.Time, now time.Time) bool {
	return taskcorestore.ShouldNotifyReadyNow(pickupNotBefore, now)
}

func (a *API) Get(ctx context.Context, id string) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Get")
	return a.taskcore.Get(ctx, id)
}

func (a *API) AgentPickup(ctx context.Context, taskID string, by taskcoredomain.Actor) (*taskcorecontract.AgentPickupResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AgentPickup", "task_id", taskID)
	return a.taskcore.AgentPickup(ctx, taskID, by)
}

func (a *API) RequestTaskRetry(ctx context.Context, in taskcorestore.RequestRetryInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RequestTaskRetry", "task_id", in.TaskID)
	updated, prev, err := a.taskcore.RequestTaskRetry(ctx, in, by)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	now := time.Now().UTC()
	a.applyNotifyDecision(ctx, *updated, scheduling.DecideNotifyAfterReadyTransition(prev, updated, false, now))
	return updated, nil
}

func (a *API) Create(ctx context.Context, in taskcorestore.CreateTaskInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Create")
	t, err := a.taskcore.Create(ctx, in, by)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	a.applyNotifyDecision(ctx, *t, scheduling.DecideNotifyAfterReadyTransition("", t, false, now))
	return t, nil
}

func (a *API) Update(ctx context.Context, id string, in taskcorestore.UpdateTaskInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Update")
	updated, prev, err := a.taskcore.Update(ctx, id, in, by)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	if updated.Status == taskcoredomain.StatusDone && prev != taskcoredomain.StatusDone {
		a.notifyUnblockedDependents(ctx, updated.ID)
	}
	now := time.Now().UTC()
	pickupTouched := in.PickupNotBefore != nil
	a.applyNotifyDecision(ctx, *updated, scheduling.DecideNotifyAfterReadyTransition(prev, updated, pickupTouched, now))
	return updated, nil
}

func (a *API) notifyUnblockedDependents(ctx context.Context, predecessorID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.notifyUnblockedDependents", "predecessor_id", predecessorID)
	dependents, err := a.taskcore.ListDependents(ctx, predecessorID)
	if err != nil {
		slog.Warn("list dependents after predecessor unblock", "task_id", predecessorID, "err", err)
		return
	}
	now := time.Now().UTC()
	for _, id := range dependents {
		t, err := a.taskcore.Get(ctx, id)
		if err != nil {
			continue
		}
		ok, _, err := a.taskcore.ReadyForAgentPickup(ctx, t, now)
		if err != nil || !ok {
			continue
		}
		a.notifyReadyTask(ctx, *t)
	}
}

func (a *API) NotifyUnblockedDependents(ctx context.Context, predecessorID string) {
	a.notifyUnblockedDependents(ctx, predecessorID)
}

func (a *API) Delete(ctx context.Context, id string, by taskcoredomain.Actor) ([]string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Delete")
	deletedIDs, err := a.taskcore.Delete(ctx, id, by)
	if err != nil {
		return nil, err
	}
	for _, tid := range deletedIDs {
		a.cancelPickupWake(tid)
	}
	return deletedIDs, nil
}

func (a *API) ListFlat(ctx context.Context, limit, offset int, filter *taskcorestore.ListFilter) ([]taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListFlat")
	return a.taskcore.ListFlat(ctx, limit, offset, filter)
}

func (a *API) ListFlatPage(ctx context.Context, limit, offset int, filter *taskcorestore.ListFilter) ([]taskcoredomain.Task, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListFlatPage")
	return a.taskcore.ListFlatPage(ctx, limit, offset, filter)
}

func (a *API) ListFlatAfter(ctx context.Context, limit int, afterID string) ([]taskcoredomain.Task, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListFlatAfter")
	return a.taskcore.ListFlatAfter(ctx, limit, afterID)
}

func (a *API) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies taskcoredomain.DependencySatisfies) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AddTaskDependency")
	return a.taskcore.AddTaskDependency(ctx, taskID, dependsOnTaskID, satisfies)
}

func (a *API) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveTaskDependency")
	return a.taskcore.RemoveTaskDependency(ctx, taskID, dependsOnTaskID)
}

func (a *API) ListTaskDependencies(ctx context.Context, taskID string) ([]taskcoredomain.DependencyEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTaskDependencies")
	return a.taskcore.ListTaskDependencies(ctx, taskID)
}

func (a *API) SetTaskDependencies(ctx context.Context, taskID string, dependsOn []taskcoredomain.DependencyEdge) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetTaskDependencies")
	return a.taskcore.SetTaskDependencies(ctx, taskID, dependsOn)
}

func (a *API) ReadyForAgentPickup(ctx context.Context, t *taskcoredomain.Task, now time.Time) (bool, taskcorecontract.FailedPredicate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReadyForAgentPickup")
	return a.taskcore.ReadyForAgentPickup(ctx, t, now)
}

func (a *API) ApplyTaskGateAction(ctx context.Context, taskID, action string, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ApplyTaskGateAction")
	return a.taskcore.ApplyTaskGateAction(ctx, taskID, action, by)
}
