package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"
	"time"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/scheduling"
)

type (
	FailedPredicate         = taskcorestore.FailedPredicate
	CreateTaskInput         = taskcorestore.CreateTaskInput
	UpdateTaskInput         = taskcorestore.UpdateTaskInput
	ProjectFieldPatch       = taskcorestore.ProjectFieldPatch
	PickupNotBeforePatch    = taskcorestore.PickupNotBeforePatch
	RequestRetryInput       = taskcorestore.RequestRetryInput
	AgentPickupResult       = taskcorestore.AgentPickupResult
	ListFilter              = taskcorestore.ListFilter
	ReadyTaskQueueCursor    = taskcorestore.ReadyTaskQueueCursor
	ReadyTaskQueueCandidate = taskcorestore.ReadyTaskQueueCandidate
	DeferredPickup          = taskcorestore.DeferredPickup
)

//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcore/store; operation trace is emitted by the BC chokepoint."
func ShouldNotifyReadyNow(pickupNotBefore *time.Time, now time.Time) bool {
	return taskcorestore.ShouldNotifyReadyNow(pickupNotBefore, now)
}

func (s *Store) Get(ctx context.Context, id string) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Get")
	return s.taskcore.Get(ctx, id)
}

func (s *Store) AgentPickup(ctx context.Context, taskID string, by domain.Actor) (*AgentPickupResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AgentPickup", "task_id", taskID)
	return s.taskcore.AgentPickup(ctx, taskID, by)
}

func (s *Store) RequestTaskRetry(ctx context.Context, in RequestRetryInput, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RequestTaskRetry", "task_id", in.TaskID)
	updated, prev, err := s.taskcore.RequestTaskRetry(ctx, in, by)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	now := time.Now().UTC()
	s.applyNotifyDecision(ctx, *updated, scheduling.DecideNotifyAfterReadyTransition(prev, updated, false, now))
	return updated, nil
}

func (s *Store) Create(ctx context.Context, in CreateTaskInput, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Create")
	t, err := s.taskcore.Create(ctx, in, by)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	s.applyNotifyDecision(ctx, *t, scheduling.DecideNotifyAfterReadyTransition("", t, false, now))
	return t, nil
}

func (s *Store) Update(ctx context.Context, id string, in UpdateTaskInput, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Update")
	updated, prev, err := s.taskcore.Update(ctx, id, in, by)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	if updated.Status == domain.StatusDone && prev != domain.StatusDone {
		s.notifyUnblockedDependents(ctx, updated.ID)
	}
	now := time.Now().UTC()
	pickupTouched := in.PickupNotBefore != nil
	s.applyNotifyDecision(ctx, *updated, scheduling.DecideNotifyAfterReadyTransition(prev, updated, pickupTouched, now))
	return updated, nil
}

func (s *Store) notifyUnblockedDependents(ctx context.Context, predecessorID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.notifyUnblockedDependents", "predecessor_id", predecessorID)
	dependents, err := s.taskcore.ListDependents(ctx, predecessorID)
	if err != nil {
		slog.Warn("list dependents after predecessor unblock", "task_id", predecessorID, "err", err)
		return
	}
	now := time.Now().UTC()
	for _, id := range dependents {
		t, err := s.taskcore.Get(ctx, id)
		if err != nil {
			continue
		}
		ok, _, err := s.taskcore.ReadyForAgentPickup(ctx, t, now)
		if err != nil || !ok {
			continue
		}
		s.notifyReadyTask(ctx, *t)
	}
}

func (s *Store) NotifyUnblockedDependents(ctx context.Context, predecessorID string) {
	s.notifyUnblockedDependents(ctx, predecessorID)
}

func (s *Store) Delete(ctx context.Context, id string, by domain.Actor) ([]string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Delete")
	deletedIDs, err := s.taskcore.Delete(ctx, id, by)
	if err != nil {
		return nil, err
	}
	for _, tid := range deletedIDs {
		s.cancelPickupWake(tid)
	}
	return deletedIDs, nil
}

func (s *Store) ListFlat(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListFlat")
	return s.taskcore.ListFlat(ctx, limit, offset, filter)
}

func (s *Store) ListFlatPage(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListFlatPage")
	return s.taskcore.ListFlatPage(ctx, limit, offset, filter)
}

func (s *Store) ListFlatAfter(ctx context.Context, limit int, afterID string) ([]domain.Task, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListFlatAfter")
	return s.taskcore.ListFlatAfter(ctx, limit, afterID)
}

func (s *Store) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies domain.DependencySatisfies) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AddTaskDependency")
	return s.taskcore.AddTaskDependency(ctx, taskID, dependsOnTaskID, satisfies)
}

func (s *Store) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveTaskDependency")
	return s.taskcore.RemoveTaskDependency(ctx, taskID, dependsOnTaskID)
}

func (s *Store) ListTaskDependencies(ctx context.Context, taskID string) ([]domain.DependencyEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTaskDependencies")
	return s.taskcore.ListTaskDependencies(ctx, taskID)
}

func (s *Store) SetTaskDependencies(ctx context.Context, taskID string, dependsOn []domain.DependencyEdge) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetTaskDependencies")
	return s.taskcore.SetTaskDependencies(ctx, taskID, dependsOn)
}

func (s *Store) ReadyForAgentPickup(ctx context.Context, t *domain.Task, now time.Time) (bool, FailedPredicate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReadyForAgentPickup")
	return s.taskcore.ReadyForAgentPickup(ctx, t, now)
}

//funclogmeasure:skip category=delegate-already-logs reason="Notify side-effect helper; parent store method emits trace at the chokepoint."
func (s *Store) applyNotifyDecision(ctx context.Context, task domain.Task, d scheduling.NotifyDecision) {
	if d.ScheduleWake != nil {
		s.schedulePickupWake(ctx, task.ID, *d.ScheduleWake)
		return
	}
	if d.CancelWake {
		s.cancelPickupWake(task.ID)
	}
	if d.NotifyQueue {
		s.notifyReadyTask(ctx, task)
	}
}

func (s *Store) ApplyTaskGateAction(ctx context.Context, taskID, action string, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ApplyTaskGateAction")
	return s.taskcore.ApplyTaskGateAction(ctx, taskID, action, by)
}
