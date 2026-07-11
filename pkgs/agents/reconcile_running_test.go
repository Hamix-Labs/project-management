package agents_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestReconcileRunningTasksNotQueued_enqueuesOpenCycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := store.NewStore(tasktestdb.OpenSQLite(t))
	q := agents.NewMemoryQueue(8)

	tsk, err := st.Create(ctx, store.CreateTaskInput{
		Title: "resume-me", InitialPrompt: "work", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, store.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("update running: %v", err)
	}
	if _, err := st.StartCycle(ctx, store.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent}); err != nil {
		t.Fatalf("start cycle: %v", err)
	}

	res, err := agents.ReconcileRunningTasksNotQueued(ctx, st, q)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if res.Scanned != 1 || res.Enqueued != 1 {
		t.Fatalf("reconcile result = %+v, want scanned=1 enqueued=1", res)
	}

	res2, err := agents.ReconcileRunningTasksNotQueued(ctx, st, q)
	if err != nil {
		t.Fatalf("second reconcile: %v", err)
	}
	if res2.Enqueued != 0 || res2.SkippedAlreadyQueued != 1 {
		t.Fatalf("second reconcile = %+v, want skipped already queued", res2)
	}
}
