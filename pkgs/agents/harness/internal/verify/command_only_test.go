package verify_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestWorker_VerifyPhase_claimOnlyNoVerifyRunner(t *testing.T) {
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

func TestWorker_VerifyPhase_mixedClaimOnlyAndCommandBacked(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.CreateReadyTask(ctx, "mixed-verify")
	claimOnly, err := h.Store.AddChecklistItem(ctx, tsk.ID, "claim only", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add claim-only: %v", err)
	}
	commandBacked, err := h.Store.AddChecklistItem(ctx, tsk.ID, "command backed", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add command-backed: %v", err)
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
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{claimOnly.ID, commandBacked.ID})
		case cyclesdomain.PhaseVerify:
			writeVerifyReport(t, reportDir, cycles[0].ID, []string{commandBacked.ID})
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
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	<-done
	cancel()

	verifyCalls := 0
	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			verifyCalls++
			if !strings.Contains(c.Prompt, commandBacked.ID) {
				t.Fatalf("verify prompt should include command-backed id; prompt=%q", c.Prompt)
			}
			if strings.Contains(c.Prompt, claimOnly.ID) {
				t.Fatalf("verify prompt must not include claim-only id %q", claimOnly.ID)
			}
		}
	}
	if verifyCalls != 1 {
		t.Fatalf("verify calls = %d, want 1 (LLM only for command-backed)", verifyCalls)
	}

	items, err := h.Store.ListChecklistForSubject(context.Background(), tsk.ID)
	if err != nil {
		t.Fatalf("list checklist: %v", err)
	}
	byID := map[string]string{}
	for _, it := range items {
		if it.Done {
			byID[it.ID] = it.VerifiedBy
		}
	}
	if byID[claimOnly.ID] != string(checklistdomain.VerifierExecuteClaim) {
		t.Fatalf("claim-only verified_by = %q, want execute_claim", byID[claimOnly.ID])
	}
	if byID[commandBacked.ID] != string(checklistdomain.VerifierExecuteAgent) {
		t.Fatalf("command-backed verified_by = %q, want execute_agent", byID[commandBacked.ID])
	}
}

func TestWorker_VerifyPhase_claimedNotDoneTerminatesAgentSelf(t *testing.T) {
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

	metrics := newRecordingMetrics()
	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
		Metrics:    metrics,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	<-done
	cancel()

	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			t.Fatalf("verify runner must not run when claimed_done=false; calls=%+v", r.Calls())
		}
	}

	disagreements := 0
	for _, v := range metrics.verdictSnapshot() {
		if v.Kind == checklistdomain.VerifierAgentSelf && !v.Passed {
			disagreements++
		}
	}
	if disagreements != 1 {
		t.Fatalf("agent_self/failed verdict count = %d, want 1", disagreements)
	}
}

// Recovery after one-shot verification_failed: Start over (RetryFresh) and
// Retry (RetryResume) both open a new cycle that can succeed.
func TestWorker_VerifyPhase_recoveryRetryAndStartOver(t *testing.T) {
	t.Parallel()
	for _, mode := range []taskcoredomain.RetryMode{
		taskcoredomain.RetryFresh,
		taskcoredomain.RetryResume,
	} {
		mode := mode
		t.Run(string(mode), func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tsk := h.CreateReadyTask(ctx, "recover-"+string(mode))
			item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion", testVerifyCmds(), taskcoredomain.ActorUser)
			if err != nil {
				t.Fatalf("add checklist: %v", err)
			}

			workDir := t.TempDir()
			reportDir := t.TempDir()
			r1 := runnerfake.New()
			r1.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
				cyclesdomain.PhaseStatusSucceeded, "exec ok", nil, ""))
			// No criteria report ΓåÆ verification_failed (one-shot).
			done1 := h.StartHarnessRun(ctx, tsk, r1, harness.Options{
				WorkingDir: workDir,
				ReportDir:  reportDir,
			})
			h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
			<-done1

			cycles, err := h.Store.ListCyclesForTask(context.Background(), tsk.ID, 1)
			if err != nil || len(cycles) != 1 {
				t.Fatalf("parent cycles: %v %+v", err, cycles)
			}
			parentID := cycles[0].ID

			r2 := runnerfake.New()
			hook := &hookRunner{Runner: r2, preRun: func(req runner.Request) {
				cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
				if len(cycles) == 0 {
					return
				}
				switch req.Phase {
				case cyclesdomain.PhaseExecute:
					writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
				case cyclesdomain.PhaseVerify:
					writeVerifyReport(t, reportDir, cycles[0].ID, []string{item.ID})
				}
			}}
			r2.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
				cyclesdomain.PhaseStatusSucceeded, "exec ok", nil, ""))
			r2.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(
				cyclesdomain.PhaseStatusSucceeded, "verify ok", nil, ""))

			har := h.NewHarness(hook, harness.Options{
				WorkingDir: workDir,
				ReportDir:  reportDir,
			})
			tsk = h.TransitionRunning(ctx, tsk)
			done2 := make(chan struct{})
			go func() {
				defer close(done2)
				har.RunWithRetry(ctx, tsk, &taskcoredomain.PendingRetry{
					Mode:          mode,
					ParentCycleID: parentID,
				})
			}()
			h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
			<-done2
			cancel()
		})
	}
}
