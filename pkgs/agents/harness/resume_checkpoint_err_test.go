package harness_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// Late Resume after successful finalize must not clobber review→failed.
func TestResume_checkpointErrAfterSucceededCycle_keepsReview(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := storefake.New(t).API

	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "late-resume", InitialPrompt: "work", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("running: %v", err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}
	exec, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start execute: %v", err)
	}
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: exec.PhaseSeq,
		Status: cyclesdomain.PhaseStatusSucceeded, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete execute: %v", err)
	}
	ver, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseVerify, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start verify: %v", err)
	}
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: ver.PhaseSeq,
		Status: cyclesdomain.PhaseStatusSucceeded, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete verify: %v", err)
	}
	if _, err := st.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusSucceeded, "", taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	review := taskcoredomain.StatusReview
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &review}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("review: %v", err)
	}

	// Stale in-memory cycle still looks "running" (duplicate Resume race).
	stale := *cycle
	stale.Status = cyclesdomain.CycleStatusRunning
	tsk, _ = st.Get(ctx, tsk.ID)

	h := harness.New(st, runnerfake.New(), harness.Options{ReportDir: t.TempDir()})
	h.Resume(ctx, tsk, &stale)

	got, err := st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusReview {
		t.Fatalf("task status = %q, want review (late Resume must not clobber)", got.Status)
	}
}

// Mid-finalize: verify succeeded, cycle still running — late Resume must not fail the task.
func TestResume_checkpointErrVerifySucceededStillRunning_doesNotFail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := storefake.New(t).API

	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "mid-finalize", InitialPrompt: "work", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("running: %v", err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}
	exec, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start execute: %v", err)
	}
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: exec.PhaseSeq,
		Status: cyclesdomain.PhaseStatusSucceeded, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete execute: %v", err)
	}
	ver, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseVerify, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start verify: %v", err)
	}
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: ver.PhaseSeq,
		Status: cyclesdomain.PhaseStatusSucceeded, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete verify: %v", err)
	}

	tsk, _ = st.Get(ctx, tsk.ID)
	h := harness.New(st, runnerfake.New(), harness.Options{ReportDir: t.TempDir()})
	h.Resume(ctx, tsk, cycle)

	got, err := st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusRunning {
		t.Fatalf("task status = %q, want running (owner still finalizing)", got.Status)
	}
}
