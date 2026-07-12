package storehooks

import (
	"context"
	"time"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/scheduling"
)

// ApplyNotifyDecision runs pickup-wake scheduling/cancel and ready-queue notify
// after a task row transition per scheduling.DecideNotifyAfterReadyTransition.
//
//funclogmeasure:skip category=hot-path reason="Scheduling orchestration; BC store and worker paths emit operation traces."
func ApplyNotifyDecision(ctx context.Context, r *Runtime, task taskcoredomain.Task, d scheduling.NotifyDecision) {
	if r == nil {
		return
	}
	if d.ScheduleWake != nil {
		r.PickupWake.SchedulePickupWake(ctx, task.ID, *d.ScheduleWake)
		return
	}
	if d.CancelWake {
		r.PickupWake.CancelPickupWake(task.ID)
	}
	if d.NotifyQueue {
		r.Notify.Notify(ctx, task)
	}
}

// NotifyReadyOnDevMirror applies dev-mirror ready transitions (simpler than full scheduling decision).
//
//funclogmeasure:skip category=hot-path reason="Dev-mirror scheduling helper; devsim and store paths emit operation traces."
func NotifyReadyOnDevMirror(ctx context.Context, r *Runtime, task *taskcoredomain.Task, prev taskcoredomain.Status, now time.Time) {
	if r == nil || task == nil {
		return
	}
	if task.Status != taskcoredomain.StatusReady || prev == taskcoredomain.StatusReady {
		return
	}
	if task.PickupNotBefore != nil && task.PickupNotBefore.After(now) {
		r.PickupWake.SchedulePickupWake(ctx, task.ID, *task.PickupNotBefore)
		return
	}
	if scheduling.ShouldNotifyReadyNow(task.PickupNotBefore, now) {
		r.Notify.Notify(ctx, *task)
	}
}
