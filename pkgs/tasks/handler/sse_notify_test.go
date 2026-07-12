package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"testing"
	"time"
)

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
