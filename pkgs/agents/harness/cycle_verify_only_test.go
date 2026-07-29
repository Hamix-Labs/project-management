package harness

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func testVerifyCmds() []checklistcontract.VerifyCommandInput {
	return []checklistcontract.VerifyCommandInput{{
		Command:         "echo ok",
		ExpectedOutcome: "prints ok",
	}}
}

type cycleVerifyHookRunner struct {
	runner.Runner
	preRun func(req runner.Request)
}

func (h *cycleVerifyHookRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	if h.preRun != nil {
		h.preRun(req)
	}
	return h.Runner.Run(ctx, req)
}

type infraFailRunner struct {
	runner.Runner
	inner   *runnerfake.Runner
	attempt atomic.Int32
}

func (r *infraFailRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	if req.Phase == cyclesdomain.PhaseVerify {
		n := r.attempt.Add(1)
		if n == 1 {
			return runner.NewResult(cyclesdomain.PhaseStatusFailed, "verify timeout", nil, ""), runner.ErrTimeout
		}
	}
	return r.Runner.Run(ctx, req)
}

func writeCriteriaReportCycleTest(t *testing.T, reportDir, cycleID string, ids []string) {
	t.Helper()
	cdir := filepath.Join(reportDir, cycleID)
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	type entry struct {
		ID          string `json:"id"`
		ClaimedDone bool   `json:"claimed_done"`
		Evidence    string `json:"evidence"`
	}
	rep := struct {
		Criteria []entry `json:"criteria"`
	}{}
	for _, id := range ids {
		rep.Criteria = append(rep.Criteria, entry{ID: id, ClaimedDone: true, Evidence: "execute did the thing"})
	}
	b, _ := json.Marshal(rep)
	if err := os.WriteFile(filepath.Join(cdir, "criteria-report.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writePartialVerifyReportCycleTest(t *testing.T, reportDir, cycleID string, verdicts map[string]bool) {
	t.Helper()
	cdir := filepath.Join(reportDir, cycleID)
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	type entry struct {
		ID        string `json:"id"`
		Verified  bool   `json:"verified"`
		Reasoning string `json:"reasoning"`
	}
	rep := struct {
		Criteria []entry `json:"criteria"`
	}{}
	for id, verified := range verdicts {
		reasoning := "verifier confirmed via diff inspection and detailed file content review of the change set under test"
		if !verified {
			reasoning = "verifier rejected: the implementation does not satisfy this criterion based on diff inspection"
		}
		rep.Criteria = append(rep.Criteria, entry{ID: id, Verified: verified, Reasoning: reasoning})
	}
	b, _ := json.Marshal(rep)
	if err := os.WriteFile(filepath.Join(cdir, "verify-report.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func startVerifyOnlyTask(t *testing.T, maxRetries int, extraItems ...string) (*composition.API, *taskcoredomain.Task, []string) {
	t.Helper()
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "verify-only", InitialPrompt: "work", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	item, err := st.AddChecklistItem(ctx, tsk.ID, "criterion", testVerifyCmds(), taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	ids = append(ids, item.ID)
	for i, title := range extraItems {
		it, err := st.AddChecklistItem(ctx, tsk.ID, title, testVerifyCmds(), taskcoredomain.ActorUser)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, it.ID)
		_ = i
	}
	if _, err := st.UpdateSettings(ctx, settingscontract.SettingsPatch{VerifyMaxRetries: &maxRetries}); err != nil {
		t.Fatal(err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	return st, tsk, ids
}

func TestEdgeCase_EC01_verifyInfra_oneShotTerminate(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 1)
	itemID := ids[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workDir := t.TempDir()
	reportDir := t.TempDir()
	execBase := runnerfake.New()
	execRunner := &infraFailRunner{Runner: execBase, inner: execBase}
	execHook := &cycleVerifyHookRunner{Runner: execRunner, preRun: func(req runner.Request) {
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		if req.Phase == cyclesdomain.PhaseExecute {
			writeCriteriaReportCycleTest(t, reportDir, cycles[0].ID, []string{itemID})
		}
	}}
	execBase.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	execBase.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, execHook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock: func() time.Time { return time.Unix(0, 0).UTC() },
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.Run(ctx, tsk)
	}()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for task failed")
		default:
		}
		got, err := st.Get(ctx, tsk.ID)
		if err == nil && got.Status == taskcoredomain.StatusFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if execCalls := countPhaseCalls(execBase, cyclesdomain.PhaseExecute); execCalls != 1 {
		t.Fatalf("execute calls = %d, want 1 (one-shot)", execCalls)
	}
	if execRunner.attempt.Load() != 1 {
		t.Fatalf("verify attempts = %d, want 1 (infra failure terminates)", execRunner.attempt.Load())
	}
}

func TestEdgeCase_EC02_verifyAgentReject_oneShotTerminate(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 1)
	itemID := ids[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workDir := t.TempDir()
	reportDir := t.TempDir()
	var execAttempt atomic.Int32
	r := runnerfake.New()
	hook := &cycleVerifyHookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		switch req.Phase {
		case cyclesdomain.PhaseExecute:
			writeCriteriaReportCycleTest(t, reportDir, cycles[0].ID, []string{itemID})
			execAttempt.Add(1)
		case cyclesdomain.PhaseVerify:
			writePartialVerifyReportCycleTest(t, reportDir, cycles[0].ID, map[string]bool{itemID: false})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	r.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, hook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock: func() time.Time { return time.Unix(0, 0).UTC() },
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.Run(ctx, tsk)
	}()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout")
		default:
		}
		got, err := st.Get(ctx, tsk.ID)
		if err == nil && got.Status == taskcoredomain.StatusFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if execAttempt.Load() != 1 {
		t.Fatalf("execute attempts = %d, want 1 (one-shot on verify-agent reject)", execAttempt.Load())
	}
}

func TestEdgeCase_EC03_claimedNotDone_oneShotTerminate(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 1)
	itemID := ids[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workDir := t.TempDir()
	reportDir := t.TempDir()
	var execAttempt atomic.Int32
	r := runnerfake.New()
	hook := &cycleVerifyHookRunner{Runner: r, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseExecute {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		execAttempt.Add(1)
		cdir := filepath.Join(reportDir, cycles[0].ID)
		_ = os.MkdirAll(cdir, 0o755)
		body := `{"criteria":[{"id":"` + itemID + `","claimed_done":false,"evidence":"not done"}]}`
		_ = os.WriteFile(filepath.Join(cdir, "criteria-report.json"), []byte(body), 0o644)
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, hook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock: func() time.Time { return time.Unix(0, 0).UTC() },
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.Run(ctx, tsk)
	}()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for one-shot terminate after claimed_not_done")
		default:
		}
		got, err := st.Get(ctx, tsk.ID)
		if err == nil && got.Status == taskcoredomain.StatusFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if execAttempt.Load() != 1 {
		t.Fatalf("execute attempts = %d, want 1 for claimed_not_done one-shot", execAttempt.Load())
	}
}

func TestEdgeCase_EC04_reportMissing_oneShotTerminate(t *testing.T) {
	st, tsk, _ := startVerifyOnlyTask(t, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	h := New(st, r, Options{WorkingDir: t.TempDir(), Clock: func() time.Time { return time.Unix(0, 0).UTC() }})
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.Run(ctx, tsk)
	}()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for one-shot terminate")
		default:
		}
		got, err := st.Get(ctx, tsk.ID)
		if err == nil && got.Status == taskcoredomain.StatusFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if execCalls := countPhaseCalls(r, cyclesdomain.PhaseExecute); execCalls != 1 {
		t.Fatalf("execute calls = %d, want 1 when criteria-report missing (one-shot)", execCalls)
	}
}

func countPhaseCalls(r *runnerfake.Runner, phase cyclesdomain.Phase) int {
	n := 0
	for _, c := range r.Calls() {
		if c.Phase == phase {
			n++
		}
	}
	return n
}

func TestEdgeCase_EC09_partialPass_infraOneShotTerminate(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 2, "criterion two")
	c1ID, c2ID := ids[0], ids[1]
	workDir := t.TempDir()
	reportDir := t.TempDir()
	execBase := runnerfake.New()
	execRunner := &infraFailRunner{Runner: execBase, inner: execBase}
	execHook := &cycleVerifyHookRunner{Runner: execRunner, preRun: func(req runner.Request) {
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		if req.Phase == cyclesdomain.PhaseExecute {
			writeCriteriaReportCycleTest(t, reportDir, cycles[0].ID, []string{c1ID, c2ID})
		}
	}}
	execBase.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	execBase.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, execHook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock: func() time.Time { return time.Unix(0, 0).UTC() },
	})
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.Run(runCtx, tsk)
	}()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout")
		default:
		}
		got, err := st.Get(runCtx, tsk.ID)
		if err == nil && got.Status == taskcoredomain.StatusFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if n := countPhaseCalls(execBase, cyclesdomain.PhaseExecute); n != 1 {
		t.Fatalf("execute calls = %d, want 1", n)
	}
	if execRunner.attempt.Load() != 1 {
		t.Fatalf("verify attempts = %d, want 1 (infra failure terminates)", execRunner.attempt.Load())
	}
}
