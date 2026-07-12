package harness_test

import (
	"context"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/harnesstest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func seedInterruptedExecute(t *testing.T, st *composition.API, ctx context.Context) (*taskcoredomain.Task, *cyclesdomain.TaskCycle, string) {
	t.Helper()
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "resume", InitialPrompt: "do the thing", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	item, err := st.AddChecklistItem(ctx, tsk.ID, "criterion one", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("update running: %v", err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}
	exec, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start execute: %v", err)
	}
	summary := cyclesdomain.PhaseInterruptReason
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: exec.PhaseSeq,
		Status: cyclesdomain.PhaseStatusFailed, Summary: &summary, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete execute interrupt: %v", err)
	}
	if err := st.UpsertVerifyReports(ctx, cycle.ID, 1, []cyclesstore.VerifyReportEntry{
		{CriterionID: item.ID, Verified: true, VerifierKind: checklistdomain.VerifierAgentSelf, Reasoning: "locked"},
	}); err != nil {
		t.Fatalf("upsert verify: %v", err)
	}
	tsk, _ = st.Get(ctx, tsk.ID)
	return tsk, cycle, item.ID
}

func TestHarness_Resume_afterInterruptedExecute_composesResumePrompt(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	st := storefake.New(t).API
	tsk, cycle, criterionID := seedInterruptedExecute(t, st, ctx)

	promptCh := make(chan string, 1)
	phaseCh := make(chan cyclesdomain.Phase, 1)
	inner := runnerfake.New()
	r := &hookRunner{
		Runner: inner,
		preRun: func(req runner.Request) {
			select {
			case promptCh <- req.Prompt:
			default:
			}
			select {
			case phaseCh <- req.Phase:
			default:
			}
			cancel()
		},
	}

	h := harness.New(st, r, harness.Options{ReportDir: t.TempDir()})
	done := make(chan struct{})
	go func() {
		h.Resume(ctx, tsk, cycle)
		close(done)
	}()

	var prompt string
	select {
	case prompt = <-promptCh:
	case <-time.After(harnesstest.DefaultPollTimeout):
		t.Fatal("timeout waiting for resume execute prompt")
	}
	select {
	case phase := <-phaseCh:
		if phase != cyclesdomain.PhaseExecute {
			t.Fatalf("first runner phase = %q, want execute", phase)
		}
	case <-time.After(harnesstest.DefaultPollTimeout):
		t.Fatal("timeout waiting for runner phase")
	}

	for _, frag := range []string{"Worker resume notice", cycle.ID, "Already verified", criterionID, "do the thing"} {
		if !strings.Contains(prompt, frag) {
			t.Fatalf("resume prompt missing %q\nprompt=%q", frag, prompt)
		}
	}

	phases, err := st.ListPhasesForCycle(context.Background(), cycle.ID)
	if err != nil {
		t.Fatalf("list phases: %v", err)
	}
	if len(phases) < 2 {
		t.Fatalf("expected new execute phase after resume start, got %d phases", len(phases))
	}
	if phases[len(phases)-1].Phase != cyclesdomain.PhaseExecute {
		t.Fatalf("last phase = %+v, want execute", phases[len(phases)-1])
	}

	select {
	case <-done:
	case <-time.After(harnesstest.DefaultPollTimeout):
		t.Fatal("timeout waiting for Resume to finish after cancel")
	}
}
