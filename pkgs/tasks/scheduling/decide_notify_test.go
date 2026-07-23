package scheduling

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"testing"
	"time"
)

func TestDecideNotifyAfterReadyTransition(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Minute)
	wt := "wt-1"

	readyTask := &taskcoredomain.Task{ID: "t1", Status: taskcoredomain.StatusReady, WorktreeID: &wt}
	noWorktree := &taskcoredomain.Task{ID: "t0", Status: taskcoredomain.StatusReady}
	futurePickup := &taskcoredomain.Task{ID: "t2", Status: taskcoredomain.StatusReady, PickupNotBefore: &future, WorktreeID: &wt}
	eligiblePickup := &taskcoredomain.Task{ID: "t3", Status: taskcoredomain.StatusReady, PickupNotBefore: &past, WorktreeID: &wt}

	cases := []struct {
		name          string
		prev          taskcoredomain.Status
		task          *taskcoredomain.Task
		pickupTouched bool
		wantNotify    bool
		wantWake      bool
		wantCancel    bool
	}{
		{"create ready with worktree", "", readyTask, false, true, false, true},
		{"create ready without worktree", "", noWorktree, false, false, false, false},
		{"transition to ready", taskcoredomain.StatusBlocked, readyTask, false, true, false, true},
		{"stay ready no pickup touch", taskcoredomain.StatusReady, readyTask, false, false, false, true},
		{"pickup touched eligible", taskcoredomain.StatusReady, eligiblePickup, true, true, false, true},
		{"future pickup schedules wake", taskcoredomain.StatusBlocked, futurePickup, false, false, true, false},
		{"not ready cancels wake", taskcoredomain.StatusBlocked, &taskcoredomain.Task{Status: taskcoredomain.StatusBlocked}, false, false, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := DecideNotifyAfterReadyTransition(c.prev, c.task, c.pickupTouched, now)
			if got.NotifyQueue != c.wantNotify {
				t.Fatalf("NotifyQueue=%v want %v", got.NotifyQueue, c.wantNotify)
			}
			if (got.ScheduleWake != nil) != c.wantWake {
				t.Fatalf("ScheduleWake set=%v want %v", got.ScheduleWake != nil, c.wantWake)
			}
			if got.CancelWake != c.wantCancel {
				t.Fatalf("CancelWake=%v want %v", got.CancelWake, c.wantCancel)
			}
		})
	}
}
