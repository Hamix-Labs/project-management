package worker_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

func TestWorker_ShutdownMidRun_writesAbortedCycleAndFailedTask(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "shutdown")

	br := newBlockingRunner()
	cancelOnce := sync.Once{}
	br.onStart = func(req runner.Request) {
		cancelOnce.Do(func() {
			cancel()
		})
	}

	_, done := h.startWorker(ctx, br, worker.Options{
		ShutdownAbortTimeout: 2 * time.Second,
	})

	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	bg := context.Background()
	cycle := assertCycleStatus(t, h.store, tsk.ID, 1, cyclesdomain.CycleStatusAborted)

	final, err := h.store.Get(bg, tsk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if final.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task status after shutdown = %q, want failed", final.Status)
	}

	phases, _ := h.store.ListPhasesForCycle(bg, cycle.ID)
	if len(phases) != 1 {
		t.Fatalf("phase count after shutdown = %d, want 1", len(phases))
	}
	if phases[0].Phase != cyclesdomain.PhaseExecute || phases[0].Status != cyclesdomain.PhaseStatusFailed {
		t.Fatalf("execute phase after shutdown = %q/%q", phases[0].Phase, phases[0].Status)
	}
	if phases[0].Summary == nil || !strings.Contains(*phases[0].Summary, worker.ShutdownReason) {
		t.Fatalf("execute phase summary = %v, want contains %q", phases[0].Summary, worker.ShutdownReason)
	}

	events, _ := h.store.ListTaskEvents(bg, tsk.ID)
	counts := eventTypeCounts(events)
	if counts[taskeventsdomain.EventCycleFailed] != 1 {
		t.Fatalf("cycle_failed (aborted folds in) count = %d, want 1", counts[taskeventsdomain.EventCycleFailed])
	}
}

func TestWorker_NoDoubleCycleOnRedelivery(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "redeliver")

	br := newBlockingRunner()
	br.result = runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "", nil, "")

	// Direct in-test attempt to write a second cycle while one is
	// running surfaces ErrInvalidInput from the store guard. This
	// pins the substrate behaviour the worker relies on (edge case
	// from the plan: "no double cycle on redelivery").
	_, done := h.startWorker(ctx, br, worker.Options{})

	select {
	case req := <-br.starts:
		if req.TaskID != tsk.ID {
			t.Fatalf("first run task id = %s, want %s", req.TaskID, tsk.ID)
		}
	case <-time.After(pollTimeout):
		t.Fatal("timed out waiting for first runner Run")
	}

	_, err := h.store.StartCycle(context.Background(), cyclescontract.StartCycleInput{
		TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent,
	})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("second StartCycle err = %v, want ErrInvalidInput", err)
	}

	close(br.release)
	h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	cycles, _ := h.store.ListCyclesForTask(context.Background(), tsk.ID, 10)
	if len(cycles) != 1 {
		t.Fatalf("final cycle count = %d, want 1", len(cycles))
	}
}

func TestWorker_NilNotifierIsNoOp(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "no-notifier")

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "", nil, ""))

	w := worker.NewWorker(h.store, h.queue, r, worker.Options{Notifier: nil})
	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}
}

// TestWorker_RunReturnsOnNilDeps fast-fails when the constructor was
// fed nil dependencies and Run is invoked anyway. Pins the contract
// for cmd/taskapi: don't go w.Run(ctx) until you've supplied real
// store/queue/runner instances.
func TestWorker_RunReturnsOnNilDeps(t *testing.T) {
	t.Parallel()
	w := &worker.Worker{}
	if err := w.Run(context.Background()); err == nil {
		t.Fatal("expected error from nil-deps Run")
	}
}

// Sanity: NotifyReadyTask via Create + Update remains the supported
// enqueue path and the `BufferDepth` getter returns to zero after a
// happy run. Used by other tests indirectly; this one pins it.
func TestWorker_QueueDrainsAfterHappyRun(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "drain")
	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "", nil, ""))

	// Use atomic to keep the closure readonly to the linter.
	var ran atomic.Bool
	wrappedRunner := &funcRunner{
		name:    "wrap",
		version: "v1",
		run: func(ctx context.Context, req runner.Request) (runner.Result, error) {
			ran.Store(true)
			return r.Run(ctx, req)
		},
	}

	_, done := h.startWorker(ctx, wrappedRunner, worker.Options{})
	h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}
	if !ran.Load() {
		t.Fatal("wrapped runner was not invoked")
	}
	if got := h.queue.BufferDepth(); got != 0 {
		t.Fatalf("queue depth after happy run = %d, want 0", got)
	}
}

