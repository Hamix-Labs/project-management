package verify_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// EC-07 (docs/domain/harness.md): verify tamper ΓåÆ terminal failure, no execute retry.
// TestWorker_VerifyPhase_failsCycleWhenVerifyTampers pins the
// integrity-check contract. A verify runner that mutates source files
// MUST cause the cycle to terminate as verify_tampered with no
// retries, regardless of settings. Tampering is verifier
// misbehaviour; retrying execute cannot fix it.
func TestWorker_VerifyPhase_failsCycleWhenVerifyTampers(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "verify-tampers")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	gitInitTestRepo(t, workDir)

	r := runnerfake.New().WithName("execute-runner")
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		switch req.Phase {
		case cyclesdomain.PhaseExecute:
			writeCriteriaReportWithGitWork(t, reportDir, cycles[0].ID, workDir, []string{item.ID})
		case cyclesdomain.PhaseVerify:
			writeVerifyReport(t, reportDir, cycles[0].ID, []string{item.ID})
			// Tamper: drop a stray file in the working dir root. After
			// PR1 the integrity-check whitelist is empty (reports live
			// outside RepoRoot), so any RepoRoot mutation is tampering.
			if err := os.WriteFile(filepath.Join(workDir, "MUTATED.txt"), []byte("hi"), 0o644); err != nil {
				t.Logf("tamper write: %v", err)
			}
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
	final := h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	<-done
	cancel()
	if final.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task status = %q, want failed", final.Status)
	}

	bg := context.Background()
	cycles, _ := h.Store.ListCyclesForTask(bg, tsk.ID, 5)
	if len(cycles) != 1 {
		t.Fatalf("cycle count = %d, want 1 (no retries on tamper)", len(cycles))
	}
	if cycles[0].Status != cyclesdomain.CycleStatusFailed {
		t.Fatalf("cycle status = %q, want failed", cycles[0].Status)
	}

	events, _ := h.Store.ListTaskEvents(bg, tsk.ID)
	sawTampered := false
	for _, e := range events {
		if e.Type != taskeventsdomain.EventCycleFailed {
			continue
		}
		if strings.Contains(string(e.Data), "verify_tampered") {
			sawTampered = true
		}
	}
	if !sawTampered {
		t.Fatalf("expected cycle_failed event with reason=verify_tampered; events=%+v", events)
	}

	// Verify must have been invoked exactly once: tampering is
	// terminal, retries do not run.
	verifyCallCount := 0
	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			verifyCallCount++
		}
	}
	if verifyCallCount != 1 {
		t.Fatalf("execute runner verify calls = %d, want 1 (terminal-not-retryable)", verifyCallCount)
	}
}

// TestWorker_VerifyPhase_finalFailureWritesNoCompletions pins the
// atomic-decision contract: on one-shot verify failure, NO completion
// rows land in task_checklist_completions even when some criteria
// passed. lockedPasses are in-memory only until finalize succeeds.
func TestWorker_VerifyPhase_finalFailureWritesNoCompletions(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "verify-no-completion")
	c1, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add c1: %v", err)
	}
	c2, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion two", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add c2: %v", err)
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
			writePartialVerifyReport(t, reportDir, cycles[0].ID, map[string]bool{c1.ID: true, c2.ID: false})
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
			t.Errorf("expected NO completed items on terminal failure; %s is done", it.ID)
		}
	}
}

// TestWorker_VerifyPhase_terminateReasonIncludesFailingIDs pins the
// SPA-renderable failure detail: on one-shot verify failure the cycle's
// terminate_reason carries the failing criterion IDs after the
// stable `verification_failed:` prefix.
func TestWorker_VerifyPhase_terminateReasonIncludesFailingIDs(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "verify-reason-ids")
	c1, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add c1: %v", err)
	}
	c2, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion two", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add c2: %v", err)
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
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{c1.ID, c2.ID})
		case cyclesdomain.PhaseVerify:
			writePartialVerifyReport(t, reportDir, cycles[0].ID, map[string]bool{
				c1.ID: false, c2.ID: false,
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
	events, err := h.Store.ListTaskEvents(bg, tsk.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	var reason string
	for _, e := range events {
		if e.Type != taskeventsdomain.EventCycleFailed {
			continue
		}
		var payload struct {
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(e.Data, &payload); err != nil {
			continue
		}
		if strings.HasPrefix(payload.Reason, "verification_failed") {
			reason = payload.Reason
		}
	}
	if reason == "" {
		t.Fatalf("no cycle_failed event with verification_failed reason; events=%+v", events)
	}
	if !strings.HasPrefix(reason, "verification_failed:") {
		t.Fatalf("reason must start with verification_failed:; got %q", reason)
	}
	// IDs are sorted; assert both appear regardless of seed order.
	if !strings.Contains(reason, c1.ID) || !strings.Contains(reason, c2.ID) {
		t.Fatalf("reason must include both failing IDs; got %q (c1=%s c2=%s)", reason, c1.ID, c2.ID)
	}
}

// EC-07: repo-root mutation during verify is tamper (terminal).
// TestWorker_VerifyPhase_repoRootMutationStillTampered pins the
// strengthened integrity contract: with the report-file allowlist
// removed in PR1, ANY mutation under RepoRoot during the verify pass
// is tampering. Even paths that mimic the legacy `.legacy-scratch/<cycleID>/...`
// shape are no longer tolerated ΓÇö the verifier has no business
// touching the working tree.
func TestWorker_VerifyPhase_repoRootMutationStillTampered(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "verify-no-allowlist")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	gitInitTestRepo(t, workDir)

	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		switch req.Phase {
		case cyclesdomain.PhaseExecute:
			writeCriteriaReportWithGitWork(t, reportDir, cycles[0].ID, workDir, []string{item.ID})
		case cyclesdomain.PhaseVerify:
			writeVerifyReport(t, reportDir, cycles[0].ID, []string{item.ID})
			// Drop a fake legacy-shaped artifact INSIDE the working
			// tree. Pre-PR1 this would have been tolerated by the
			// allowlist; post-PR1 it must trip integrity.
			legacyDir := filepath.Join(workDir, ".legacy-scratch", cycles[0].ID)
			_ = os.MkdirAll(legacyDir, 0o755)
			_ = os.WriteFile(filepath.Join(legacyDir, "verify-report.json"), []byte("{}"), 0o644)
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
	events, _ := h.Store.ListTaskEvents(context.Background(), tsk.ID)
	sawTampered := false
	for _, e := range events {
		if e.Type == taskeventsdomain.EventCycleFailed && strings.Contains(string(e.Data), "verify_tampered") {
			sawTampered = true
		}
	}
	if !sawTampered {
		t.Fatalf("expected verify_tampered cycle_failed event after legacy-shaped RepoRoot write; events=%+v", events)
	}
}
