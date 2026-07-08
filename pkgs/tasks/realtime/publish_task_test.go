package realtime_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

type recordingPublisher struct {
	events []realtime.Event
}

func (r *recordingPublisher) Publish(ev realtime.Event) {
	r.events = append(r.events, ev)
}

func TestPublishEnrichedTaskUpdated_includesData(t *testing.T) {
	t.Parallel()
	pub := &recordingPublisher{}
	taskTree := map[string]string{"id": "task-1", "status": "done"}
	err := realtime.PublishEnrichedTaskUpdated(context.Background(), pub, func(_ context.Context, id string) (any, error) {
		if id != "task-1" {
			t.Fatalf("task id = %q, want task-1", id)
		}
		return taskTree, nil
	}, "task-1")
	if err != nil {
		t.Fatalf("PublishEnrichedTaskUpdated: %v", err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("events: got %d want 1", len(pub.events))
	}
	ev := pub.events[0]
	if ev.Type != realtime.TaskUpdated {
		t.Fatalf("type = %q, want task_updated", ev.Type)
	}
	if ev.ID != "task-1" {
		t.Fatalf("id = %q, want task-1", ev.ID)
	}
	if ev.Data == nil {
		t.Fatal("expected enriched data")
	}
}

func TestPublishEnrichedTaskUpdated_loadError(t *testing.T) {
	t.Parallel()
	pub := &recordingPublisher{}
	sentinel := errors.New("not found")
	err := realtime.PublishEnrichedTaskUpdated(context.Background(), pub, func(context.Context, string) (any, error) {
		return nil, sentinel
	}, "task-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if len(pub.events) != 0 {
		t.Fatalf("expected no publish on load error, got %d", len(pub.events))
	}
}

func TestPublishEnrichedTaskUpdated_nilPublisherNoOp(t *testing.T) {
	t.Parallel()
	err := realtime.PublishEnrichedTaskUpdated(context.Background(), nil, func(context.Context, string) (any, error) {
		t.Fatal("load should not run when publisher is nil")
		return nil, nil
	}, "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
