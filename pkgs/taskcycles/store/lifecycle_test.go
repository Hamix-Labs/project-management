package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestCycleLifecycle_startPhaseCompleteTerminate(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task := mustCreateReadyTask(t, tasks, "lifecycle")

	cycle, err := cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cycle.Status != cyclesdomain.CycleStatusRunning || cycle.AttemptSeq != 1 {
		t.Fatalf("cycle = %+v", cycle)
	}

	running, err := cycles.ListRunningCycles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !containsCycleID(running, cycle.ID) {
		t.Fatalf("ListRunningCycles missing %s: %+v", cycle.ID, running)
	}

	phase, err := cycles.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	if phase.Status != cyclesdomain.PhaseStatusRunning || phase.PhaseSeq != 1 {
		t.Fatalf("phase = %+v", phase)
	}

	runningPhases, err := cycles.ListRunningCyclePhases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !containsPhaseID(runningPhases, phase.ID) {
		t.Fatalf("ListRunningCyclePhases missing %s", phase.ID)
	}

	summary := "ok"
	completed, err := cycles.CompletePhase(ctx, cyclesstore.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusSucceeded,
		Summary:  &summary,
		By:       taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != cyclesdomain.PhaseStatusSucceeded || completed.EndedAt == nil {
		t.Fatalf("completed = %+v", completed)
	}

	runningPhases, err = cycles.ListRunningCyclePhases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if containsPhaseID(runningPhases, phase.ID) {
		t.Fatalf("completed phase still listed as running")
	}

	terminated, err := cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusSucceeded, "done", taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	if terminated.Status != cyclesdomain.CycleStatusSucceeded || terminated.EndedAt == nil {
		t.Fatalf("terminated = %+v", terminated)
	}

	running, err = cycles.ListRunningCycles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if containsCycleID(running, cycle.ID) {
		t.Fatalf("terminated cycle still in ListRunningCycles")
	}
}

func TestCycleLifecycle_illegalTransitions(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task := mustCreateReadyTask(t, tasks, "illegal")

	cycle, err := cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("second running cycle: err = %v, want ErrInvalidInput", err)
	}

	phase, err := cycles.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "still running phase", taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("terminate with running phase: err = %v, want ErrInvalidInput", err)
	}

	_, err = cycles.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseVerify, taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("second running phase: err = %v, want ErrInvalidInput", err)
	}

	if _, err := cycles.CompletePhase(ctx, cyclesstore.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusSucceeded,
		By:       taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatal(err)
	}

	_, err = cycles.CompletePhase(ctx, cyclesstore.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusFailed,
		By:       taskcoredomain.ActorAgent,
	})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("complete already terminal: err = %v, want ErrInvalidInput", err)
	}

	_, err = cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusRunning, "not terminal", taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("non-terminal terminate status: err = %v, want ErrInvalidInput", err)
	}

	if _, err := cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusAborted, "abort", taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	_, err = cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "again", taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("re-terminate: err = %v, want ErrInvalidInput", err)
	}

	_, err = cycles.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("start phase on terminal cycle: err = %v, want ErrInvalidInput", err)
	}
}

func mustCreateReadyTask(t *testing.T, st *taskcorestore.Store, title string) *taskcoredomain.Task {
	t.Helper()
	task, err := st.Create(context.Background(), taskcorestore.CreateTaskInput{
		Title:    title,
		Status:   taskcoredomain.StatusReady,
		Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

func containsCycleID(cycles []cyclesdomain.TaskCycle, id string) bool {
	for _, c := range cycles {
		if c.ID == id {
			return true
		}
	}
	return false
}

func containsPhaseID(phases []cyclesdomain.TaskCyclePhase, id string) bool {
	for _, p := range phases {
		if p.ID == id {
			return true
		}
	}
	return false
}
