package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"log/slog"
	"time"
)

func (s *Store) ApplyDevTaskRowMirror(ctx context.Context, taskID string, typ taskeventsdomain.EventType, data []byte) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ApplyDevTaskRowMirror")
	nt, prev, err := s.taskcore.ApplyDevTaskRowMirror(ctx, taskID, typ, data)
	if err != nil {
		return err
	}
	if nt == nil {
		return nil
	}
	if nt.Status == taskcoredomain.StatusReady && prev != taskcoredomain.StatusReady {
		now := time.Now().UTC()
		if nt.PickupNotBefore != nil && nt.PickupNotBefore.After(now) {
			s.schedulePickupWake(ctx, nt.ID, *nt.PickupNotBefore)
		} else if ShouldNotifyReadyNow(nt.PickupNotBefore, now) {
			s.notifyReadyTask(ctx, *nt)
		}
	}
	return nil
}

func (s *Store) ListDevsimTasks(ctx context.Context, idLikePattern string) ([]taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListDevsimTasks")
	return s.taskcore.ListDevsimTasks(ctx, idLikePattern)
}
