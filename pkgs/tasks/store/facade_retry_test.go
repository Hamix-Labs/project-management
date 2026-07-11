package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_PendingRetry_setAndAgentPickupConsumes(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenSQLite(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title: "retry-intent", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusFailed,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	intent := &taskcoredomain.PendingRetry{Mode: taskcoredomain.RetryFresh, ParentCycleID: "parent-cycle-id"}
	ready := taskcoredomain.StatusReady
	got, err := s.Update(ctx, tsk.ID, store.UpdateTaskInput{
		Status:       &ready,
		PendingRetry: intent,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if got.PendingRetry == nil || got.PendingRetry.Mode != taskcoredomain.RetryFresh {
		t.Fatalf("pending_retry: %+v", got.PendingRetry)
	}

	pickup, err := s.AgentPickup(ctx, tsk.ID, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	if pickup.Task.Status != taskcoredomain.StatusRunning {
		t.Fatalf("status %q want running", pickup.Task.Status)
	}
	if pickup.Task.PendingRetry != nil {
		t.Fatal("pending_retry should be cleared on pickup")
	}
	if pickup.ConsumedRetry == nil || pickup.ConsumedRetry.ParentCycleID != intent.ParentCycleID {
		t.Fatalf("consumed: %+v", pickup.ConsumedRetry)
	}

	after, err := s.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.PendingRetry != nil {
		t.Fatal("pending_retry persisted after pickup")
	}
}

func TestStore_AgentPickup_rejectsNonReady(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenSQLite(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title: "not-ready", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusFailed,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.AgentPickup(ctx, tsk.ID, taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v want ErrInvalidInput", err)
	}
}

func TestStore_Update_pendingRetrySetAndClearConflict(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenSQLite(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title: "conflict", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	intent := &taskcoredomain.PendingRetry{Mode: taskcoredomain.RetryResume, ParentCycleID: "c1"}
	_, err = s.Update(ctx, tsk.ID, store.UpdateTaskInput{
		PendingRetry:      intent,
		ClearPendingRetry: true,
	}, taskcoredomain.ActorUser)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v want ErrInvalidInput", err)
	}
}

func TestStore_RequestTaskRetry_freshAndResume(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenSQLite(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title: "retry", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusFailed,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := s.StartCycle(ctx, store.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "x", taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	got, err := s.RequestTaskRetry(ctx, store.RequestRetryInput{
		TaskID: tsk.ID, Mode: taskcoredomain.RetryFresh,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusReady || got.PendingRetry == nil || got.PendingRetry.ParentCycleID != cycle.ID {
		t.Fatalf("task=%+v", got)
	}

	got2, err := s.RequestTaskRetry(ctx, store.RequestRetryInput{
		TaskID: tsk.ID, Mode: taskcoredomain.RetryFresh,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if got2.PendingRetry == nil || !got2.PendingRetry.Equal(got.PendingRetry) {
		t.Fatalf("idempotent pending_retry=%+v", got2.PendingRetry)
	}

	failed := taskcoredomain.StatusFailed
	if _, err := s.Update(ctx, tsk.ID, store.UpdateTaskInput{Status: &failed}, taskcoredomain.ActorUser); err != nil {
		t.Fatal(err)
	}
	got3, err := s.RequestTaskRetry(ctx, store.RequestRetryInput{
		TaskID: tsk.ID, Mode: taskcoredomain.RetryResume,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if got3.PendingRetry == nil || got3.PendingRetry.Mode != taskcoredomain.RetryResume {
		t.Fatalf("resume pending_retry=%+v", got3.PendingRetry)
	}
}

func TestStore_RequestTaskRetry_rejectsNonFailed(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenSQLite(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title: "ready", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.RequestTaskRetry(ctx, store.RequestRetryInput{
		TaskID: tsk.ID, Mode: taskcoredomain.RetryFresh,
	}, taskcoredomain.ActorUser)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v want ErrInvalidInput", err)
	}
}

func TestStore_RequestTaskRetry_noTerminalCycle(t *testing.T) {
	ctx := context.Background()
	db := tasktestdb.OpenSQLite(t)
	s := store.NewStore(db)

	tsk, err := s.Create(ctx, store.CreateTaskInput{
		Title: "no-cycle", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusFailed,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.RequestTaskRetry(ctx, store.RequestRetryInput{
		TaskID: tsk.ID, Mode: taskcoredomain.RetryFresh,
	}, taskcoredomain.ActorUser)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v want ErrInvalidInput", err)
	}
}
