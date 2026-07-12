package composition

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/storehooks"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/scheduling"
)

//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; scheduling traces at BC chokepoints."
func (a *API) schedulePickupWake(ctx context.Context, taskID string, notBefore time.Time) {
	if a == nil || a.hooks == nil {
		return
	}
	a.hooks.PickupWake.SchedulePickupWake(ctx, taskID, notBefore)
}

//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; scheduling traces at BC chokepoints."
func (a *API) cancelPickupWake(taskID string) {
	if a == nil || a.hooks == nil {
		return
	}
	a.hooks.PickupWake.CancelPickupWake(taskID)
}

//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; notify traces at BC chokepoints."
func (a *API) notifyReadyTask(ctx context.Context, task taskcoredomain.Task) {
	if a == nil || a.hooks == nil {
		return
	}
	a.hooks.Notify.Notify(ctx, task)
}

//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; scheduling traces at BC chokepoints."
func (a *API) applyNotifyDecision(ctx context.Context, task taskcoredomain.Task, d scheduling.NotifyDecision) {
	storehooks.ApplyNotifyDecision(ctx, a.hooks, task, d)
}
