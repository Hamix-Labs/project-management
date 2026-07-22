package harness_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/harnesstest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/metricsfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestHarness_RunMetrics_observesHappyPathOnce(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "metrics-happy"))

	r := runnerfake.New().WithName("fake")
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok",
		json.RawMessage(`{"ok":true}`), "",
	))

	metrics := metricsfake.New()
	done := env.RunHarness(ctx, env.NewHarness(r, harness.Options{Metrics: metrics}), tsk)
	<-done
	env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)

	calls := metrics.SnapshotRuns()
	if len(calls) != 1 {
		t.Fatalf("RecordRun calls = %d, want 1 (calls=%+v)", len(calls), calls)
	}
	if calls[0].Runner != "fake" {
		t.Fatalf("runner label = %q, want %q", calls[0].Runner, "fake")
	}
	if calls[0].TerminalStatus != string(cyclesdomain.CycleStatusSucceeded) {
		t.Fatalf("terminal_status = %q, want %q",
			calls[0].TerminalStatus, cyclesdomain.CycleStatusSucceeded)
	}
}

func TestHarness_RunMetrics_observesRunnerFailure(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "metrics-fail"))

	r := runnerfake.New().WithName("fake")
	r.FailWithResult(tsk.ID, cyclesdomain.PhaseExecute,
		runner.NewResult(cyclesdomain.PhaseStatusFailed, "exit 7",
			json.RawMessage(`{"exit_code":7}`), "stderr tail"),
		fmt.Errorf("cli exit: %w", runner.ErrNonZeroExit))

	metrics := metricsfake.New()
	done := env.RunHarness(ctx, env.NewHarness(r, harness.Options{Metrics: metrics}), tsk)
	<-done
	env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)

	calls := metrics.SnapshotRuns()
	if len(calls) != 1 {
		t.Fatalf("RecordRun calls = %d, want 1", len(calls))
	}
	if calls[0].TerminalStatus != string(cyclesdomain.CycleStatusFailed) {
		t.Fatalf("terminal_status = %q, want %q",
			calls[0].TerminalStatus, cyclesdomain.CycleStatusFailed)
	}
}

func TestHarness_RunMetrics_observesShutdownAbort(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "metrics-shutdown"))

	br := harnesstest.NewBlockingRunner()
	br.OnStart = func(req runner.Request) {
		cancel()
	}
	br.Result = runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "", nil, "")

	metrics := metricsfake.New()
	done := env.RunHarness(ctx, env.NewHarness(br, harness.Options{Metrics: metrics}), tsk)
	<-done

	calls := metrics.SnapshotRuns()
	if len(calls) != 1 {
		t.Fatalf("RecordRun calls = %d, want 1 (calls=%+v)", len(calls), calls)
	}
	if calls[0].TerminalStatus != string(cyclesdomain.CycleStatusAborted) {
		t.Fatalf("terminal_status = %q, want %q",
			calls[0].TerminalStatus, cyclesdomain.CycleStatusAborted)
	}
}

func TestHarness_RunMetrics_recordsEffectiveModelLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		taskModel     string
		runnerDefault string
		wantModel     string
	}{
		{name: "task_wins_over_default", taskModel: "sonnet-4.5", runnerDefault: "opus-4", wantModel: "sonnet-4.5"},
		{name: "fallback_to_runner_default", taskModel: "", runnerDefault: "opus-4", wantModel: "opus-4"},
		{name: "no_model_configured_anywhere", taskModel: "", runnerDefault: "", wantModel: ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := harnesstest.NewEnv(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tsk := env.TransitionRunning(ctx, env.CreateReadyTaskWithModel(ctx, "metrics-model-"+tc.name, tc.taskModel))

			r := runnerfake.New().WithName("fake").WithDefaultModel(tc.runnerDefault)
			r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
				cyclesdomain.PhaseStatusSucceeded, "ok",
				json.RawMessage(`{"ok":true}`), ""))

			metrics := metricsfake.New()
			done := env.RunHarness(ctx, env.NewHarness(r, harness.Options{Metrics: metrics}), tsk)
			<-done
			env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)

			calls := metrics.SnapshotRuns()
			if len(calls) != 1 {
				t.Fatalf("RecordRun calls = %d, want 1", len(calls))
			}
			if calls[0].Model != tc.wantModel {
				t.Fatalf("model label = %q, want %q", calls[0].Model, tc.wantModel)
			}
		})
	}
}

func TestHarness_RunMetrics_nilMetricsIsNoop(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "metrics-nil"))

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok", json.RawMessage(`{"ok":true}`), ""))

	done := env.RunHarness(ctx, env.NewHarness(r, harness.Options{}), tsk)
	<-done
	env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
}
