package scheduling

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// EdgeSatisfied reports whether predecessor meets the edge predicate.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EdgeSatisfied(predecessor *taskcoredomain.Task, satisfies taskcoredomain.DependencySatisfies) bool {
	if predecessor == nil {
		return false
	}
	_ = satisfies
	return predecessor.Status == taskcoredomain.StatusDone
}
