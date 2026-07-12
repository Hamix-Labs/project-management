package devsim

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"log/slog"
)

const listPage = 200 // store.ListFlat maximum page size

// EventCycle is the full set of taskeventsdomain.EventType values used by the dev ticker, in display order.
// Keep in sync with taskevents/domain event types (every EventType exactly once).
// Index 0 is chosen when len(events)%len(cycle)==0; index 1 is the first tick after task_created.
var EventCycle = []taskeventsdomain.EventType{
	taskeventsdomain.EventTaskCreated,
	taskeventsdomain.EventStatusChanged,
	taskeventsdomain.EventPriorityChanged,
	taskeventsdomain.EventPromptAppended,
	taskeventsdomain.EventMessageAdded,
	taskeventsdomain.EventContextAdded,
	taskeventsdomain.EventConstraintAdded,
	taskeventsdomain.EventSuccessCriterionAdded,
	taskeventsdomain.EventNonGoalAdded,
	taskeventsdomain.EventPlanAdded,
	taskeventsdomain.EventChecklistItemAdded,
	taskeventsdomain.EventChecklistItemToggled,
	taskeventsdomain.EventChecklistItemUpdated,
	taskeventsdomain.EventChecklistItemRemoved,
	taskeventsdomain.EventArtifactAdded,
	taskeventsdomain.EventApprovalRequested,
	taskeventsdomain.EventApprovalGranted,
	taskeventsdomain.EventTaskCompleted,
	taskeventsdomain.EventTaskFailed,
	taskeventsdomain.EventSyncPing,
}

func samplePayloadForType(typ taskeventsdomain.EventType) ([]byte, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "devsim.samplePayloadForType", "type", typ)
	if f, ok := samplePayloadByType[typ]; ok {
		return f()
	}
	return json.Marshal(map[string]string{"dev_sample": string(typ)})
}

func nextEventTypeFromCount(n int64) taskeventsdomain.EventType {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "devsim.nextEventTypeFromCount")
	if len(EventCycle) == 0 {
		return taskeventsdomain.EventSyncPing
	}
	idx := int(n % int64(len(EventCycle)))
	return EventCycle[idx]
}

func persistSampleEvent(ctx context.Context, st *composition.API, t *taskcoredomain.Task, opts Options, publish func(ChangeKind, string)) error {
	if st == nil || t == nil {
		return errors.New("store or task nil")
	}
	if publish == nil {
		return errors.New("publish nil")
	}
	n, err := st.TaskEventCount(ctx, t.ID)
	if err != nil {
		return err
	}
	typ := nextEventTypeFromCount(n)
	payload, err := samplePayloadForType(typ)
	if err != nil {
		return err
	}
	if err := st.AppendTaskEvent(ctx, t.ID, typ, taskcoredomain.ActorAgent, payload); err != nil {
		return err
	}
	if opts.SyncTaskRow {
		if err := st.ApplyDevTaskRowMirror(ctx, t.ID, typ, payload); err != nil {
			slog.Debug("sse dev mirror skipped", "cmd", calltrace.LogCmd, "operation", "devsim.mirror_task",
				"task_id", t.ID, "type", typ, "err", err)
		}
	}
	if opts.UserResponse && taskeventsdomain.EventTypeAcceptsUserResponse(typ) {
		seq, err := st.LastEventSeq(ctx, t.ID)
		if err != nil {
			return err
		}
		if seq < 1 {
			return errors.New("no events after append")
		}
		msg := "Synthetic user reply (devsim)."
		if typ == taskeventsdomain.EventTaskFailed {
			msg = "Synthetic triage note (devsim)."
		}
		if err := st.AppendTaskEventResponseMessage(ctx, t.ID, seq, msg, taskcoredomain.ActorUser); err != nil {
			slog.Debug("sse dev user_response skipped", "cmd", calltrace.LogCmd, "operation", "devsim.user_response",
				"task_id", t.ID, "seq", seq, "err", err)
		}
	}
	publish(ChangeUpdated, t.ID)
	return nil
}

// PersistAllTasks walks every task using store.ListFlat (id ASC, paginated), same data as flat list.
// For each task it appends up to opts.EventsPerTick sample audit events and invokes publish after each successful cycle.
func PersistAllTasks(ctx context.Context, st *composition.API, opts Options, publish func(ChangeKind, string)) {
	if st == nil || publish == nil {
		return
	}
	per := opts.EventsPerTick
	if per < 1 {
		per = 1
	}
	if per > maxEventsPerTick {
		per = maxEventsPerTick
	}
	for offset := 0; ; offset += listPage {
		rows, err := st.ListFlat(ctx, listPage, offset, nil)
		if err != nil {
			slog.Debug("sse dev ticker list failed", "cmd", calltrace.LogCmd, "operation", "devsim.tick_list", "err", err)
			return
		}
		for i := range rows {
			for range per {
				if err := persistSampleEvent(ctx, st, &rows[i], opts, publish); err != nil {
					slog.Debug("sse dev ticker task skipped", "cmd", calltrace.LogCmd, "operation", "devsim.tick_task",
						"task_id", rows[i].ID, "err", err)
					break
				}
			}
		}
		if len(rows) < listPage {
			return
		}
	}
}
