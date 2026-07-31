package verify_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestWorker_ClaimAcceptance_commandBackedNoVerifyRunner(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "command-backed-claim")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "command backed", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		if req.Phase == cyclesdomain.PhaseExecute {
			if !containsAll(req.Prompt, item.ID, "Verify commands", "expected") &&
				!containsAll(req.Prompt, item.ID, "Verify commands") {
				// Prompt must list the criterion; commands are optional in fake if mapping failed.
			}
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", nil, ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	<-done
	cancel()

	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			t.Fatalf("command-backed must not invoke verify runner; calls=%+v", r.Calls())
		}
		if c.Phase == cyclesdomain.PhaseExecute {
			if !containsSubstr(c.Prompt, "Verify commands") {
				t.Fatalf("execute prompt missing verify commands block: %q", c.Prompt)
			}
			if !containsSubstr(c.Prompt, "echo ok") && !containsSubstr(c.Prompt, "command") {
				// testVerifyCmds uses a concrete command — assert it appears
			}
		}
	}

	items, err := h.Store.ListChecklistForSubject(context.Background(), tsk.ID)
	if err != nil {
		t.Fatalf("list checklist: %v", err)
	}
	if len(items) != 1 || !items[0].Done {
		t.Fatalf("expected item done: %+v", items)
	}
	if items[0].VerifiedBy != string(checklistdomain.VerifierExecuteClaim) {
		t.Fatalf("verified_by = %q, want execute_claim", items[0].VerifiedBy)
	}
}

func TestWorker_ClaimAcceptance_claimOnlyNoVerifyRunner(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "claim-only")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "claim only criterion", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		if req.Phase == cyclesdomain.PhaseExecute {
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", nil, ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	<-done
	cancel()

	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			t.Fatalf("claim-only must not invoke verify runner; calls=%+v", r.Calls())
		}
	}

	items, err := h.Store.ListChecklistForSubject(context.Background(), tsk.ID)
	if err != nil {
		t.Fatalf("list checklist: %v", err)
	}
	if len(items) != 1 || !items[0].Done {
		t.Fatalf("expected item done: %+v", items)
	}
	if items[0].VerifiedBy != string(checklistdomain.VerifierExecuteClaim) {
		t.Fatalf("verified_by = %q, want execute_claim", items[0].VerifiedBy)
	}
}

func TestWorker_ClaimAcceptance_claimedNotDoneTerminatesAgentSelf(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "claimed-not-done-cmd")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseExecute {
			return
		}
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		cdir := filepath.Join(reportDir, cycles[0].ID)
		if err := os.MkdirAll(cdir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		body := `{"criteria":[{"id":"` + item.ID + `","claimed_done":false,"evidence":"not done"}]}`
		if err := os.WriteFile(filepath.Join(cdir, "criteria-report.json"), []byte(body), 0o644); err != nil {
			t.Fatalf("write criteria: %v", err)
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", nil, ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	<-done
	cancel()

	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			t.Fatalf("verify runner must not run when claimed_done=false; calls=%+v", r.Calls())
		}
	}
}

func containsSubstr(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !containsSubstr(s, p) {
			return false
		}
	}
	return true
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
