// Package notify holds the thread-safe ReadyTaskNotifier registration
// shared by task CRUD paths that may transition a task into the ready state.
package notify

import (
	"context"
	"log/slog"
	"sync"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// Notifier is invoked after a task row is committed in the ready state
// (on create, update, or dev row mirror). Implementations should not
// block the calling store goroutine for long; the recommended pattern
// is to publish onto a buffered channel and return.
type Notifier interface {
	NotifyReadyTask(ctx context.Context, task taskcoredomain.Task) error
}

// Holder owns the optional notifier registration with sync.RWMutex
// semantics so reads stay cheap on the hot store path. The zero value
// is ready to use; pass nil to Set to clear.
type Holder struct {
	mu sync.RWMutex
	n  Notifier
}

// Set installs n as the active notifier. nil clears the registration.
func (h *Holder) Set(n Notifier) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.storehooks.notify.Set", "enabled", n != nil)
	if h == nil {
		return
	}
	h.mu.Lock()
	h.n = n
	h.mu.Unlock()
}

// Notify invokes the registered notifier for task. It is a best-effort
// fire-and-forget: a missing registration, an empty task id, or a nil
// holder is a no-op.
func (h *Holder) Notify(ctx context.Context, task taskcoredomain.Task) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.storehooks.notify.Notify", "task_id", task.ID)
	if h == nil || task.ID == "" {
		return
	}
	h.mu.RLock()
	n := h.n
	h.mu.RUnlock()
	if n == nil {
		return
	}
	notifyCtx := context.Background()
	if ctx != nil {
		notifyCtx = context.WithoutCancel(ctx)
	}
	if err := n.NotifyReadyTask(notifyCtx, task); err != nil {
		slog.Warn("ready task notifier failed", "cmd", calltrace.LogCmd, "operation", "taskapi.storehooks.notify.Notify",
			"task_id", task.ID, "err", err)
	}
}
