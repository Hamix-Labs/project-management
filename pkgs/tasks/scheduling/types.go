package scheduling

import (
	"time"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
)

// FailedPredicate identifies the first worker readiness check that failed.
// String values are stable for logs and metrics.
type FailedPredicate = taskcorecontract.FailedPredicate

const (
	FailedPredicateNone         FailedPredicate = taskcorecontract.FailedPredicateNone
	FailedPredicateStatus       FailedPredicate = taskcorecontract.FailedPredicateStatus
	FailedPredicatePickup       FailedPredicate = taskcorecontract.FailedPredicatePickup
	FailedPredicateGate         FailedPredicate = taskcorecontract.FailedPredicateGate
	FailedPredicateDependencies FailedPredicate = taskcorecontract.FailedPredicateDependencies
)

// ReadinessResult is the outcome of EvaluateWorkerReadiness.
type ReadinessResult = taskcorecontract.ReadinessResult

// NotifyDecision is the post-commit notify/wake action for a ready transition.
type NotifyDecision struct {
	NotifyQueue  bool
	ScheduleWake *time.Time
	CancelWake   bool
}
