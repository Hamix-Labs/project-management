package store_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestAgentPickup_happyPath(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "pickup-ready",
		Status:   domain.StatusReady,
		Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	got, err := st.AgentPickup(ctx, task.ID, domain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	if got.Task == nil || got.Task.Status != domain.StatusRunning {
		t.Fatalf("task status = %v, want running", got.Task)
	}
	if got.ConsumedRetry != nil {
		t.Fatalf("ConsumedRetry = %+v, want nil", got.ConsumedRetry)
	}

	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.StatusRunning {
		t.Fatalf("stored status = %q, want running", stored.Status)
	}
	if stored.PendingRetry != nil {
		t.Fatalf("stored PendingRetry = %+v, want nil", stored.PendingRetry)
	}
}

func TestAgentPickup_rejectsNonReady(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "pickup-blocked",
		Status:   domain.StatusBlocked,
		Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	_, err = st.AgentPickup(ctx, task.ID, domain.ActorAgent)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestAgentPickup_rejectsEmptyIDAndBadActor(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	_, err := st.AgentPickup(ctx, "  ", domain.ActorAgent)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty id: err = %v, want ErrInvalidInput", err)
	}
	_, err = st.AgentPickup(ctx, "some-id", domain.Actor("nope"))
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("bad actor: err = %v, want ErrInvalidInput", err)
	}
}

func TestAgentPickup_concurrentOneWins(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "pickup-race",
		Status:   domain.StatusReady,
		Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	const n = 8
	var wg sync.WaitGroup
	errs := make(chan error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, err := st.AgentPickup(ctx, task.ID, domain.ActorAgent)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var wins, rejects, other int
	for err := range errs {
		switch {
		case err == nil:
			wins++
		case errors.Is(err, domain.ErrInvalidInput):
			rejects++
		default:
			other++
			t.Errorf("unexpected err: %v", err)
		}
	}
	if wins != 1 {
		t.Fatalf("wins = %d, want 1", wins)
	}
	if rejects != n-1 {
		t.Fatalf("rejects = %d, want %d", rejects, n-1)
	}
	if other != 0 {
		t.Fatalf("other errors = %d", other)
	}

	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.StatusRunning {
		t.Fatalf("stored status = %q, want running", stored.Status)
	}
}

func TestAgentPickup_consumesPendingRetry(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, cycleID := mustFailedTaskWithTerminalCycle(t, st, cycles)

	updated, _, err := st.RequestTaskRetry(ctx, taskcorestore.RequestRetryInput{
		TaskID:        task.ID,
		Mode:          domain.RetryFresh,
		ParentCycleID: cycleID,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.StatusReady || updated.PendingRetry == nil {
		t.Fatalf("retry setup: status=%q pending=%+v", updated.Status, updated.PendingRetry)
	}

	got, err := st.AgentPickup(ctx, task.ID, domain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	if got.ConsumedRetry == nil {
		t.Fatal("ConsumedRetry is nil")
	}
	if got.ConsumedRetry.Mode != domain.RetryFresh || got.ConsumedRetry.ParentCycleID != cycleID {
		t.Fatalf("ConsumedRetry = %+v", got.ConsumedRetry)
	}
	if got.Task.PendingRetry != nil {
		t.Fatalf("task still has PendingRetry: %+v", got.Task.PendingRetry)
	}

	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry != nil {
		t.Fatalf("stored PendingRetry = %+v, want nil", stored.PendingRetry)
	}
}

func mustFailedTaskWithTerminalCycle(
	t *testing.T,
	st *taskcorestore.Store,
	cycles *cyclesstore.Store,
) (*domain.Task, string) {
	t.Helper()
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "failed-for-retry",
		Status:   domain.StatusReady,
		Priority: domain.PriorityMedium,
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
	if _, err := cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "test fail", domain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	failed := domain.StatusFailed
	updated, _, err := st.Update(ctx, task.ID, taskcorestore.UpdateTaskInput{Status: &failed}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.StatusFailed {
		t.Fatalf("task status = %q, want failed", updated.Status)
	}
	return updated, cycle.ID
}
