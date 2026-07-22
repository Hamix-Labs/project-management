package harness

import (
	"context"
	"encoding/json"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

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

type infraFailVerifyRunner struct {
	runner.Runner
	inner   *runnerfake.Runner
	attempt atomic.Int32
}

func (r *infraFailVerifyRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
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
	item, err := st.AddChecklistItem(ctx, tsk.ID, "criterion", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	ids = append(ids, item.ID)
	for i, title := range extraItems {
		it, err := st.AddChecklistItem(ctx, tsk.ID, title, nil, taskcoredomain.ActorUser)
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

func TestEdgeCase_EC01_verifyInfra_skipsExecute(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 1)
	itemID := ids[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workDir := t.TempDir()
	reportDir := t.TempDir()
	execRunner := runnerfake.New()
	execHook := &cycleVerifyHookRunner{Runner: execRunner, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseExecute {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		writeCriteriaReportCycleTest(t, reportDir, cycles[0].ID, []string{itemID})
	}}
	execRunner.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	verifyBase := runnerfake.New()
	verifyRunner := &infraFailVerifyRunner{Runner: verifyBase, inner: verifyBase}
	verifyHook := &cycleVerifyHookRunner{Runner: verifyRunner, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseVerify {
			return
		}
		// attempt is 1 after the first infra failure; preRun runs before the second Run.
		if verifyRunner.attempt.Load() != 1 {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		writePartialVerifyReportCycleTest(t, reportDir, cycles[0].ID, map[string]bool{itemID: true})
	}}
	verifyBase.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, execHook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock:        func() time.Time { return time.Unix(0, 0).UTC() },
		VerifyRunner: verifyHook,
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
			t.Fatal("timeout waiting for task done")
		default:
		}
		got, err := st.Get(ctx, tsk.ID)
		if err == nil && got.Status == taskcoredomain.StatusReview {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	execCalls := 0
	for _, c := range execRunner.Calls() {
		if c.Phase == cyclesdomain.PhaseExecute {
			execCalls++
		}
	}
	if execCalls != 1 {
		t.Fatalf("execute calls = %d, want 1 (verify-only retry)", execCalls)
	}
	if verifyRunner.attempt.Load() != 2 {
		t.Fatalf("verify attempts = %d, want 2", verifyRunner.attempt.Load())
	}
}

func TestEdgeCase_EC02_verifyAgentReject_fullReexecute(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 1)
	itemID := ids[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workDir := t.TempDir()
	reportDir := t.TempDir()
	var execAttempt atomic.Int32
	execRunner := runnerfake.New()
	execHook := &cycleVerifyHookRunner{Runner: execRunner, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseExecute {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		writeCriteriaReportCycleTest(t, reportDir, cycles[0].ID, []string{itemID})
		execAttempt.Add(1)
	}}
	execRunner.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	verifyRunner := runnerfake.New()
	verifyHook := &cycleVerifyHookRunner{Runner: verifyRunner, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseVerify {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		n := len(verifyRunner.Calls())
		if n == 0 {
			writePartialVerifyReportCycleTest(t, reportDir, cycles[0].ID, map[string]bool{itemID: false})
		} else {
			writePartialVerifyReportCycleTest(t, reportDir, cycles[0].ID, map[string]bool{itemID: true})
		}
	}}
	verifyRunner.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, execHook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock:        func() time.Time { return time.Unix(0, 0).UTC() },
		VerifyRunner: verifyHook,
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
		if err == nil && got.Status == taskcoredomain.StatusReview {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if execAttempt.Load() != 2 {
		t.Fatalf("execute attempts = %d, want 2 (full re-execute on verify-agent reject)", execAttempt.Load())
	}
}

func TestEdgeCase_EC03_claimedNotDone_fullReexecute(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 1)
	itemID := ids[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workDir := t.TempDir()
	reportDir := t.TempDir()
	var execAttempt atomic.Int32
	execRunner := runnerfake.New()
	execHook := &cycleVerifyHookRunner{Runner: execRunner, preRun: func(req runner.Request) {
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
	execRunner.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	verifyRunner := runnerfake.New()
	verifyRunner.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, execHook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock:        func() time.Time { return time.Unix(0, 0).UTC() },
		VerifyRunner: verifyRunner,
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
			t.Fatal("timeout waiting for full re-execute after claimed_not_done")
		default:
		}
		if execAttempt.Load() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if execAttempt.Load() < 2 {
		t.Fatalf("execute attempts = %d, want >=2 for claimed_not_done retry", execAttempt.Load())
	}
}

func TestEdgeCase_EC04_reportMissing_fullReexecute(t *testing.T) {
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
			t.Fatal("timeout waiting for full re-execute")
		default:
		}
		execCalls := 0
		for _, c := range r.Calls() {
			if c.Phase == cyclesdomain.PhaseExecute {
				execCalls++
			}
		}
		if execCalls >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	execCalls := 0
	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseExecute {
			execCalls++
		}
	}
	if execCalls != 2 {
		t.Fatalf("execute calls = %d, want 2 when criteria-report missing", execCalls)
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

func TestEdgeCase_EC09_partialPass_infraVerifyOnly(t *testing.T) {
	st, tsk, ids := startVerifyOnlyTask(t, 2, "criterion two")
	c1ID, c2ID := ids[0], ids[1]
	workDir := t.TempDir()
	reportDir := t.TempDir()
	execRunner := runnerfake.New()
	execHook := &cycleVerifyHookRunner{Runner: execRunner, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseExecute {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		writeCriteriaReportCycleTest(t, reportDir, cycles[0].ID, []string{c1ID, c2ID})
	}}
	execRunner.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	verifyBase := runnerfake.New()
	verifyRunner := &infraFailVerifyRunner{Runner: verifyBase, inner: verifyBase}
	verifyHook := &cycleVerifyHookRunner{Runner: verifyRunner, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseVerify {
			return
		}
		cycles, _ := st.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		n := verifyRunner.attempt.Load()
		if n != 1 {
			return
		}
		writePartialVerifyReportCycleTest(t, reportDir, cycles[0].ID, map[string]bool{c1ID: true, c2ID: true})
	}}
	verifyBase.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	h := New(st, execHook, Options{
		WorkingDir: workDir, ReportDir: reportDir,
		Clock:        func() time.Time { return time.Unix(0, 0).UTC() },
		VerifyRunner: verifyHook,
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
		if err == nil && got.Status == taskcoredomain.StatusReview {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-done
	cancel()

	if n := countPhaseCalls(execRunner, cyclesdomain.PhaseExecute); n != 1 {
		t.Fatalf("execute calls = %d, want 1", n)
	}
}
