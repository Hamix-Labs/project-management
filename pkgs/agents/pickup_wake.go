package agents

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"container/heap"
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

// pickupWakeQueueFullBackoff is how long to wait before retrying a wake
// that failed because the ready-task queue was full. Short so deferred
// ready tasks recover once capacity frees, without busy-spinning.
const pickupWakeQueueFullBackoff = time.Second

// PickupWakeScheduler implements taskcorestore.PickupWake: a min-heap of
// (pickup_not_before, task_id) with one timer for the earliest deadline.
// On fire it loads the task and enqueues when ShouldNotifyReadyNow holds.
type PickupWakeScheduler struct {
	st worker.Store
	q  *MemoryQueue

	mu      sync.Mutex
	byID    map[string]*wakeItem
	heap    wakeHeap
	timer   *time.Timer
	stopped bool
}

var _ taskcorestore.PickupWake = (*PickupWakeScheduler)(nil)

type wakeItem struct {
	taskID    string
	notBefore time.Time
	index     int
}

type wakeHeap []*wakeItem

func (h wakeHeap) Len() int { return len(h) }

func (h wakeHeap) Less(i, j int) bool {
	a, b := h[i].notBefore, h[j].notBefore
	if !a.Equal(b) {
		return a.Before(b)
	}
	return h[i].taskID < h[j].taskID
}

func (h wakeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *wakeHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*wakeItem)
	item.index = n
	*h = append(*h, item)
}

func (h *wakeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// NewPickupWakeScheduler returns a scheduler backed by st and q. The
// caller must register it with (*composition.API).SetPickupWake and call
// Hydrate once at startup.
func NewPickupWakeScheduler(st worker.Store, q *MemoryQueue) *PickupWakeScheduler {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agents.NewPickupWakeScheduler")
	return &PickupWakeScheduler{
		st:   st,
		q:    q,
		byID: make(map[string]*wakeItem),
	}
}

// Hydrate schedules wake timers for every ready task with pickup_not_before
// in the future (bounded list). Safe to call once after SetPickupWake.
func (w *PickupWakeScheduler) Hydrate(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agents.PickupWakeScheduler.Hydrate")
	if w == nil || w.st == nil {
		return nil
	}
	rows, err := w.st.ListDeferredReadyPickupTasks(ctx, 10_000)
	if err != nil {
		return err
	}
	for i := range rows {
		r := rows[i]
		w.Schedule(ctx, r.ID, r.PickupNotBefore)
	}
	return nil
}

// Schedule implements taskcorestore.PickupWake.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *PickupWakeScheduler) Schedule(ctx context.Context, taskID string, notBefore time.Time) {
	if w == nil || taskID == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return
	}
	if it, ok := w.byID[taskID]; ok {
		heap.Remove(&w.heap, it.index)
		delete(w.byID, taskID)
	}
	nb := notBefore.UTC()
	item := &wakeItem{taskID: taskID, notBefore: nb}
	heap.Push(&w.heap, item)
	w.byID[taskID] = item
	w.resetTimerLocked()
}

// Cancel implements taskcorestore.PickupWake.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *PickupWakeScheduler) Cancel(taskID string) {
	if w == nil || taskID == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return
	}
	it, ok := w.byID[taskID]
	if !ok {
		return
	}
	heap.Remove(&w.heap, it.index)
	delete(w.byID, taskID)
	w.resetTimerLocked()
}

// Stop implements taskcorestore.PickupWake.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *PickupWakeScheduler) Stop() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stopped = true
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	w.byID = nil
	w.heap = nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *PickupWakeScheduler) resetTimerLocked() {
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	if len(w.heap) == 0 {
		return
	}
	next := w.heap[0]
	d := time.Until(next.notBefore)
	if d < 0 {
		d = 0
	}
	w.timer = time.AfterFunc(d, w.fire)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *PickupWakeScheduler) fire() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	now := time.Now().UTC()
	for len(w.heap) > 0 {
		peek := w.heap[0]
		if peek.notBefore.After(now) {
			break
		}
		item := heap.Pop(&w.heap).(*wakeItem)
		delete(w.byID, item.taskID)
		tid := item.taskID
		w.mu.Unlock()
		w.tryNotify(tid, now)
		w.mu.Lock()
		if w.stopped {
			w.mu.Unlock()
			return
		}
		now = time.Now().UTC()
	}
	w.resetTimerLocked()
	w.mu.Unlock()
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (w *PickupWakeScheduler) tryNotify(taskID string, now time.Time) {
	if w.st == nil || w.q == nil {
		return
	}
	ctx := context.Background()
	t, err := w.st.Get(ctx, taskID)
	if err != nil {
		slog.Warn("pickup wake Get failed", "cmd", calltrace.LogCmd,
			"operation", "agents.PickupWakeScheduler.tryNotify.get_err",
			"task_id", taskID, "err", err)
		return
	}
	if t == nil || t.Status != taskcoredomain.StatusReady {
		return
	}
	if !taskcorestore.ShouldNotifyReadyNow(t.PickupNotBefore, now) {
		return
	}
	if err := w.q.NotifyReadyTask(ctx, *t); err != nil {
		slog.Warn("pickup wake NotifyReadyTask failed", "cmd", calltrace.LogCmd,
			"operation", "agents.PickupWakeScheduler.tryNotify.notify_err",
			"task_id", taskID, "err", err)
		if errors.Is(err, ErrQueueFull) {
			w.Schedule(ctx, taskID, now.Add(pickupWakeQueueFullBackoff))
		}
	}
}
