package verify_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// EC-09 (docs/domain/harness.md): locked passes survive infra verify retries.
// Integration: TestEdgeCase_EC09_partialPass_infraVerifyOnly in cycle_verify_only_test.go.
// TestWorker_VerifyPhase_carriesPassesAcrossRetries pins one-shot terminate:
// when attempt 1 passes c1 and fails c2, the cycle ends failed with no
// in-cycle retry and no completion rows committed.
func TestWorker_VerifyPhase_carriesPassesAcrossRetries(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "verify-carry")
	c1, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add c1: %v", err)
	}
	c2, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion two", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add c2: %v", err)
	}

	maxRetries := 2
	if _, err := h.Store.UpdateSettings(ctx, settingscontract.SettingsPatch{VerifyMaxRetries: &maxRetries}); err != nil {
		t.Fatalf("set max retries: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		switch req.Phase {
		case cyclesdomain.PhaseExecute:
			writeCriteriaReportFor(t, reportDir, cycles[0].ID, []string{c1.ID, c2.ID})
		case cyclesdomain.PhaseVerify:
			writePartialVerifyReport(t, reportDir, cycles[0].ID, map[string]bool{
				c1.ID: true, c2.ID: false,
			})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", nil, ""))
	r.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "verify ok", nil, ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	<-done
	cancel()

	bg := context.Background()
	items, err := h.Store.ListChecklistForSubject(bg, tsk.ID)
	if err != nil {
		t.Fatalf("list checklist: %v", err)
	}
	for _, it := range items {
		if it.Done {
			t.Errorf("expected NO completed items on one-shot failure; %s is done", it.ID)
		}
	}

	executeCalls, verifyCalls := 0, 0
	for _, c := range r.Calls() {
		switch c.Phase {
		case cyclesdomain.PhaseExecute:
			executeCalls++
		case cyclesdomain.PhaseVerify:
			verifyCalls++
		}
	}
	if executeCalls != 1 {
		t.Fatalf("execute calls = %d, want 1 (one-shot)", executeCalls)
	}
	if verifyCalls != 1 {
		t.Fatalf("verify calls = %d, want 1 (one-shot)", verifyCalls)
	}

	cycles, err := h.Store.ListCyclesForTask(bg, tsk.ID, 5)
	if err != nil {
		t.Fatalf("list cycles: %v", err)
	}
	if len(cycles) == 0 {
		t.Fatalf("no cycles recorded")
	}
	if cycles[0].Status != cyclesdomain.CycleStatusFailed {
		t.Fatalf("cycle status = %q, want failed", cycles[0].Status)
	}
}
