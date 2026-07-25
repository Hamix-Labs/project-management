package harness_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/harnesstest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestHarness_CancelRunForTask_idleIsNoOp(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	h := env.NewHarness(runnerfake.New(), harness.Options{})

	if h.CancelRunForTask("11111111-1111-4111-8111-111111111111") {
		t.Error("CancelRunForTask() = true, want false (idle)")
	}
	if h.CancelRunForTask("") {
		t.Error("CancelRunForTask(\"\") = true, want false")
	}
}

func TestHarness_CancelRunForTask_nonMatchingIDDoesNothing(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "close-nomatch"))

	br := harnesstest.NewBlockingRunner()
	h := env.NewHarness(br, harness.Options{RunTimeout: 0})
	done := env.RunHarness(ctx, h, tsk)

	select {
	case <-br.Starts:
	case <-time.After(harnesstest.DefaultPollTimeout):
		t.Fatal("runner did not start")
	}

	if h.CancelRunForTask("00000000-0000-4000-8000-000000000000") {
		t.Fatal("CancelRunForTask(unknown) = true, want false")
	}
	// Also confirm CancelCurrentRun still works, i.e. run is genuinely in flight.
	if !h.CancelCurrentRun() {
		t.Fatal("CancelCurrentRun() = false while runner was blocked")
	}
	<-done
	final := env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	if final.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task status = %q, want failed", final.Status)
	}
}

func TestHarness_CancelRunForTask_matchingIDCancels(t *testing.T) {
	t.Parallel()
	env := harnesstest.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := env.TransitionRunning(ctx, env.CreateReadyTask(ctx, "close-match"))

	br := harnesstest.NewBlockingRunner()
	h := env.NewHarness(br, harness.Options{RunTimeout: 0})
	done := env.RunHarness(ctx, h, tsk)

	select {
	case <-br.Starts:
	case <-time.After(harnesstest.DefaultPollTimeout):
		t.Fatal("runner did not start")
	}

	if !h.CancelRunForTask(tsk.ID) {
		t.Fatal("CancelRunForTask(match) = false, want true")
	}
	<-done
	final := env.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	if final.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task status = %q, want failed", final.Status)
	}
}
