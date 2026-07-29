package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/notifierfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

// TestWorker_AgentPickup_publishesTaskUpdated pins the contract that
// ready→running admission publishes enriched task_updated so the SPA
// detail toolbar updates without a hard reload (ADR-0026 S5).
func TestWorker_AgentPickup_publishesTaskUpdated(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "pickup-publishes")
	taskUpdated := notifierfake.NewRecordingTaskUpdatedNotifier()
	br := newBlockingRunner()

	_, done := h.startWorker(ctx, br, worker.Options{
		TaskUpdatedNotifier: taskUpdated,
	})

	select {
	case <-br.starts:
	case <-time.After(pollTimeout):
		t.Fatal("timeout waiting for runner start after pickup")
	}

	deadline := time.Now().Add(pollTimeout)
	var ids []string
	for time.Now().Before(deadline) {
		ids = taskUpdated.Snapshot()
		for _, id := range ids {
			if id == tsk.ID {
				cancel()
				close(br.release)
				<-done
				return
			}
		}
		time.Sleep(pollInterval)
	}
	cancel()
	close(br.release)
	<-done
	t.Fatalf("PublishTaskUpdated never recorded for pickup; got %v want %q", ids, tsk.ID)
}

func TestWorker_failStuckRunning_publishesTaskUpdated(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wb := h.gitBinding()
	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "stuck-running-publishes",
		InitialPrompt: "do the thing",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		WorktreeID:    wb,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := h.store.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("set running: %v", err)
	}

	taskUpdated := notifierfake.NewRecordingTaskUpdatedNotifier()
	_, done := h.startWorker(ctx, runnerfake.New(), worker.Options{
		TaskUpdatedNotifier: taskUpdated,
	})
	h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	cancel()
	<-done

	found := false
	for _, id := range taskUpdated.Snapshot() {
		if id == tsk.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PublishTaskUpdated never recorded for failStuckRunning; got %v", taskUpdated.Snapshot())
	}
}
