package agentworker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

type recordingPublisher struct {
	events []realtime.Event
}

func (r *recordingPublisher) Publish(ev realtime.Event) {
	r.events = append(r.events, ev)
}

func TestTaskUpdatedSSEAdapter_publishesEnrichedWireShape(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	ctx := context.Background()

	created, err := st.Create(ctx, store.CreateTaskInput{
		Title:    "sse-task-updated",
		Priority: taskcoredomain.PriorityMedium,
		Status:   taskcoredomain.StatusDone,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	pub := &recordingPublisher{}
	adapter := newTaskUpdatedSSEAdapter(pub, st)
	adapter.PublishTaskUpdated(created.ID)

	if len(pub.events) != 1 {
		t.Fatalf("events: got %d want 1", len(pub.events))
	}
	ev := pub.events[0]
	if ev.Type != realtime.TaskUpdated {
		t.Fatalf("type = %q, want task_updated", ev.Type)
	}
	if ev.ID != created.ID {
		t.Fatalf("id = %q, want %q", ev.ID, created.ID)
	}
	if ev.Data == nil {
		t.Fatal("expected enriched data")
	}
	raw, err := json.Marshal(ev.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var wire struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("unmarshal task data: %v", err)
	}
	if wire.Status != string(taskcoredomain.StatusDone) {
		t.Fatalf("data.status = %q, want done", wire.Status)
	}
}
