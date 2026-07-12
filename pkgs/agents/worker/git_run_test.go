package worker_test

import (
	"context"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestWorker_missingGitBinding_defersPickup(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "unbound",
		InitialPrompt: "do the thing",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, done := h.startWorker(ctx, runnerfake.New(), worker.Options{})
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	got, err := h.store.Get(context.Background(), tsk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != taskcoredomain.StatusReady {
		t.Fatalf("status=%q want ready (missing binding should defer, not run)", got.Status)
	}
	if got.PickupNotBefore == nil {
		t.Fatal("expected pickup_not_before defer")
	}
}
