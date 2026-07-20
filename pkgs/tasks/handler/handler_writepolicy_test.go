package handler

import (
	"context"
	"testing"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"

	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/policy"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func newWritepolicyTestHandler(t *testing.T) (*Handler, *composition.API, *realtime.SSEHub) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	hub := realtime.NewSSEHub()
	return &Handler{store: st, hub: hub, repoProv: NewStaticRepoProvider(nil)}, st, hub
}

func TestWritepolicy_publishClassification(t *testing.T) {
	t.Parallel()
	tests := []struct {
		typ      realtime.ChangeType
		enriched bool
		hintOnly bool
	}{
		{realtime.TaskCreated, true, false},
		{realtime.TaskUpdated, true, false},
		{realtime.TaskDeleted, false, true},
		{realtime.TaskGateChanged, false, true},
		{realtime.TaskDependencyChanged, false, true},
		{realtime.TaskCycleChanged, false, false},
	}
	for _, tc := range tests {
		if got := policy.EnrichedTaskChangeEvent(tc.typ); got != tc.enriched {
			t.Errorf("EnrichedTaskChangeEvent(%q) = %v, want %v", tc.typ, got, tc.enriched)
		}
		if got := policy.IsHintOnly(tc.typ); got != tc.hintOnly {
			t.Errorf("IsHintOnly(%q) = %v, want %v", tc.typ, got, tc.hintOnly)
		}
	}
}

func TestHandler_notifyTaskUpdatedEnriched_storeErrorNeverPublishes(t *testing.T) {
	h, _, hub := newWritepolicyTestHandler(t)

	ch, unsub := hub.Subscribe()
	defer unsub()

	err := h.notifyTaskUpdatedEnriched(context.Background(), "11111111-1111-4111-8111-111111111111")
	if err == nil {
		t.Fatal("expected error for missing task")
	}
	got := summarize(drainSSE(t, ch, 0, 200*time.Millisecond))
	if len(got) != 0 {
		t.Fatalf("drained SSE %v after store Get failure; want zero publishes", got)
	}
}

func TestHandler_notifyTaskUpdatedEnriched_publishesData(t *testing.T) {
	h, st, hub := newWritepolicyTestHandler(t)

	task, err := st.Create(context.Background(), taskcorestore.CreateTaskInput{
		Title:         "enriched notify",
		InitialPrompt: "p",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	ch, unsub := hub.Subscribe()
	defer unsub()

	if err := h.notifyTaskUpdatedEnriched(context.Background(), task.ID); err != nil {
		t.Fatal(err)
	}
	events := drainSSE(t, ch, 1, 2*time.Second)
	mustEqualEvents(t, "notifyTaskUpdatedEnriched", summarize(events), []string{"task_updated:" + task.ID})
	mustHaveTaskUpdatedData(t, "notifyTaskUpdatedEnriched", events, task.ID)
}
