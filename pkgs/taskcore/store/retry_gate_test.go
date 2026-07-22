package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestRequestTaskRetry_fromFailedSetsReadyAndPendingRetry(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, cycleID := mustFailedTaskWithTerminalCycle(t, st, cycles)

	updated, prev, err := st.RequestTaskRetry(ctx, taskcorestore.RequestRetryInput{
		TaskID:        task.ID,
		Mode:          domain.RetryResume,
		ParentCycleID: cycleID,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusFailed {
		t.Fatalf("prev = %q, want failed", prev)
	}
	if updated.Status != domain.StatusReady {
		t.Fatalf("status = %q, want ready", updated.Status)
	}
	if updated.PendingRetry == nil || updated.PendingRetry.Mode != domain.RetryResume || updated.PendingRetry.ParentCycleID != cycleID {
		t.Fatalf("PendingRetry = %+v", updated.PendingRetry)
	}

	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry == nil || !stored.PendingRetry.Equal(updated.PendingRetry) {
		t.Fatalf("stored PendingRetry = %+v, want %+v", stored.PendingRetry, updated.PendingRetry)
	}
}

func TestRequestTaskRetry_idempotentSameIntent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, cycleID := mustFailedTaskWithTerminalCycle(t, st, cycles)
	in := taskcorestore.RequestRetryInput{
		TaskID:        task.ID,
		Mode:          domain.RetryFresh,
		ParentCycleID: cycleID,
	}
	first, _, err := st.RequestTaskRetry(ctx, in, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	second, prev, err := st.RequestTaskRetry(ctx, in, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusReady {
		t.Fatalf("prev on idempotent = %q, want ready", prev)
	}
	if second.PendingRetry == nil || !second.PendingRetry.Equal(first.PendingRetry) {
		t.Fatalf("second PendingRetry = %+v, want %+v", second.PendingRetry, first.PendingRetry)
	}
}

func TestRequestTaskRetry_conflictDifferentIntent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, cycleID := mustFailedTaskWithTerminalCycle(t, st, cycles)
	if _, _, err := st.RequestTaskRetry(ctx, taskcorestore.RequestRetryInput{
		TaskID:        task.ID,
		Mode:          domain.RetryFresh,
		ParentCycleID: cycleID,
	}, domain.ActorUser); err != nil {
		t.Fatal(err)
	}

	_, _, err := st.RequestTaskRetry(ctx, taskcorestore.RequestRetryInput{
		TaskID:        task.ID,
		Mode:          domain.RetryResume,
		ParentCycleID: cycleID,
	}, domain.ActorUser)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestRequestTaskRetry_rejectsNonFailedStatuses(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "running-no-retry",
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
	if _, err := cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "seed", domain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	for _, status := range []domain.Status{domain.StatusReady, domain.StatusRunning, domain.StatusOnHold} {
		s := status
		if _, _, err := st.Update(ctx, task.ID, taskcorestore.UpdateTaskInput{Status: &s}, domain.ActorUser); err != nil {
			t.Fatalf("set %s: %v", status, err)
		}
		_, _, err := st.RequestTaskRetry(ctx, taskcorestore.RequestRetryInput{
			TaskID:        task.ID,
			Mode:          domain.RetryFresh,
			ParentCycleID: cycle.ID,
		}, domain.ActorUser)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("status %s: err = %v, want ErrInvalidInput", status, err)
		}
	}
}

func TestApplyTaskGateAction_releaseHoldClearHold(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	gate := &domain.TaskGate{
		Kind:   domain.GateKindManualApproval,
		Status: domain.GateStatusActive,
	}
	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "gated",
		Status:   domain.StatusReady,
		Priority: domain.PriorityMedium,
		Gate:     gate,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	released, err := st.ApplyTaskGateAction(ctx, task.ID, contract.GateActionRelease, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if released.Gate == nil || released.Gate.Status != domain.GateStatusReleased || released.Gate.Hold {
		t.Fatalf("after release: %+v", released.Gate)
	}

	pending := &domain.TaskGate{
		Kind:   domain.GateKindManualApproval,
		Status: domain.GateStatusPendingRelease,
	}
	gatePtr := pending
	if _, _, err := st.Update(ctx, task.ID, taskcorestore.UpdateTaskInput{Gate: &gatePtr}, domain.ActorUser); err != nil {
		t.Fatal(err)
	}

	held, err := st.ApplyTaskGateAction(ctx, task.ID, contract.GateActionHold, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if held.Gate == nil || !held.Gate.Hold || held.Gate.Status != domain.GateStatusPendingRelease {
		t.Fatalf("after hold: %+v", held.Gate)
	}

	cleared, err := st.ApplyTaskGateAction(ctx, task.ID, contract.GateActionClearHold, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Gate == nil || cleared.Gate.Hold {
		t.Fatalf("after clear_hold: %+v", cleared.Gate)
	}
}

func TestApplyTaskGateAction_rejectsIllegal(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	noGate, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "no-gate",
		Status:   domain.StatusReady,
		Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.ApplyTaskGateAction(ctx, noGate.ID, contract.GateActionRelease, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("no gate: err = %v, want ErrInvalidInput", err)
	}

	active := &domain.TaskGate{
		Kind:   domain.GateKindManualApproval,
		Status: domain.GateStatusActive,
	}
	gated, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "active-gate",
		Status:   domain.StatusReady,
		Priority: domain.PriorityMedium,
		Gate:     active,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.ApplyTaskGateAction(ctx, gated.ID, contract.GateActionHold, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("hold while active: err = %v, want ErrInvalidInput", err)
	}
	_, err = st.ApplyTaskGateAction(ctx, gated.ID, contract.GateAction("not-a-real-action"), domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("bad action: err = %v, want ErrInvalidInput", err)
	}
	_, err = st.ApplyTaskGateAction(ctx, gated.ID, contract.GateAction("  "), domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty action: err = %v, want ErrInvalidInput", err)
	}
}
