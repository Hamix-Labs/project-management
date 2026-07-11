package worker_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
	"gorm.io/gorm"
)

type sweepHarness struct {
	t  *testing.T
	st *store.Store
	db *gorm.DB
}

func newSweepHarness(t *testing.T) *sweepHarness {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	return &sweepHarness{t: t, st: store.NewStore(db), db: db}
}

func (h *sweepHarness) makeRunningTaskWithRunningCycleAndPhase(t *testing.T, ctx context.Context, title string, phase cyclesdomain.Phase) (*taskcoredomain.Task, *cyclesdomain.TaskCycle, *cyclesdomain.TaskCyclePhase) {
	t.Helper()
	tsk, err := h.st.Create(ctx, store.CreateTaskInput{
		Title:         title,
		InitialPrompt: "do work",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := h.st.Update(ctx, tsk.ID, store.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("transition task to running: %v", err)
	}
	cycle, err := h.st.StartCycle(ctx, store.StartCycleInput{
		TaskID:      tsk.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}
	if phase == "" {
		return tsk, cycle, nil
	}
	ph, err := h.st.StartPhase(ctx, cycle.ID, phase, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start phase: %v", err)
	}
	return tsk, cycle, ph
}

func TestFinalize_CleanDB_isNoOp(t *testing.T) {
	t.Parallel()
	h := newSweepHarness(t)
	ctx := context.Background()

	res, err := worker.FinalizeInterruptedPhases(ctx, h.st)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if res.PhasesFinalized != 0 {
		t.Fatalf("clean DB finalize result = %+v, want zero phases", res)
	}
}

func TestFinalize_InterruptedRunningPhase_keepsCycleAndTaskRunning(t *testing.T) {
	t.Parallel()
	h := newSweepHarness(t)
	ctx := context.Background()

	tsk, cycle, ph := h.makeRunningTaskWithRunningCycleAndPhase(t, ctx, "interrupt", cyclesdomain.PhaseExecute)

	res, err := worker.FinalizeInterruptedPhases(ctx, h.st)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if res.PhasesFinalized != 1 {
		t.Fatalf("PhasesFinalized = %d, want 1", res.PhasesFinalized)
	}

	gotCycle, err := h.st.GetCycle(ctx, cycle.ID)
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	if gotCycle.Status != cyclesdomain.CycleStatusRunning {
		t.Fatalf("cycle status = %q, want running", gotCycle.Status)
	}

	phases, err := h.st.ListPhasesForCycle(ctx, cycle.ID)
	if err != nil {
		t.Fatalf("list phases: %v", err)
	}
	if len(phases) != 1 || phases[0].ID != ph.ID || phases[0].Status != cyclesdomain.PhaseStatusFailed {
		t.Fatalf("phases = %+v, want one failed phase id=%s", phases, ph.ID)
	}
	if phases[0].Summary == nil || *phases[0].Summary != cyclesdomain.PhaseInterruptReason {
		t.Fatalf("phase summary = %v, want %q", phases[0].Summary, cyclesdomain.PhaseInterruptReason)
	}

	gotTask, err := h.st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if gotTask.Status != taskcoredomain.StatusRunning {
		t.Fatalf("task status = %q, want running", gotTask.Status)
	}
}

func TestSweep_OrphanPhaseUnderTerminalCycle_isFailed(t *testing.T) {
	t.Parallel()
	h := newSweepHarness(t)
	ctx := context.Background()

	tsk, cycle, ph := h.makeRunningTaskWithRunningCycleAndPhase(t, ctx, "orphan-phase", cyclesdomain.PhaseExecute)

	if _, err := h.st.CompletePhase(ctx, store.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: ph.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusSucceeded,
		By:       taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete execute: %v", err)
	}
	if _, err := h.st.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusSucceeded, "", taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("terminate cycle: %v", err)
	}

	tx := h.db.Exec(
		"UPDATE task_cycle_phases SET status = ?, ended_at = NULL WHERE phase_seq = ? AND cycle_id = ?",
		cyclesdomain.PhaseStatusRunning, ph.PhaseSeq, cycle.ID,
	)
	if tx.Error != nil {
		t.Fatalf("synthesize orphan running phase: %v", tx.Error)
	}

	fr, err := worker.FinalizeInterruptedPhases(ctx, h.st)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if fr.PhasesFinalized != 1 {
		t.Fatalf("PhasesFinalized = %d, want 1", fr.PhasesFinalized)
	}

	phases, err := h.st.ListPhasesForCycle(ctx, cycle.ID)
	if err != nil {
		t.Fatalf("list phases: %v", err)
	}
	for _, p := range phases {
		if p.Status == cyclesdomain.PhaseStatusRunning {
			t.Fatalf("phase %d still running", p.PhaseSeq)
		}
	}

	if _, err := h.st.Get(ctx, tsk.ID); err != nil {
		t.Fatalf("get task: %v", err)
	}
}

func TestFinalize_Idempotent(t *testing.T) {
	t.Parallel()
	h := newSweepHarness(t)
	ctx := context.Background()

	_, _, _ = h.makeRunningTaskWithRunningCycleAndPhase(t, ctx, "idemp", cyclesdomain.PhaseExecute)

	first, err := worker.FinalizeInterruptedPhases(ctx, h.st)
	if err != nil {
		t.Fatalf("first finalize: %v", err)
	}
	if first.PhasesFinalized != 1 {
		t.Fatalf("first finalize PhasesFinalized = %d, want 1", first.PhasesFinalized)
	}
	second, err := worker.FinalizeInterruptedPhases(ctx, h.st)
	if err != nil {
		t.Fatalf("second finalize: %v", err)
	}
	if second.PhasesFinalized != 0 {
		t.Fatalf("second finalize should be no-op, got %+v", second)
	}
}
