package worker_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

func TestWorker_RunnerFailure_marksCycleAndTaskFailed(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "boom")

	r := runnerfake.New()
	r.FailWithResult(tsk.ID, cyclesdomain.PhaseExecute,
		runner.NewResult(cyclesdomain.PhaseStatusFailed, "exit 7", json.RawMessage(`{"exit_code":7}`), "stderr tail"),
		fmt.Errorf("cli exit: %w", runner.ErrNonZeroExit))

	_, done := h.startWorker(ctx, r, worker.Options{})
	h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	bg := context.Background()
	cycle := assertCycleStatus(t, h.store, tsk.ID, 1, cyclesdomain.CycleStatusFailed)
	phases, _ := h.store.ListPhasesForCycle(bg, cycle.ID)
	if len(phases) != 1 {
		t.Fatalf("phase count = %d, want 1", len(phases))
	}
	if phases[0].Phase != cyclesdomain.PhaseExecute || phases[0].Status != cyclesdomain.PhaseStatusFailed {
		t.Fatalf("execute phase = %q/%q, want execute/failed", phases[0].Phase, phases[0].Status)
	}

	events, _ := h.store.ListTaskEvents(bg, tsk.ID)
	counts := eventTypeCounts(events)
	if counts[taskeventsdomain.EventCycleFailed] != 1 {
		t.Fatalf("cycle_failed count = %d, want 1", counts[taskeventsdomain.EventCycleFailed])
	}
	if counts[taskeventsdomain.EventPhaseFailed] != 1 {
		t.Fatalf("phase_failed count = %d, want 1", counts[taskeventsdomain.EventPhaseFailed])
	}

	if got := h.queue.BufferDepth(); got != 0 {
		t.Fatalf("queue depth = %d, want 0 (acked)", got)
	}
}

func TestWorker_StaleTaskAtDequeue_ackAndSkip(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "stale")

	// Move the task off `ready` AFTER it was enqueued by Create.
	doneStatus := taskcoredomain.StatusDone
	if _, err := h.store.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &doneStatus}, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("update to done: %v", err)
	}

	r := runnerfake.New()
	_, done := h.startWorker(ctx, r, worker.Options{})

	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) && h.queue.BufferDepth() > 0 {
		time.Sleep(pollInterval)
	}
	if h.queue.BufferDepth() != 0 {
		t.Fatalf("queue still has %d after stale dequeue", h.queue.BufferDepth())
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	h.waitNoCycle(ctx, tsk.ID)
	if calls := h.notifier.snapshot(); len(calls) != 0 {
		t.Fatalf("notifier publish count = %d, want 0 on stale", len(calls))
	}
	if calls := r.Calls(); len(calls) != 0 {
		t.Fatalf("runner Run calls = %d, want 0 on stale", len(calls))
	}
}

func TestWorker_TaskDeletedMidCycle_logsAndAcks(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "deleteme")

	br := newBlockingRunner()
	br.onStart = func(req runner.Request) {
		// Cascade-deletes the cycle + phase rows.
		if _, err := h.store.Delete(context.Background(), tsk.ID, taskcoredomain.ActorUser); err != nil {
			t.Logf("delete during run: %v", err)
		}
		close(br.release)
	}
	br.result = runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "", nil, "")

	_, done := h.startWorker(ctx, br, worker.Options{})

	bg := context.Background()
	deadline := time.Now().Add(pollTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		_, err := h.store.Get(bg, tsk.ID)
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			lastErr = err
			break
		}
		lastErr = err
		time.Sleep(pollInterval)
	}
	if !errors.Is(lastErr, taskcoredomain.ErrNotFound) {
		t.Fatalf("expected task deleted, last err=%v", lastErr)
	}

	deadline = time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) && h.queue.BufferDepth() > 0 {
		time.Sleep(pollInterval)
	}
	if h.queue.BufferDepth() != 0 {
		t.Fatalf("queue depth = %d, want 0 after delete", h.queue.BufferDepth())
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}
}

func TestWorker_PanicInRunner_terminatesAndContinues(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	first := h.createReadyTask(ctx, "panicker")
	second := h.createReadyTask(ctx, "after-panic")

	r := runnerfake.New()
	r.Script(second.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

	// Wrap the fake to panic on the first task and delegate to the
	// fake on the second so the loop must keep going.
	pr := &panickyRunner{
		Runner:    r,
		panicTask: first.ID,
	}

	_, done := h.startWorker(ctx, pr, worker.Options{})

	h.waitTaskStatus(ctx, first.ID, taskcoredomain.StatusFailed)
	h.waitTaskStatus(ctx, second.ID, taskcoredomain.StatusDone)

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	bg := context.Background()
	c := assertCycleStatus(t, h.store, first.ID, 1, cyclesdomain.CycleStatusFailed)
	phases, _ := h.store.ListPhasesForCycle(bg, c.ID)
	if len(phases) != 1 {
		t.Fatalf("panic cycle phase count = %d, want 1", len(phases))
	}
	if phases[0].Phase != cyclesdomain.PhaseExecute || phases[0].Status != cyclesdomain.PhaseStatusFailed {
		t.Fatalf("execute phase after panic = %q/%q, want execute/failed", phases[0].Phase, phases[0].Status)
	}

	events, _ := h.store.ListTaskEvents(bg, first.ID)
	counts := eventTypeCounts(events)
	if counts[taskeventsdomain.EventCycleFailed] != 1 || counts[taskeventsdomain.EventPhaseFailed] != 1 {
		t.Fatalf("panic cycle event counts = %+v", counts)
	}
}

type panickyRunner struct {
	*runnerfake.Runner
	panicTask string
}

func (p *panickyRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	if req.TaskID == p.panicTask {
		panic("worker test induced panic")
	}
	return p.Runner.Run(ctx, req)
}
