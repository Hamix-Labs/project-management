package scheduling

import (
	"strings"
	"time"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// DecideNotifyAfterReadyTransition chooses post-commit queue notify and pickup wake
// after Create, Update, or RequestTaskRetry. It encodes I4 and I7: pickup deferral
// vs immediate notify on transition or pickup patch — not full worker readiness.
// Tasks without a worktree binding are never enqueued (ADR-0083); the worktree
// provisioner notifies after allocate.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideNotifyAfterReadyTransition(
	prevStatus taskcoredomain.Status,
	task *taskcoredomain.Task,
	pickupTouched bool,
	now time.Time,
) NotifyDecision {
	if task == nil || task.Status != taskcoredomain.StatusReady {
		if task != nil && task.Status != taskcoredomain.StatusReady {
			return NotifyDecision{CancelWake: true}
		}
		return NotifyDecision{}
	}
	if task.WorktreeID == nil || strings.TrimSpace(*task.WorktreeID) == "" {
		return NotifyDecision{}
	}
	if task.PickupNotBefore != nil && task.PickupNotBefore.After(now) {
		at := task.PickupNotBefore.UTC()
		return NotifyDecision{ScheduleWake: &at}
	}
	transitionedToReady := prevStatus != taskcoredomain.StatusReady
	notify := transitionedToReady || pickupTouched
	if notify && !ShouldNotifyReadyNow(task.PickupNotBefore, now) {
		notify = false
	}
	return NotifyDecision{
		NotifyQueue: notify,
		CancelWake:  true,
	}
}
