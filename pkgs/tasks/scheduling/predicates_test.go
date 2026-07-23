package scheduling

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"testing"
	"time"
)

func TestEvaluateWorkerReadiness_predicateOrder(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)
	gateHeld := &taskcoredomain.TaskGate{Kind: taskcoredomain.GateKindManualApproval, Status: taskcoredomain.GateStatusActive}

	cases := []struct {
		name      string
		task      *taskcoredomain.Task
		depsMet   bool
		wantReady bool
		wantPred  FailedPredicate
	}{
		{"nil task", nil, true, false, FailedPredicateStatus},
		{"not ready", &taskcoredomain.Task{Status: taskcoredomain.StatusBlocked}, true, false, FailedPredicateStatus},
		{"future pickup", &taskcoredomain.Task{Status: taskcoredomain.StatusReady, PickupNotBefore: &future}, true, false, FailedPredicatePickup},
		{"held gate", &taskcoredomain.Task{Status: taskcoredomain.StatusReady, Gate: gateHeld}, true, false, FailedPredicateGate},
		{"open dependency", &taskcoredomain.Task{Status: taskcoredomain.StatusReady}, false, false, FailedPredicateDependencies},
		{"missing worktree", &taskcoredomain.Task{Status: taskcoredomain.StatusReady}, true, false, FailedPredicateWorktree},
		{"all clear", &taskcoredomain.Task{Status: taskcoredomain.StatusReady, PickupNotBefore: &past, WorktreeID: strPtr("wt-1")}, true, true, FailedPredicateNone},
		{"nil pickup", &taskcoredomain.Task{Status: taskcoredomain.StatusReady, WorktreeID: strPtr("wt-1")}, true, true, FailedPredicateNone},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := EvaluateWorkerReadiness(c.task, now, c.depsMet)
			if got.Ready != c.wantReady || got.FailedPredicate != c.wantPred {
				t.Fatalf("got %+v want ready=%v pred=%q", got, c.wantReady, c.wantPred)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
