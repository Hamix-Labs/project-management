package agentworker

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

const taskUpdatedPublishTimeout = 5 * time.Second

type taskGetter interface {
	Get(ctx context.Context, id string) (*domain.Task, error)
}

type cycleChangeSSEAdapter struct {
	pub realtime.Publisher
}

func newCycleChangeSSEAdapter(pub realtime.Publisher) *cycleChangeSSEAdapter {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.newCycleChangeSSEAdapter")
	return &cycleChangeSSEAdapter{pub: pub}
}

func (a *cycleChangeSSEAdapter) PublishCycleChange(taskID, cycleID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.cycleChangeSSEAdapter.PublishCycleChange",
		"task_id", taskID, "cycle_id", cycleID)
	if a == nil || a.pub == nil || taskID == "" {
		return
	}
	a.pub.Publish(realtime.Event{
		Type:    realtime.TaskCycleChanged,
		ID:      taskID,
		CycleID: cycleID,
	})
}

type taskUpdatedSSEAdapter struct {
	pub   realtime.Publisher
	store taskGetter
}

func newTaskUpdatedSSEAdapter(pub realtime.Publisher, store taskGetter) *taskUpdatedSSEAdapter {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.newTaskUpdatedSSEAdapter")
	return &taskUpdatedSSEAdapter{pub: pub, store: store}
}

func (a *taskUpdatedSSEAdapter) PublishTaskUpdated(taskID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.taskUpdatedSSEAdapter.PublishTaskUpdated",
		"task_id", taskID)
	if a == nil || a.pub == nil || a.store == nil || taskID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), taskUpdatedPublishTimeout)
	defer cancel()
	if err := realtime.PublishEnrichedTaskUpdated(ctx, a.pub, func(ctx context.Context, id string) (any, error) {
		return a.store.Get(ctx, id)
	}, taskID); err != nil {
		slog.Warn("agent worker task_updated publish failed", "cmd", calltrace.LogCmd,
			"operation", "taskapi.taskUpdatedSSEAdapter.PublishTaskUpdated.err",
			"task_id", taskID, "err", err)
	}
}

const (
	agentRunProgressMinInterval     = 750 * time.Millisecond
	agentRunProgressThrottleEntries = 512
)

type runProgressSSEAdapter struct {
	pub         realtime.Publisher
	minInterval time.Duration

	mu       sync.Mutex
	lastSent map[string]time.Time
}

func newRunProgressSSEAdapter(pub realtime.Publisher, minInterval time.Duration) *runProgressSSEAdapter {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.newRunProgressSSEAdapter")
	return &runProgressSSEAdapter{
		pub:         pub,
		minInterval: minInterval,
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
	a.pub.Publish(realtime.Event{
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
	})
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
