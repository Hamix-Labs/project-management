package agentworker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

const (
	taskUpdatedPublishTimeout = 5 * time.Second
	taskUpdatedQueueDepth     = 32
)

type taskGetter interface {
	Get(ctx context.Context, id string) (*taskcoredomain.Task, error)
}

type cycleChangeSSEAdapter struct {
	pub     realtime.Publisher
	metrics NotifierMetrics
}

func newCycleChangeSSEAdapter(pub realtime.Publisher, metrics NotifierMetrics) *cycleChangeSSEAdapter {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.newCycleChangeSSEAdapter")
	return &cycleChangeSSEAdapter{pub: pub, metrics: metrics}
}

func (a *cycleChangeSSEAdapter) PublishCycleChange(taskID, cycleID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.cycleChangeSSEAdapter.PublishCycleChange",
		"task_id", taskID, "cycle_id", cycleID)
	if a == nil || a.pub == nil || taskID == "" {
		return
	}
	ev := realtime.Event{
		Type:    realtime.TaskCycleChanged,
		ID:      taskID,
		CycleID: cycleID,
	}
	go publishEventNonBlocking(a.pub, a.metrics, "cycle_change", ev)
}

type taskUpdatedSSEAdapter struct {
	pub     realtime.Publisher
	store   taskGetter
	metrics NotifierMetrics

	startOnce sync.Once
	jobs      chan string
}

func newTaskUpdatedSSEAdapter(pub realtime.Publisher, store taskGetter, metrics NotifierMetrics) *taskUpdatedSSEAdapter {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.newTaskUpdatedSSEAdapter")
	a := &taskUpdatedSSEAdapter{
		pub:     pub,
		store:   store,
		metrics: metrics,
		jobs:    make(chan string, taskUpdatedQueueDepth),
	}
	a.startOnce.Do(func() { go a.worker() })
	return a
}

func (a *taskUpdatedSSEAdapter) worker() {
	for taskID := range a.jobs {
		ctx, cancel := context.WithTimeout(context.Background(), taskUpdatedPublishTimeout)
		if err := realtime.PublishEnrichedTaskUpdated(ctx, a.pub, func(ctx context.Context, id string) (any, error) {
			return a.store.Get(ctx, id)
		}, taskID); err != nil {
			slog.Warn("agent worker task_updated publish failed", "cmd", calltrace.LogCmd,
				"operation", "taskapi.taskUpdatedSSEAdapter.worker.err",
				"task_id", taskID, "err", err)
		}
		cancel()
	}
}

func (a *taskUpdatedSSEAdapter) PublishTaskUpdated(taskID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.taskUpdatedSSEAdapter.PublishTaskUpdated",
		"task_id", taskID)
	if a == nil || a.pub == nil || a.store == nil || taskID == "" {
		return
	}
	select {
	case a.jobs <- taskID:
	default:
		recordNotifierDropped(a.metrics, "task_updated")
		slog.Warn("agent worker task_updated enqueue dropped",
			"cmd", calltrace.LogCmd, "operation", "taskapi.taskUpdatedSSEAdapter.drop",
			"task_id", taskID)
	}
}

const (
	agentRunProgressMinInterval     = 750 * time.Millisecond
	agentRunProgressThrottleEntries = 512
)

type runProgressSSEAdapter struct {
	pub         realtime.Publisher
	minInterval time.Duration
	metrics     NotifierMetrics

	mu       sync.Mutex
	lastSent map[string]time.Time
}

func newRunProgressSSEAdapter(pub realtime.Publisher, minInterval time.Duration, metrics NotifierMetrics) *runProgressSSEAdapter {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.newRunProgressSSEAdapter")
	return &runProgressSSEAdapter{
		pub:         pub,
		minInterval: minInterval,
		metrics:     metrics,
		lastSent:    make(map[string]time.Time),
	}
}

func (a *runProgressSSEAdapter) PublishRunProgress(taskID, cycleID string, phaseSeq int64, runCorrelationID string, ev runner.ProgressEvent) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.runProgressSSEAdapter.PublishRunProgress",
		"task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq,
		"run_correlation_id", runCorrelationID,
		"kind", ev.Kind, "subtype", ev.Subtype)
	if a == nil || a.pub == nil || taskID == "" || cycleID == "" || phaseSeq <= 0 || ev.Kind == "" {
		return
	}
	if a.shouldDrop(taskID, cycleID, phaseSeq) {
		return
	}
	event := realtime.Event{
		Type:             realtime.AgentRunProgress,
		ID:               taskID,
		CycleID:          cycleID,
		PhaseSeq:         phaseSeq,
		RunCorrelationID: runCorrelationID,
		Progress: &realtime.RunProgressPayload{
			Kind:    ev.Kind,
			Subtype: ev.Subtype,
			Message: ev.Message,
			Tool:    ev.Tool,
		},
	}
	go publishEventNonBlocking(a.pub, a.metrics, "run_progress", event)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (a *runProgressSSEAdapter) shouldDrop(taskID, cycleID string, phaseSeq int64) bool {
	if a.minInterval <= 0 {
		return false
	}
	key := fmt.Sprintf("%s:%s:%d", taskID, cycleID, phaseSeq)
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	last, ok := a.lastSent[key]
	if ok && now.Sub(last) < a.minInterval {
		return true
	}
	a.lastSent[key] = now
	if len(a.lastSent) > agentRunProgressThrottleEntries {
		for old := range a.lastSent {
			if old != key {
				delete(a.lastSent, old)
				break
			}
		}
	}
	return false
}

func publishEventNonBlocking(pub realtime.Publisher, metrics NotifierMetrics, kind string, ev realtime.Event) {
	if pub == nil {
		return
	}
	done := make(chan struct{}, 1)
	go func() {
		pub.Publish(ev)
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		recordNotifierDropped(metrics, kind)
		slog.Warn("agent worker notifier publish slow; continuing without blocking harness",
			"cmd", calltrace.LogCmd, "operation", "taskapi.notifier.publish_timeout",
			"kind", kind, "task_id", ev.ID)
	}
}

//funclogmeasure:skip category=hot-path reason="Metrics delegate; drop events trace at notifier publish boundary."
func recordNotifierDropped(metrics NotifierMetrics, kind string) {
	if metrics == nil {
		return
	}
	metrics.RecordNotifierDropped(kind)
}
