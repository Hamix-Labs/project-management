package harness

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestRunWithRetry_openPRSkipsVerifyAndLandsPrReady(t *testing.T) {
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "open-pr", InitialPrompt: "work", Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	item, err := st.AddChecklistItem(ctx, tsk.ID, "criterion", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetChecklistItemDone(ctx, tsk.ID, item.ID, true, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	parent, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.TerminateCycle(ctx, parent.ID, cyclesdomain.CycleStatusSucceeded, "", taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	review := taskcoredomain.StatusReview
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &review}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	running = taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	reportDir := t.TempDir()
	base := runnerfake.New()
	base.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	r := &retryHookRunner{Runner: base, preRun: func(req runner.Request) {
		if req.Phase != cyclesdomain.PhaseExecute {
			return
		}
		cycles, err := st.ListCyclesForTask(ctx, tsk.ID, 1)
		if err != nil || len(cycles) == 0 {
			t.Fatalf("list cycles: %v len=%d", err, len(cycles))
		}
		if err := sidecar.WritePullRequestReport(reportDir, cycles[0].ID, sidecar.PullRequestReport{
			URL:    "https://github.com/example/repo/pull/42",
			Number: 42,
			Title:  "open-pr",
			Base:   "main",
			Head:   "feat/open-pr",
		}); err != nil {
			t.Fatalf("write pull-request report: %v", err)
		}
	}}
	h := New(st, r, Options{
		WorkingDir: t.TempDir(),
		ReportDir:  reportDir,
		Clock:      func() time.Time { return time.Unix(0, 0).UTC() },
	})
	intent := &taskcoredomain.PendingRetry{
		Kind:          taskcoredomain.PendingKindOpenPR,
		Mode:          taskcoredomain.RetryResume,
		ParentCycleID: parent.ID,
	}
	if err := intent.Validate(); err != nil {
		t.Fatal(err)
	}
	if !intent.SkipVerify {
		t.Fatal("expected SkipVerify")
	}
	h.RunWithRetry(ctx, tsk, intent)

	for _, call := range r.Calls() {
		if call.Phase == cyclesdomain.PhaseVerify {
			t.Fatalf("open_pr must skip verify; got %+v", call)
		}
	}
	if len(r.Calls()) == 0 {
		t.Fatal("expected execute runner call")
	}
	got, err := st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusPrReady {
		t.Fatalf("status=%q want pr_ready", got.Status)
	}
	cycles, err := st.ListCyclesForTask(ctx, tsk.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) < 2 {
		t.Fatalf("cycles=%d want >=2", len(cycles))
	}
	meta := string(cycles[0].MetaJSON)
	if !strings.Contains(meta, `"run_kind":"open_pr"`) {
		t.Fatalf("meta=%s want run_kind open_pr", meta)
	}
}

func TestRunWithRetry_openPRMissingReportFails(t *testing.T) {
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "open-pr-missing", InitialPrompt: "work", Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	item, err := st.AddChecklistItem(ctx, tsk.ID, "criterion", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetChecklistItemDone(ctx, tsk.ID, item.ID, true, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	parent, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.TerminateCycle(ctx, parent.ID, cyclesdomain.CycleStatusSucceeded, "", taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	review := taskcoredomain.StatusReview
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &review}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	running = taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	h := New(st, r, Options{
		WorkingDir: t.TempDir(),
		ReportDir:  t.TempDir(),
		Clock:      func() time.Time { return time.Unix(0, 0).UTC() },
	})
	intent := &taskcoredomain.PendingRetry{
		Kind:          taskcoredomain.PendingKindOpenPR,
		Mode:          taskcoredomain.RetryResume,
		ParentCycleID: parent.ID,
	}
	if err := intent.Validate(); err != nil {
		t.Fatal(err)
	}
	h.RunWithRetry(ctx, tsk, intent)

	got, err := st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusFailed {
		t.Fatalf("status=%q want failed", got.Status)
	}
}

type retryHookRunner struct {
	*runnerfake.Runner
	preRun func(req runner.Request)
}

func (h *retryHookRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	if h.preRun != nil {
		h.preRun(req)
	}
	return h.Runner.Run(ctx, req)
}
