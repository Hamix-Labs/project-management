package storehooks

import (
	"context"
	"sync"
	"time"
)

// PickupWake schedules in-process wakeups for ready tasks whose
// pickup_not_before is in the future.
type PickupWake interface {
	Schedule(ctx context.Context, taskID string, notBefore time.Time)
	Cancel(taskID string)
	Stop()
}

// PickupWakeRegistry holds an optional PickupWake hook for task CRUD paths.
type PickupWakeRegistry struct {
	mu   sync.RWMutex
	wake PickupWake
}

// Set registers w for deferred-pickup scheduling (nil clears).
//
//funclogmeasure:skip category=hot-path reason="In-process hook registry; scheduling traces at BC chokepoints."
func (r *PickupWakeRegistry) Set(w PickupWake) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.wake = w
	r.mu.Unlock()
}

// SchedulePickupWake registers (or replaces) a wake at notBefore UTC for taskID.
//
//funclogmeasure:skip category=hot-path reason="In-process hook forwarder; worker pickup traces at queue chokepoints."
func (r *PickupWakeRegistry) SchedulePickupWake(ctx context.Context, taskID string, notBefore time.Time) {
	r.mu.RLock()
	w := r.wake
	r.mu.RUnlock()
	if w == nil || taskID == "" {
		return
	}
	w.Schedule(ctx, taskID, notBefore)
}

// CancelPickupWake removes any pending wake for taskID.
//
//funclogmeasure:skip category=hot-path reason="In-process hook forwarder; worker pickup traces at queue chokepoints."
func (r *PickupWakeRegistry) CancelPickupWake(taskID string) {
	r.mu.RLock()
	w := r.wake
	r.mu.RUnlock()
	if w == nil || taskID == "" {
		return
	}
	w.Cancel(taskID)
}
