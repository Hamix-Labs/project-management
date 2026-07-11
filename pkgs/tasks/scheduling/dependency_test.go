package scheduling

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"testing"
)

func TestEdgeSatisfied_doneOnly(t *testing.T) {
	t.Parallel()
	done := &taskcoredomain.Task{Status: taskcoredomain.StatusDone}
	ready := &taskcoredomain.Task{Status: taskcoredomain.StatusReady}
	if !EdgeSatisfied(done, taskcoredomain.DependencySatisfiesDone) {
		t.Fatal("done predecessor should satisfy")
	}
	if EdgeSatisfied(ready, taskcoredomain.DependencySatisfiesDone) {
		t.Fatal("ready predecessor should not satisfy")
	}
}
