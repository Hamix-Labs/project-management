package handler

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

func (h *Handler) notifyChange(typ TaskChangeType, id string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyChange", "change_type", typ)
	if h.hub == nil || id == "" {
		return
	}
	h.hub.Publish(TaskChangeEvent{Type: typ, ID: id})
}

func (h *Handler) notifyTaskChanged(typ TaskChangeType, id string, data any) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyTaskChanged", "change_type", typ)
	if h.hub == nil || id == "" {
		return
	}
	h.hub.Publish(TaskChangeEvent{Type: typ, ID: id, Data: data})
}

func (h *Handler) notifyCycleChange(taskID, cycleID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyCycleChange", "task_id", taskID, "cycle_id", cycleID)
	if h.hub == nil || taskID == "" || cycleID == "" {
		return
	}
	h.hub.Publish(TaskChangeEvent{Type: TaskCycleChanged, ID: taskID, CycleID: cycleID})
}

func (h *Handler) notifyCycleChanged(taskID, cycleID string, data any) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyCycleChanged", "task_id", taskID, "cycle_id", cycleID)
	if h.hub == nil || taskID == "" || cycleID == "" {
		return
	}
	h.hub.Publish(TaskChangeEvent{Type: TaskCycleChanged, ID: taskID, CycleID: cycleID, Data: data})
}

func (h *Handler) notifyTaskEventChanged(taskID string, eventSeq int64) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyTaskEventChanged", "task_id", taskID, "event_seq", eventSeq)
	if h.hub == nil || taskID == "" || eventSeq < 1 {
		return
	}
	h.hub.Publish(TaskChangeEvent{Type: TaskEventChanged, ID: taskID, EventSeq: eventSeq})
}

func (h *Handler) notifyScopelessChange(typ TaskChangeType) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyScopelessChange", "change_type", typ)
	if h.hub == nil {
		return
	}
	h.hub.Publish(TaskChangeEvent{Type: typ})
}
