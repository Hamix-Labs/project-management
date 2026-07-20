package worker_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// TestWorker_runningWithoutOpenCycle_failsTask pins B-08: a Running dequeue
// with no open cycle must fail the task instead of Ack-and-drop.
func TestWorker_runningWithoutOpenCycle_failsTask(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wb := h.gitBinding()
	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "stuck-running-no-cycle",
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

	_, done := h.startWorker(ctx, runnerfake.New(), worker.Options{})
	got := h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	cancel()
	<-done

	if got.Status != taskcoredomain.StatusFailed {
		t.Fatalf("status=%q want failed", got.Status)
	}
	cycles, err := h.store.ListCyclesForTask(context.Background(), tsk.ID, 0)
	if err != nil {
		t.Fatalf("list cycles: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("cycles=%d want 0", len(cycles))
	}
}

// TestWorker_runningMissingBinding_failsTaskAndTerminatesCycle pins B-08:
// Running + open cycle + missing git binding must terminate the cycle and
// fail the task, not Ack-and-drop.
func TestWorker_runningMissingBinding_failsTaskAndTerminatesCycle(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "stuck-running-no-binding",
		InitialPrompt: "do the thing",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := h.store.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("set running: %v", err)
	}
	cycle, err := h.store.StartCycle(ctx, cyclescontract.StartCycleInput{
		TaskID:      tsk.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}
	ph, err := h.store.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start phase: %v", err)
	}

	_, done := h.startWorker(ctx, runnerfake.New(), worker.Options{})
	h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	cancel()
	<-done

	gotCycle, err := h.store.GetCycle(context.Background(), cycle.ID)
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	if gotCycle.Status != cyclesdomain.CycleStatusFailed {
		t.Fatalf("cycle status = %q, want %q", gotCycle.Status, cyclesdomain.CycleStatusFailed)
	}

	phases, err := h.store.ListPhasesForCycle(context.Background(), cycle.ID)
	if err != nil {
		t.Fatalf("list phases: %v", err)
	}
	if len(phases) != 1 || phases[0].ID != ph.ID {
		t.Fatalf("phases = %+v, want one phase id=%s", phases, ph.ID)
	}
	if phases[0].Status != cyclesdomain.PhaseStatusFailed {
		t.Fatalf("phase status = %q, want failed", phases[0].Status)
	}
	if phases[0].Summary == nil || *phases[0].Summary != "running_missing_git_binding" {
		t.Fatalf("phase summary = %v, want running_missing_git_binding", phases[0].Summary)
	}
}
