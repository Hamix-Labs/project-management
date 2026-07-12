package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"log/slog"
	"time"
)

func (a *API) ApplyDevTaskRowMirror(ctx context.Context, taskID string, typ taskeventsdomain.EventType, data []byte) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ApplyDevTaskRowMirror")
	nt, prev, err := a.taskcore.ApplyDevTaskRowMirror(ctx, taskID, typ, data)
	if err != nil {
		return err
	}
	if nt == nil {
		return nil
	}
	if nt.Status == taskcoredomain.StatusReady && prev != taskcoredomain.StatusReady {
		now := time.Now().UTC()
		if nt.PickupNotBefore != nil && nt.PickupNotBefore.After(now) {
			a.schedulePickupWake(ctx, nt.ID, *nt.PickupNotBefore)
		} else if ShouldNotifyReadyNow(nt.PickupNotBefore, now) {
			a.notifyReadyTask(ctx, *nt)
		}
	}
	return nil
}

func (a *API) ListDevsimTasks(ctx context.Context, idLikePattern string) ([]taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListDevsimTasks")
	return a.taskcore.ListDevsimTasks(ctx, idLikePattern)
}
