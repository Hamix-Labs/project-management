package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestRequestTaskApprove_fromReviewSetsDone(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)

	updated, prev, err := st.RequestTaskApprove(ctx, task.ID, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusReview {
		t.Fatalf("prev = %q, want review", prev)
	}
	if updated.Status != domain.StatusDone {
		t.Fatalf("status = %q, want done", updated.Status)
	}

	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.StatusDone {
		t.Fatalf("stored status = %q, want done", stored.Status)
	}
}

func TestRequestTaskApprove_rejectsNonReview(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "ready", Status: domain.StatusReady, Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = st.RequestTaskApprove(ctx, task.ID, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskApprove_rejectsAgent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	_, _, err := st.RequestTaskApprove(ctx, task.ID, domain.ActorAgent)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestUpdate_rejectsStatusDone(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", Status: domain.StatusReview, Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	done := domain.StatusDone
	_, _, err = st.Update(ctx, task.ID, taskcorestore.UpdateTaskInput{Status: &done}, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func mustReviewTaskWithSucceededCycle(t *testing.T, st *taskcorestore.Store, cycles *cyclesstore.Store) (*domain.Task, string) {
	t.Helper()
	ctx := context.Background()
	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "review", Status: domain.StatusReview, Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: domain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusSucceeded, "ok", domain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	return task, cycle.ID
}
