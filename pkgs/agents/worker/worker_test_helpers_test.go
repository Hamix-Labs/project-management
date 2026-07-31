package worker_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

const pollInterval = 10 * time.Millisecond
const pollTimeout = 3 * time.Second

// --- shared harness ------------------------------------------------------

type harness struct {
	t          *testing.T
	store      *composition.API
	queue      *agents.MemoryQueue
	notifier   *recordingNotifier
	worktreeID string
	workDir    string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	q := agents.NewMemoryQueue(8)
	st.SetReadyTaskNotifier(q)
	wtID, dir := seedWorkerTestGit(t, st)
	return &harness{
		t: t, store: st, queue: q, notifier: newRecordingNotifier(),
		worktreeID: wtID, workDir: dir,
	}
}

func (h *harness) createReadyTask(ctx context.Context, title string) *taskcoredomain.Task {
	h.t.Helper()
	wb := h.gitBinding()
	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         title,
		InitialPrompt: "do the thing",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		WorktreeID:    wb,
	}, taskcoredomain.ActorUser)
	if err != nil {
		h.t.Fatalf("create task: %v", err)
	}
	return tsk
}

// createReadyTaskWithModel mirrors createReadyTask but pins the
// operator-intent CursorModel on the task row. Used by Phase 1a-ii
// tests that exercise the buildCycleMeta wiring.
func (h *harness) createReadyTaskWithModel(ctx context.Context, title, model string) *taskcoredomain.Task {
	h.t.Helper()
	wb := h.gitBinding()
	tsk, err := h.store.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         title,
		InitialPrompt: "do the thing",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		CursorModel:   model,
		WorktreeID:    wb,
	}, taskcoredomain.ActorUser)
	if err != nil {
		h.t.Fatalf("create task: %v", err)
	}
	return tsk
}

func (h *harness) startWorker(ctx context.Context, r runner.Runner, opts worker.Options) (*worker.Worker, <-chan error) {
	h.t.Helper()
	if opts.Notifier == nil {
		opts.Notifier = h.notifier
	}
	w := worker.NewWorker(h.store, h.queue, r, opts)
	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()
	return w, done
}

func (h *harness) waitTaskStatus(ctx context.Context, taskID string, want taskcoredomain.Status) *taskcoredomain.Task {
	h.t.Helper()
	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		got, err := h.store.Get(ctx, taskID)
		if err == nil && got.Status == want {
			return got
		}
		time.Sleep(pollInterval)
	}
	got, _ := h.store.Get(ctx, taskID)
	gotStatus := taskcoredomain.Status("")
	if got != nil {
		gotStatus = got.Status
	}
	h.t.Fatalf("timeout waiting for task %s status=%q (last=%q)", taskID, want, gotStatus)
	return nil
}

func (h *harness) waitNoCycle(ctx context.Context, taskID string) {
	h.t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		cycles, err := h.store.ListCyclesForTask(ctx, taskID, 10)
		if err == nil && len(cycles) == 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if len(cycles) > 0 {
			h.t.Fatalf("expected no cycles for %s, got %d", taskID, len(cycles))
		}
		time.Sleep(pollInterval)
	}
}

// --- recording notifier -------------------------------------------------

type publishCall struct {
	TaskID  string
	CycleID string
}

type recordingNotifier struct {
	mu    sync.Mutex
	calls []publishCall
}

func newRecordingNotifier() *recordingNotifier {
	return &recordingNotifier{}
}

func (r *recordingNotifier) PublishCycleChange(taskID, cycleID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, publishCall{TaskID: taskID, CycleID: cycleID})
}

func (r *recordingNotifier) snapshot() []publishCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]publishCall, len(r.calls))
	copy(out, r.calls)
	return out
}

// --- programmable runners ------------------------------------------------

type blockingRunner struct {
	name    string
	version string

	starts chan runner.Request

	// release is closed when the runner should return; result/err are
	// returned together. If panicMsg is non-empty the runner panics
	// after starts is signalled.
	release  chan struct{}
	result   runner.Result
	err      error
	panicMsg string

	// onStart is invoked synchronously after starts is signalled and
	// before the blocking select; tests use it to drive side effects
	// (delete the task, cancel parent ctx, etc).
	onStart func(req runner.Request)

	// honorCtx, when true, returns wrapped runner.ErrTimeout if ctx
	// fires while we are blocked (matches runnerfake semantics).
	honorCtx bool
}

func newBlockingRunner() *blockingRunner {
	return &blockingRunner{
		name:     "block",
		version:  "v0",
		starts:   make(chan runner.Request, 4),
		release:  make(chan struct{}),
		honorCtx: true,
	}
}

func (b *blockingRunner) Name() string    { return b.name }
func (b *blockingRunner) Version() string { return b.version }

// EffectiveModel mirrors runnerfake/cursor: trim req.CursorModel; empty
// stays empty (the worker tests don't pin a default for the blocking
// runner, so empty here truthfully maps to "no model configured").
func (b *blockingRunner) EffectiveModel(req runner.Request) string {
	return strings.TrimSpace(req.CursorModel)
}

func (b *blockingRunner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	b.starts <- req
	if b.onStart != nil {
		b.onStart(req)
	}
	if b.panicMsg != "" {
		panic(b.panicMsg)
	}
	select {
	case <-b.release:
		if b.err == nil {
			if seedErr := runnerfake.SeedCommitRegisterForTests(req); seedErr != nil {
				return runner.Result{}, seedErr
			}
		}
		return b.result, b.err
	case <-ctx.Done():
		if b.honorCtx {
			return runner.Result{}, fmt.Errorf("blocking runner cancelled: %w", runner.ErrTimeout)
		}
		return b.result, b.err
	}
}

// --- helper assertions --------------------------------------------------

func assertCycleStatus(t *testing.T, st *composition.API, taskID string, wantCount int, wantStatus cyclesdomain.CycleStatus) *cyclesdomain.TaskCycle {
	t.Helper()
	cycles, err := st.ListCyclesForTask(context.Background(), taskID, 10)
	if err != nil {
		t.Fatalf("list cycles: %v", err)
	}
	if len(cycles) != wantCount {
		t.Fatalf("cycle count = %d, want %d", len(cycles), wantCount)
	}
	if wantCount == 0 {
		return nil
	}
	c := cycles[0]
	if c.Status != wantStatus {
		t.Fatalf("cycle status = %q, want %q", c.Status, wantStatus)
	}
	return &c
}

func eventTypeCounts(events []taskeventsdomain.TaskEvent) map[taskeventsdomain.EventType]int {
	out := map[taskeventsdomain.EventType]int{}
	for _, e := range events {
		out[e.Type]++
	}
	return out
}

// --- test cases ----------------------------------------------------------
