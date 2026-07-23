package contract

import (
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// ReadinessResult is the outcome of EvaluateWorkerReadiness.
type ReadinessResult struct {
	Ready           bool
	FailedPredicate FailedPredicate
}

// ShouldNotifyReadyNow returns true when a freshly-ready task may enter the
// in-memory queue immediately. Only pickup_not_before is checked — see ADR-0023 I7.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ShouldNotifyReadyNow(pickupNotBefore *time.Time, now time.Time) bool {
	if pickupNotBefore == nil {
		return true
	}
	return !pickupNotBefore.After(now)
}

// EvaluateWorkerReadiness applies the worker predicates in fixed order.
// dependenciesMet must reflect store-loaded edge satisfaction when the
// dependencies predicate applies.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EvaluateWorkerReadiness(task *domain.Task, now time.Time, dependenciesMet bool) ReadinessResult {
	if task == nil || task.Status != domain.StatusReady {
		return ReadinessResult{Ready: false, FailedPredicate: FailedPredicateStatus}
	}
	if task.PickupNotBefore != nil && task.PickupNotBefore.After(now) {
		return ReadinessResult{Ready: false, FailedPredicate: FailedPredicatePickup}
	}
	if task.Gate != nil && task.Gate.GateBlocksWorker() {
		return ReadinessResult{Ready: false, FailedPredicate: FailedPredicateGate}
	}
	if !dependenciesMet {
		return ReadinessResult{Ready: false, FailedPredicate: FailedPredicateDependencies}
	}
	if task.WorktreeID == nil || strings.TrimSpace(*task.WorktreeID) == "" {
		return ReadinessResult{Ready: false, FailedPredicate: FailedPredicateWorktree}
	}
	return ReadinessResult{Ready: true, FailedPredicate: FailedPredicateNone}
}

// EdgeSatisfied reports whether predecessor meets the edge predicate.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EdgeSatisfied(predecessor *domain.Task, satisfies domain.DependencySatisfies) bool {
	if predecessor == nil {
		return false
	}
	_ = satisfies
	return predecessor.Status == domain.StatusDone
}
