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

func TestRequestTaskApprove_fromPrReadySetsDone(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustPrReadyTaskWithSucceededCycle(t, st, cycles)

	updated, prev, err := st.RequestTaskApprove(ctx, task.ID, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusPrReady {
		t.Fatalf("prev = %q, want pr_ready", prev)
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

func TestRequestTaskApprove_rejectsReview(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	_, _, err := st.RequestTaskApprove(ctx, task.ID, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskApprove_rejectsNonPrReady(t *testing.T) {
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

	task, _ := mustPrReadyTaskWithSucceededCycle(t, st, cycles)
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

func TestUpdate_rejectsStatusPrReady(t *testing.T) {
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
	pr := domain.StatusPrReady
	_, _, err = st.Update(ctx, task.ID, taskcorestore.UpdateTaskInput{Status: &pr}, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskOpenPR_fromReviewQueuesIntent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, cycleID := mustReviewTaskWithSucceededCycle(t, st, cycles)
	updated, prev, err := st.RequestTaskOpenPR(ctx, taskcorestore.RequestOpenPRInput{TaskID: task.ID}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusReview {
		t.Fatalf("prev = %q, want review", prev)
	}
	if updated.Status != domain.StatusReady {
		t.Fatalf("status = %q, want ready", updated.Status)
	}
	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry == nil {
		t.Fatal("expected pending_retry")
	}
	if stored.PendingRetry.NormalizeKind() != domain.PendingKindOpenPR {
		t.Fatalf("kind = %q, want open_pr", stored.PendingRetry.Kind)
	}
	if stored.PendingRetry.ParentCycleID != cycleID {
		t.Fatalf("parent = %q, want %q", stored.PendingRetry.ParentCycleID, cycleID)
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

func mustPrReadyTaskWithSucceededCycle(t *testing.T, st *taskcorestore.Store, cycles *cyclesstore.Store) (*domain.Task, string) {
	t.Helper()
	ctx := context.Background()
	task, cycleID := mustReviewTaskWithSucceededCycle(t, st, cycles)
	// Create path allows seeding statuses; pr_ready is not client-writable on PATCH.
	pr := domain.StatusPrReady
	// Direct save via Update is rejected — seed by creating with pr_ready if allowed,
	// else open-pr then manually set. Create with StatusPrReady is blocked by ValidClientWritableStatus.
	// Use SQLite raw update for test fixture after cycle exists.
	if err := st.DB().WithContext(ctx).Exec(`UPDATE tasks SET status = ? WHERE id = ?`, string(domain.StatusPrReady), task.ID).Error; err != nil {
		t.Fatal(err)
	}
	_ = pr
	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.StatusPrReady {
		t.Fatalf("fixture status = %q, want pr_ready", stored.Status)
	}
	return stored, cycleID
}
