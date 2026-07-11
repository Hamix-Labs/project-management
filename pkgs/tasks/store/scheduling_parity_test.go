package store

import (
	"context"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// I3 contract gate: scheduling.EvaluateWorkerReadiness must match ListQueueCandidates.
func TestSchedulingParity_GoReadinessMatchesListQueueCandidates(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := context.Background()
	s := NewStore(db)
	now := time.Now().UTC()

	dep, err := s.Create(ctx, CreateTaskInput{
		Title: "dep", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	blocked, err := s.Create(ctx, CreateTaskInput{
		Title: "blocked", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddTaskDependency(ctx, blocked.ID, dep.ID, taskcoredomain.DependencySatisfiesDone); err != nil {
		t.Fatal(err)
	}

	heldGate := &taskcoredomain.TaskGate{Kind: taskcoredomain.GateKindManualApproval, Status: taskcoredomain.GateStatusActive}
	held, err := s.Create(ctx, CreateTaskInput{
		Title: "held", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium, Gate: heldGate,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	future := now.Add(2 * time.Hour)
	if _, err := s.Create(ctx, CreateTaskInput{
		Title: "deferred", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
		PickupNotBefore: &future,
	}, taskcoredomain.ActorUser); err != nil {
		t.Fatal(err)
	}

	eligible, err := s.Create(ctx, CreateTaskInput{
		Title: "eligible", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	candidateIDs := map[string]struct{}{}
	cands, err := s.ListReadyTaskQueueCandidates(ctx, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cands {
		candidateIDs[c.Task.ID] = struct{}{}
	}

	allTasks, err := s.ListFlat(ctx, 100, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := range allTasks {
		if allTasks[i].Status != taskcoredomain.StatusReady {
			continue
		}
		task := &allTasks[i]
		goReady, pred, err := s.ReadyForAgentPickup(ctx, task, now)
		if err != nil {
			t.Fatal(err)
		}
		_, inSQL := candidateIDs[task.ID]
		if goReady != inSQL {
			t.Fatalf("task %q: Go ready=%v (pred=%s) SQL candidate=%v", task.ID, goReady, pred, inSQL)
		}
	}

	if _, ok := candidateIDs[eligible.ID]; !ok {
		t.Fatalf("eligible task %q should be SQL candidate", eligible.ID)
	}
	if _, ok := candidateIDs[blocked.ID]; ok {
		t.Fatalf("blocked task %q should not be SQL candidate", blocked.ID)
	}
	if _, ok := candidateIDs[held.ID]; ok {
		t.Fatalf("held task %q should not be SQL candidate", held.ID)
	}
}
