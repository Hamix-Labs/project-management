package worker_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

type agentPickupFailStore struct {
	*composition.API
}

func (s *agentPickupFailStore) AgentPickup(ctx context.Context, taskID string, by taskcoredomain.Actor) (*taskcorecontract.AgentPickupResult, error) {
	return nil, fmt.Errorf("save task: simulated persistence failure")
}

func TestWorker_pickupPersistenceFailure_defersAndRecordsEvent(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "pickup-fail")
	wrapped := &agentPickupFailStore{API: h.store}
	w := worker.NewWorker(wrapped, h.queue, runnerfake.New(), worker.Options{Notifier: h.notifier})

	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		events, err := h.store.ListTaskEvents(context.Background(), tsk.ID)
		if err == nil {
			for _, ev := range events {
				if ev.Type == taskeventsdomain.EventTaskPickupFailed {
					goto verified
				}
			}
		}
		time.Sleep(pollInterval)
	}
	t.Fatal("timeout waiting for task_pickup_failed event")

verified:
	cancel()
	<-done

	got, err := h.store.Get(context.Background(), tsk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != taskcoredomain.StatusReady {
		t.Fatalf("status=%q want ready", got.Status)
	}
	if got.PickupNotBefore == nil {
		t.Fatal("expected pickup_not_before defer after persistence failure")
	}
}
