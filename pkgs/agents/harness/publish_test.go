package harness_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/harnesstest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/notifierfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestHarness_HappyPath_emitsTrailingPublishAfterTerminalStatus(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "trailing-publish"))

	snap := &harnesstest.StatusSnappingNotifier{Store: env.Store}
	taskUpdated := notifierfake.NewRecordingTaskUpdatedNotifier()

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "all green",
		json.RawMessage(`{"ok":true}`), "",
	))

	done := env.RunHarness(ctx, env.NewHarness(r, harness.Options{
		Notifier:            snap,
		TaskUpdatedNotifier: taskUpdated,
	}), tsk)
	<-done
	final := env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	if final.Status != taskcoredomain.StatusReview {
		t.Fatalf("task status = %q, want done", final.Status)
	}

	statuses, cycles := snap.Snapshot()
	if len(statuses) == 0 {
		t.Fatal("notifier received zero publishes")
	}
	if got := statuses[len(statuses)-1]; got != taskcoredomain.StatusReview {
		t.Fatalf("last publish observed task status = %q, want done; full snapshot=%+v", got, statuses)
	}
	if cycles[len(cycles)-1] == "" {
		t.Fatal("trailing publish used empty cycle id")
	}

	taskIDs := taskUpdated.Snapshot()
	if len(taskIDs) != 1 {
		t.Fatalf("task_updated publishes: got %d want 1 (%v)", len(taskIDs), taskIDs)
	}
	if taskIDs[0] != tsk.ID {
		t.Fatalf("task_updated task id = %q, want %q", taskIDs[0], tsk.ID)
	}
}

func TestHarness_PublishesRunnerProgressWithCycleAndPhaseContext(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "live-progress"))
	progress := notifierfake.NewRecordingProgressNotifier()
	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "all green", nil, ""))
	r.ScriptProgress(tsk.ID, cyclesdomain.PhaseExecute, runner.ProgressEvent{
		Kind:    "tool_call",
		Subtype: "started",
		Tool:    "ReadFile",
		Message: "Started ReadFile",
		Payload: json.RawMessage(`{"type":"tool_call","name":"ReadFile","input":{"path":"README.md"}}`),
	})

	done := env.RunHarness(ctx, env.NewHarness(r, harness.Options{ProgressNotifier: progress}), tsk)
	<-done
	env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)

	calls := progress.Snapshot()
	if len(calls) != 1 {
		t.Fatalf("progress calls: got %d want 1 (%+v)", len(calls), calls)
	}
	got := calls[0]
	if got.TaskID != tsk.ID {
		t.Fatalf("TaskID: got %q want %q", got.TaskID, tsk.ID)
	}
	if got.CycleID == "" {
		t.Fatal("CycleID must be populated")
	}
	if got.PhaseSeq != 1 {
		t.Fatalf("PhaseSeq: got %d want 1", got.PhaseSeq)
	}
	if got.RunCorrelationID == "" {
		t.Fatal("RunCorrelationID must be populated")
	}
	stream, err := env.Store.ListCycleStreamEvents(context.Background(), got.CycleID, 0, 10)
	if err != nil {
		t.Fatalf("list persisted progress: %v", err)
	}
	if len(stream) != 1 {
		t.Fatalf("persisted stream events: got %d want 1", len(stream))
	}
}
