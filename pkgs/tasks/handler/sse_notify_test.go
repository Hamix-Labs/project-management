package handler

import (
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func TestHandler_publishPolicyEvent_stripsDataOnHintOnly(t *testing.T) {
	h, _, hub := newWritepolicyTestHandler(t)
	ch, unsub := hub.Subscribe()
	defer unsub()

	h.notifyTaskChanged(realtime.TaskDeleted, "task-1", map[string]string{"title": "leak"})

	events := drainSSE(t, ch, 1, 2*time.Second)
	mustEqualEvents(t, "hint-only strips data", summarize(events), []string{"task_deleted:task-1"})
	for _, ev := range events {
		if ev.Data != nil {
			t.Fatalf("hint-only frame carried data: %#v", ev.Data)
		}
	}
}

func TestHandler_notifyTaskChanged_carriesDataOnEnrichedType(t *testing.T) {
	h, _, hub := newWritepolicyTestHandler(t)
	ch, unsub := hub.Subscribe()
	defer unsub()

	payload := map[string]string{"id": "task-1", "title": "enriched"}
	h.notifyTaskChanged(realtime.TaskUpdated, "task-1", payload)

	events := drainSSE(t, ch, 1, 2*time.Second)
	mustHaveTaskUpdatedData(t, "notifyTaskChanged enriched", events, "task-1")
}

func TestHandler_notifyScopelessChange_publishesIdLessFrame(t *testing.T) {
	tests := []struct {
		name string
		typ  realtime.ChangeType
		want string
	}{
		{name: "settings_changed", typ: realtime.SettingsChanged, want: "settings_changed:"},
		{name: "agent_run_cancelled", typ: realtime.AgentRunCancelled, want: "agent_run_cancelled:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _, hub := newWritepolicyTestHandler(t)
			ch, unsub := hub.Subscribe()
			defer unsub()

			h.notifyScopelessChange(tt.typ)

			got := summarize(drainSSE(t, ch, 1, 2*time.Second))
			mustEqualEvents(t, "notifyScopelessChange/"+tt.name, got, []string{tt.want})
		})
	}
}

func TestHandler_notifyScopelessChange_nilHubNoop(t *testing.T) {
	h, _, _ := newWritepolicyTestHandler(t)
	h.hub = nil
	h.notifyScopelessChange(realtime.SettingsChanged)
}
