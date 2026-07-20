package worker_test

import (
	"context"
	"os"
	"testing"
	"time"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestWorker_missingGitBinding_defersPickup(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "unbound",
		InitialPrompt: "do the thing",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, done := h.startWorker(ctx, runnerfake.New(), worker.Options{})
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	got, err := h.store.Get(context.Background(), tsk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != taskcoredomain.StatusReady {
		t.Fatalf("status=%q want ready (missing binding should defer, not run)", got.Status)
	}
	if got.PickupNotBefore == nil {
		t.Fatal("expected pickup_not_before defer")
	}
}

// TestWorker_gitPrepAbort_terminatesOpenCycleAndFailsTask pins B-07:
// when git prep fails for a Running task that already has an open cycle
// (and phase), abort must CompletePhase + TerminateCycle + Fail — not
// leave an orphan Running cycle under a Failed task.
func TestWorker_gitPrepAbort_terminatesOpenCycleAndFailsTask(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wb := h.gitBinding()
	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "prep-abort",
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

	if err := os.RemoveAll(h.workDir); err != nil {
		t.Fatalf("remove worktree dir: %v", err)
	}

	// Create already enqueued the Ready row; processOne reloads Running from store.
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
	if phases[0].Summary == nil || *phases[0].Summary != "git_prep_failed" {
		t.Fatalf("phase summary = %v, want git_prep_failed", phases[0].Summary)
	}
}
