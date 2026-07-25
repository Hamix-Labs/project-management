package model

import (
	"reflect"
	"testing"
	"time"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestTaskCycleCriteriaReport_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := cyclesdomain.TaskCycleCriteriaReport{
		ID:          "r1",
		CycleID:     "cyc-1",
		AttemptSeq:  cyclesdomain.ExecuteCriteriaReportAttemptSeq,
		CriterionID: "crit-1",
		ClaimedDone: true,
		Evidence:    "done",
		WrittenAt:   now,
	}
	m := FromDomainTaskCycleCriteriaReport(orig)
	back := ToDomainTaskCycleCriteriaReport(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestTaskCycleVerifyReport_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := cyclesdomain.TaskCycleVerifyReport{
		ID:           "v1",
		CycleID:      "cyc-1",
		AttemptSeq:   1,
		CriterionID:  "crit-1",
		Verified:     true,
		VerifierKind: checklistdomain.VerifierExecuteAgent,
		Reasoning:    "looks good",
		WrittenAt:    now,
	}
	m := FromDomainTaskCycleVerifyReport(orig)
	back := ToDomainTaskCycleVerifyReport(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestTaskCycleCommandRun_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := cyclesdomain.TaskCycleCommandRun{
		ID:          "cmd-1",
		CycleID:     "cyc-1",
		AttemptSeq:  1,
		CriterionID: "crit-1",
		CommandSeq:  0,
		ExitCode:    0,
		MetaPath:    "/tmp/out.meta",
		WrittenAt:   now,
	}
	m := FromDomainTaskCycleCommandRun(orig)
	back := ToDomainTaskCycleCommandRun(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestTaskCycleCommit_roundTrip(t *testing.T) {
	t.Parallel()
	when := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	recorded := when.Add(time.Minute)
	orig := cyclesdomain.TaskCycleCommit{
		ID:          "c1",
		TaskID:      "task-1",
		CycleID:     "cyc-1",
		PhaseSeq:    1,
		Seq:         1,
		Repo:        "hamix",
		Worktree:    "/wt",
		Branch:      "main",
		SHA:         "abc123",
		CommittedAt: when,
		Message:     "fix",
		RecordedAt:  recorded,
	}
	m := FromDomainTaskCycleCommit(orig)
	back := ToDomainTaskCycleCommit(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}