type funcRunner struct {
	name, version string
	run           func(ctx context.Context, req runner.Request) (runner.Result, error)
}

func (f *funcRunner) Name() string    { return f.name }
func (f *funcRunner) Version() string { return f.version }
func (f *funcRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	return f.run(ctx, req)
}
func (f *funcRunner) EffectiveModel(req runner.Request) string {
	return strings.TrimSpace(req.CursorModel)
}

// TestWorker_CompletePhaseFailure_terminatesCycleAndFailsTask pins the
// fix for the orphaned-cycle bug documented in worker.completeExecutePhase:
// when CompletePhase(execute) fails for any reason other than process
// shutdown or panic, the worker must still terminate the cycle and walk
// the task to `failed` so the next dequeue does not re-enter the loop
// and so the operator does not have to wait for the startup orphan
// sweep to clean up.
//
// The reproducer preempts the execute-phase row from inside the runner:
// before the runner returns its successful Result, it directly calls
// store.CompletePhase to mark the same phase row terminal. The worker's
// happy-path CompletePhase call then surfaces "phase already terminal"
// (taskcoredomain.ErrInvalidInput). Without the fix this strands the cycle in
// `running` and the task in `running` until the next process restart;
// with the fix the cycle is `failed` with the dedicated reason and the
// task is walked to `failed` synchronously.
func TestWorker_CompletePhaseFailure_terminatesCycleAndFailsTask(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "complete-phase-failure")

	bg := context.Background()
	preemptOnce := sync.Once{}
	var preempted atomic.Bool
	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "all green",
		json.RawMessage(`{"ok":true}`), ""))

	wrapped := &funcRunner{
		name:    "preempt",
		version: "v0",
		run: func(rctx context.Context, req runner.Request) (runner.Result, error) {
			preemptOnce.Do(func() {
				cycles, err := h.store.ListCyclesForTask(bg, req.TaskID, 5)
				if err != nil || len(cycles) == 0 {
					t.Errorf("preempt: list cycles: err=%v len=%d", err, len(cycles))
					return
				}
				phases, err := h.store.ListPhasesForCycle(bg, cycles[0].ID)
				if err != nil {
					t.Errorf("preempt: list phases: %v", err)
					return
				}
				for _, ph := range phases {
					if ph.Phase != cyclesdomain.PhaseExecute {
						continue
					}
					if cyclesdomain.TerminalPhaseStatus(ph.Status) {
						continue
					}
					summary := "preempted by test"
					if _, err := h.store.CompletePhase(bg, cyclescontract.CompletePhaseInput{
						CycleID:  cycles[0].ID,
						PhaseSeq: ph.PhaseSeq,
						Status:   cyclesdomain.PhaseStatusFailed,
						Summary:  &summary,
						By:       taskcoredomain.ActorAgent,
					}); err != nil {
						t.Errorf("preempt: CompletePhase: %v", err)
						return
					}
					preempted.Store(true)
					return
				}
				t.Errorf("preempt: no running execute phase to preempt")
			})
			return r.Run(rctx, req)
		},
	}

	_, done := h.startWorker(ctx, wrapped, worker.Options{})
	final := h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	if !preempted.Load() {
		t.Fatal("test setup did not preempt the execute phase; reproducer is a no-op")
	}

	if final.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task final status = %q, want %q", final.Status, taskcoredomain.StatusFailed)
	}

	cycles, err := h.store.ListCyclesForTask(bg, tsk.ID, 5)
	if err != nil {
		t.Fatalf("list cycles: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("cycle count = %d, want 1", len(cycles))
	}
	c := cycles[0]
	if c.Status == cyclesdomain.CycleStatusRunning || c.Status == "" {
		t.Fatalf("cycle status after CompletePhase failure = %q, want a terminal status (cycle was orphaned)", c.Status)
	}
	if c.Status != cyclesdomain.CycleStatusFailed {
		t.Fatalf("cycle status = %q, want %q (CompletePhase write failures must mark the cycle failed)", c.Status, cyclesdomain.CycleStatusFailed)
	}
}
