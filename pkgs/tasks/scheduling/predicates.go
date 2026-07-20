package scheduling

import (
	"time"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// EvaluateWorkerReadiness applies the four worker predicates in fixed order.
// dependenciesMet must reflect store-loaded edge satisfaction when predicate 4 applies.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EvaluateWorkerReadiness(task *taskcoredomain.Task, now time.Time, dependenciesMet bool) ReadinessResult {
	return taskcorecontract.EvaluateWorkerReadiness(task, now, dependenciesMet)
}
