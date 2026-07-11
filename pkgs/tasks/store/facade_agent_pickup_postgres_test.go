package store_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// TestStore_AgentPickup_freshReadyTask_postgres mirrors template-instantiate rows:
// ready status, no pending_retry, no gate. SQLite accepts full-model Save on UPDATE;
// Postgres rejects invalid empty json on nullable jsonb (SQLSTATE 22P02).
func TestStore_AgentPickup_freshReadyTask_postgres(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenPostgres(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title:         "template-shaped",
		InitialPrompt: "split the longest function",
		Priority:      taskcoredomain.PriorityMedium,
		Status:        taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = s.Delete(ctx, tsk.ID, taskcoredomain.ActorUser)
	})

	pickup, err := s.AgentPickup(ctx, tsk.ID, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("AgentPickup: %v", err)
	}
	if pickup.Task.Status != taskcoredomain.StatusRunning {
		t.Fatalf("status %q want running", pickup.Task.Status)
	}
	if pickup.Task.PendingRetry != nil {
		t.Fatal("pending_retry should be nil after pickup")
	}
}
