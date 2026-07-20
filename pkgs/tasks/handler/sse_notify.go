package handler

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/writepolicy"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func (h *Handler) notifyChange(typ realtime.ChangeType, id string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyChange", "change_type", typ)
	if h.hub == nil || id == "" {
		return
	}
	if writepolicy.EnrichedTaskChangeEvent(typ) {
		slog.Warn("notifyChange called for enriched SSE type; publishing id-only hint",
			"change_type", typ, "task_id", id)
	}
	h.publishPolicyEvent(realtime.Event{Type: typ, ID: id})
}

func (h *Handler) notifyTaskChanged(typ realtime.ChangeType, id string, data any) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyTaskChanged", "change_type", typ)
	if h.hub == nil || id == "" {
		return
	}
	ev := realtime.Event{Type: typ, ID: id}
	if writepolicy.EnrichedTaskChangeEvent(typ) {
		if data != nil {
			ev.Data = data
		}
	} else if !writepolicy.IsHintOnly(typ) && data != nil {
		ev.Data = data
	}
	h.publishPolicyEvent(ev)
}

func (h *Handler) notifyCycleChange(taskID, cycleID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyCycleChange", "task_id", taskID, "cycle_id", cycleID)
	if h.hub == nil || taskID == "" || cycleID == "" {
		return
	}
	h.publishPolicyEvent(realtime.Event{Type: realtime.TaskCycleChanged, ID: taskID, CycleID: cycleID})
}

func (h *Handler) notifyCycleChanged(taskID, cycleID string, data any) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyCycleChanged", "task_id", taskID, "cycle_id", cycleID)
	if h.hub == nil || taskID == "" || cycleID == "" {
		return
	}
	ev := realtime.Event{Type: realtime.TaskCycleChanged, ID: taskID, CycleID: cycleID}
	if data != nil {
		ev.Data = data
	}
	h.publishPolicyEvent(ev)
}

func (h *Handler) notifyTaskEventChanged(taskID string, eventSeq int64) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyTaskEventChanged", "task_id", taskID, "event_seq", eventSeq)
	if h.hub == nil || taskID == "" || eventSeq < 1 {
		return
	}
	h.publishPolicyEvent(realtime.Event{Type: realtime.TaskEventChanged, ID: taskID, EventSeq: eventSeq})
}

func (h *Handler) notifyScopelessChange(typ realtime.ChangeType) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyScopelessChange", "change_type", typ)
	if h.hub == nil {
		return
	}
	if !writepolicy.IsScopelessHint(typ) {
		slog.Warn("notifyScopelessChange called for scoped SSE type", "change_type", typ)
	}
	h.publishPolicyEvent(realtime.Event{Type: typ})
}

// publishPolicyEvent is the runtime choke for handler SSE publishes. It strips
// Data from hint-only types and preserves enriched / cycle-specific frames.
//
//funclogmeasure:skip category=delegate-already-logs reason="SSE publish choke; operation trace is emitted by notify* callers."
func (h *Handler) publishPolicyEvent(ev realtime.Event) {
	if h.hub == nil {
		return
	}
	if ev.ID != "" && writepolicy.IsHintOnly(ev.Type) {
		ev.Data = nil
	}
	h.hub.Publish(ev)
}
