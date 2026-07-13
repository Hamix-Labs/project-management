package agentworker

import (
	"context"
	"encoding/json"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"sync"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

type recordingPublisher struct {
	events []realtime.Event
}

func (r *recordingPublisher) Publish(ev realtime.Event) {
	r.events = append(r.events, ev)
}

func waitForEvents(t *testing.T, pub *recordingPublisher, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for len(pub.events) < want && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if len(pub.events) != want {
		t.Fatalf("events: got %d want %d", len(pub.events), want)
	}
}

func TestTaskUpdatedSSEAdapter_publishesEnrichedWireShape(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	ctx := context.Background()

	created, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "sse-task-updated",
		Priority: taskcoredomain.PriorityMedium,
		Status:   taskcoredomain.StatusDone,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	pub := &recordingPublisher{}
	adapter := newTaskUpdatedSSEAdapter(pub, st, nil)
	adapter.PublishTaskUpdated(created.ID)
	waitForEvents(t, pub, 1, 2*time.Second)

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

type blockingPublisher struct {
	mu      sync.Mutex
	block   chan struct{}
	events  []realtime.Event
	unblock chan struct{}
}

func newBlockingPublisher() *blockingPublisher {
	return &blockingPublisher{
		block:   make(chan struct{}),
		unblock: make(chan struct{}, 1),
	}
}

func (b *blockingPublisher) Publish(ev realtime.Event) {
	b.mu.Lock()
	b.events = append(b.events, ev)
	b.mu.Unlock()
	select {
	case <-b.unblock:
	case <-b.block:
	}
}

func TestTaskUpdatedSSEAdapter_returnsWithoutBlockingOnSlowPublisher(t *testing.T) {
	t.Parallel()
	pub := newBlockingPublisher()
	adapter := newTaskUpdatedSSEAdapter(pub, &slowTaskGetter{}, nil)

	start := time.Now()
	adapter.PublishTaskUpdated("task-1")
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("PublishTaskUpdated blocked for %v, want <100ms", elapsed)
	}
	close(pub.block)
}

type slowTaskGetter struct{}

func (slowTaskGetter) Get(context.Context, string) (*taskcoredomain.Task, error) {
	time.Sleep(200 * time.Millisecond)
	return &taskcoredomain.Task{ID: "task-1", Status: taskcoredomain.StatusReady}, nil
}

func TestTaskUpdatedSSEAdapter_dropsWhenQueueFull(t *testing.T) {
	pub := &recordingPublisher{}
	metrics := &fakeNotifierMetrics{}
	block := make(chan struct{})
	started := make(chan struct{})
	getter := &blockingTaskGetter{block: block, started: started}
	adapter := newTaskUpdatedSSEAdapter(pub, getter, metrics)

	adapter.PublishTaskUpdated("task-1")
	<-started
	for i := 0; i < taskUpdatedQueueDepth; i++ {
		adapter.PublishTaskUpdated("task-fill")
	}

	beforeOverflow := metrics.dropped
	overflowStart := time.Now()
	adapter.PublishTaskUpdated("task-overflow")
	if elapsed := time.Since(overflowStart); elapsed > 100*time.Millisecond {
		t.Fatalf("PublishTaskUpdated blocked for %v, want <100ms", elapsed)
	}
	if metrics.dropped <= beforeOverflow {
		t.Fatal("overflow publish should drop when queue is full")
	}
	close(block)
}

type blockingTaskGetter struct {
	block   chan struct{}
	started chan struct{}
	once    sync.Once
}

func (g *blockingTaskGetter) Get(context.Context, string) (*taskcoredomain.Task, error) {
	g.once.Do(func() {
		if g.started != nil {
			close(g.started)
		}
	})
	<-g.block
	return &taskcoredomain.Task{ID: "task-1", Status: taskcoredomain.StatusReady}, nil
}

type fakeNotifierMetrics struct {
	dropped int
}

func (f *fakeNotifierMetrics) RecordNotifierDropped(string) {
	f.dropped++
}
