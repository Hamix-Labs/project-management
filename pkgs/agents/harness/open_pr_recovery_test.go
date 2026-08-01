package harness

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestBuildRecoveryContext_openPR(t *testing.T) {
	t.Parallel()
	h := New(nil, runnerfake.New(), Options{ReportDir: t.TempDir()})
	cycle := &cyclesdomain.TaskCycle{ID: "cyc-open", AttemptSeq: 3, MetaJSON: []byte(`{"run_kind":"open_pr"}`)}
	state := &processState{
		verify: verifyLifecycleState{
			lockedPasses: map[string]criterionVerdict{"a": {ID: "a", Passed: true}},
			verifySnap: verificationSnapshot{
				Criteria: []checklistcontract.ChecklistVerifyItem{{ID: "a"}, {ID: "b"}},
			},
		},
	}
	known := []cyclesdomain.TaskCycleCommit{{SHA: "abc", Message: "m"}}
	ctx := h.buildRecoveryContext(
		cyclesdomain.PhaseExecute,
		&taskcoredomain.Task{ID: "t1"},
		cycle,
		state,
		cycleLoopOpts{resumeNotice: true, knownCommits: known},
		taskcoredomain.RetryResume,
	)
	if ctx.Kind != prompt.RecoveryHumanOpenPR {
		t.Fatalf("kind=%q want %q", ctx.Kind, prompt.RecoveryHumanOpenPR)
	}
	if ctx.CycleID != cycle.ID || ctx.AttemptSeq != 3 {
		t.Fatalf("cycle fields: %+v", ctx)
	}
	if len(ctx.OpenPRKnownCommits) != 1 || ctx.OpenPRKnownCommits[0].SHA != "abc" {
		t.Fatalf("known commits: %+v", ctx.OpenPRKnownCommits)
	}
	if len(ctx.LockedCriteria) != 1 || ctx.LockedCriteria[0] != "a" {
		t.Fatalf("locked: %+v", ctx.LockedCriteria)
	}
	if len(ctx.ExpectedIDs) != 1 || ctx.ExpectedIDs[0] != "b" {
		t.Fatalf("expected active ids: %+v", ctx.ExpectedIDs)
	}
}
