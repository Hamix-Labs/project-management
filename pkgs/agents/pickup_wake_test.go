package agents

import (
	"context"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestPickupWakeScheduler_WakeEnqueuesNearFutureTask(t *testing.T) {
	ctx := context.Background()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	q := NewMemoryQueue(32)
	st.SetReadyTaskNotifier(q)
	w := NewPickupWakeScheduler(st, q)
	st.SetPickupWake(w)
	defer w.Stop()

	wtID := "wt-wake-1"
	future := time.Now().UTC().Add(40 * time.Millisecond)
	tk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "wake-test", Priority: taskcoredomain.PriorityMedium,
		PickupNotBefore: &future,
		WorktreeID:      &wtID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if q.BufferDepth() != 0 {
		t.Fatalf("buffer depth before wake: %d want 0", q.BufferDepth())
	}

	deadline := time.After(2 * time.Second)
	for q.BufferDepth() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for pickup wake to enqueue")
		case <-time.After(5 * time.Millisecond):
		}
	}
	got, err := q.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != tk.ID {
		t.Fatalf("task id %q want %q", got.ID, tk.ID)
	}
	q.AckAfterRecv(got.ID)
}

func TestPickupWakeScheduler_CancelPreventsWake(t *testing.T) {
	ctx := context.Background()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	q := NewMemoryQueue(32)
	st.SetReadyTaskNotifier(q)
	w := NewPickupWakeScheduler(st, q)
	st.SetPickupWake(w)
	defer w.Stop()

	future := time.Now().UTC().Add(200 * time.Millisecond)
	wtID := "wt-wake-cancel"
	tk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "cancel-test", Priority: taskcoredomain.PriorityMedium,
		PickupNotBefore: &future,
		WorktreeID:      &wtID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	w.Cancel(tk.ID)

	select {
	case got := <-q.Recv():
		t.Fatalf("unexpected enqueue for cancelled task: %v", got.ID)
	case <-time.After(300 * time.Millisecond):
	}
}

func TestPickupWakeScheduler_QueueFullReschedulesWake(t *testing.T) {
	ctx := context.Background()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	q := NewMemoryQueue(1)
	st.SetReadyTaskNotifier(q)
	w := NewPickupWakeScheduler(st, q)
	st.SetPickupWake(w)
	defer w.Stop()

	filler := taskcoredomain.Task{ID: "filler-hold-slot", Status: taskcoredomain.StatusReady, Title: "filler"}
	if err := q.NotifyReadyTask(ctx, filler); err != nil {
		t.Fatalf("fill queue: %v", err)
	}

	future := time.Now().UTC().Add(40 * time.Millisecond)
	wtID := "wt-wake-full"
	tk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "queue-full-wake", Priority: taskcoredomain.PriorityMedium,
		PickupNotBefore: &future,
		WorktreeID:      &wtID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	// First wake should hit ErrQueueFull and re-schedule; give it time to fire.
	time.Sleep(200 * time.Millisecond)
	if q.BufferDepth() != 1 {
		t.Fatalf("buffer depth while full: %d want 1 (only filler)", q.BufferDepth())
	}

	got, err := q.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != filler.ID {
		t.Fatalf("first receive id %q want filler %q", got.ID, filler.ID)
	}
	q.AckAfterRecv(got.ID)

	deadline := time.After(3 * time.Second)
	for q.BufferDepth() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for queue-full backoff wake to enqueue")
		case <-time.After(20 * time.Millisecond):
		}
	}
	woken, err := q.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if woken.ID != tk.ID {
		t.Fatalf("woken task id %q want %q", woken.ID, tk.ID)
	}
	q.AckAfterRecv(woken.ID)
}
