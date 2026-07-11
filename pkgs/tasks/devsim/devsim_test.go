package devsim

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestEventCycle_exhaustive(t *testing.T) {
	want := map[taskeventsdomain.EventType]struct{}{
		taskeventsdomain.EventTaskCreated:           {},
		taskeventsdomain.EventStatusChanged:         {},
		taskeventsdomain.EventPriorityChanged:       {},
		taskeventsdomain.EventPromptAppended:        {},
		taskeventsdomain.EventContextAdded:          {},
		taskeventsdomain.EventConstraintAdded:       {},
		taskeventsdomain.EventSuccessCriterionAdded: {},
		taskeventsdomain.EventNonGoalAdded:          {},
		taskeventsdomain.EventPlanAdded:             {},
		taskeventsdomain.EventChecklistItemAdded:    {},
		taskeventsdomain.EventChecklistItemToggled:  {},
		taskeventsdomain.EventChecklistItemUpdated:  {},
		taskeventsdomain.EventChecklistItemRemoved:  {},
		taskeventsdomain.EventMessageAdded:          {},
		taskeventsdomain.EventArtifactAdded:         {},
		taskeventsdomain.EventApprovalRequested:     {},
		taskeventsdomain.EventApprovalGranted:       {},
		taskeventsdomain.EventTaskCompleted:         {},
		taskeventsdomain.EventTaskFailed:            {},
		taskeventsdomain.EventSyncPing:              {},
	}
	if len(EventCycle) != len(want) {
		t.Fatalf("EventCycle len %d want %d", len(EventCycle), len(want))
	}
	seen := make(map[taskeventsdomain.EventType]int)
	for _, e := range EventCycle {
		seen[e]++
		delete(want, e)
	}
	if len(want) != 0 {
		t.Fatalf("missing event types in cycle: %v", want)
	}
	for e, n := range seen {
		if n != 1 {
			t.Fatalf("duplicate %q count %d", e, n)
		}
	}
}

func TestPersistAllTasks_emitsOnePublishPerTask(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	ctx := context.Background()
	a, err := st.Create(ctx, store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "a"}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	b, err := st.Create(ctx, store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "b"}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	var lines []ChangeKind
	publish := func(k ChangeKind, id string) {
		lines = append(lines, k)
		_ = id
	}

	PersistAllTasks(ctx, st, Options{}, publish)

	if len(lines) != 2 {
		t.Fatalf("want 2 publish calls, got %d (%v)", len(lines), lines)
	}
	for _, k := range lines {
		if k != ChangeUpdated {
			t.Fatalf("want ChangeUpdated, got %v", k)
		}
	}
	for _, id := range []string{a.ID, b.ID} {
		tsk, err := st.Get(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if tsk.Status != taskcoredomain.StatusReady {
			t.Fatalf("task %s status = %q want ready (ticker does not patch task row)", id, tsk.Status)
		}
		evs, err := st.ListTaskEvents(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if len(evs) != 2 {
			t.Fatalf("task %s: want 2 events, got %d", id, len(evs))
		}
		last := evs[len(evs)-1]
		if last.Type != EventCycle[1] {
			t.Fatalf("task %s: last event type = %q want %q", id, last.Type, EventCycle[1])
		}
		if last.By != taskcoredomain.ActorAgent {
			t.Fatalf("task %s: last event by = %q want agent", id, last.By)
		}
	}
}

func TestPersistAllTasks_burst_emitsMultiplePublishes(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	ctx := context.Background()
	if _, err := st.Create(ctx, store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "solo"}, taskcoredomain.ActorUser); err != nil {
		t.Fatal(err)
	}
	var n int
	PersistAllTasks(ctx, st, Options{EventsPerTick: 3}, func(ChangeKind, string) { n++ })
	if n != 3 {
		t.Fatalf("want 3 publishes, got %d", n)
	}
}

func TestPersistAllTasks_syncRow_updatesStatus(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	ctx := context.Background()
	tsk, err := st.Create(ctx, store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "x"}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	PersistAllTasks(ctx, st, Options{SyncTaskRow: true}, func(ChangeKind, string) {})
	got, err := st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusRunning {
		t.Fatalf("status %q want running after mirror", got.Status)
	}
}

func TestPersistAllTasks_userResponse_appendsThread(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	ctx := context.Background()
	tsk, err := st.Create(ctx, store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "t"}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	reached := false
	for range 200 {
		n, err := st.TaskEventCount(ctx, tsk.ID)
		if err != nil {
			t.Fatal(err)
		}
		if nextEventTypeFromCount(n) == taskeventsdomain.EventApprovalRequested {
			reached = true
			break
		}
		PersistAllTasks(ctx, st, Options{}, func(ChangeKind, string) {})
	}
	if !reached {
		t.Fatal("did not reach approval_requested in event cycle")
	}
	PersistAllTasks(ctx, st, Options{UserResponse: true}, func(ChangeKind, string) {})
	evs, err := st.ListTaskEvents(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	last := evs[len(evs)-1]
	if last.Type != taskeventsdomain.EventApprovalRequested {
		t.Fatalf("last type %q want approval_requested", last.Type)
	}
	entries := store.ThreadEntriesForDisplay(&last)
	if len(entries) == 0 {
		t.Fatal("expected response thread entries")
	}
	if entries[len(entries)-1].By != taskcoredomain.ActorUser {
		t.Fatalf("want user message, got %+v", entries[len(entries)-1])
	}
}

func TestRunTicker_StopsOnContextCancel(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	if _, err := st.Create(context.Background(), store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "pre-cancel"}, taskcoredomain.ActorUser); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var ticks atomic.Int64
	publish := func(ChangeKind, string) { ticks.Add(1) }

	RunTicker(ctx, st, time.Second, Options{}, publish)
	cancel()

	// After cancel, give the goroutine a generous window to observe ctx.Done()
	// and exit instead of firing on the next tick. With ticker interval 1s,
	// waiting > 2s is enough to detect a leak: any tick that fires after
	// cancellation would publish at least once for the seeded task.
	time.Sleep(2500 * time.Millisecond)
	if got := ticks.Load(); got != 0 {
		t.Fatalf("ticker fired %d publish(es) after ctx cancel; goroutine did not honor ctx.Done()", got)
	}
}

func TestRunTicker_NoOpOnInvalidArgs(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	publish := func(ChangeKind, string) {}

	RunTicker(context.Background(), nil, time.Second, Options{}, publish)
	RunTicker(context.Background(), st, 500*time.Millisecond, Options{}, publish)
	RunTicker(context.Background(), st, time.Second, Options{}, nil)
}

func TestSamplePayload_JSON(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	ctx := context.Background()
	task, err := st.Create(ctx, store.CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "x"}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	var saw bool
	PersistAllTasks(ctx, st, Options{}, func(ChangeKind, string) { saw = true })
	if !saw {
		t.Fatal("expected publish")
	}
	evs, err := st.ListTaskEvents(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("want 2 events, got %d", len(evs))
	}
	last := evs[len(evs)-1]
	if last.Type != taskeventsdomain.EventStatusChanged {
		t.Fatalf("last type %q want %q", last.Type, taskeventsdomain.EventStatusChanged)
	}
	var m map[string]string
	if err := json.Unmarshal(last.Data, &m); err != nil {
		t.Fatal(err)
	}
	if m["from"] != "ready" || m["to"] != "running" {
		t.Fatalf("got %+v", m)
	}
}
