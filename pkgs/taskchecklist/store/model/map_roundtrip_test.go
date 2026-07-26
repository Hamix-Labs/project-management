package model

import (
	"reflect"
	"testing"
	"time"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

func TestTaskChecklistItem_roundTrip(t *testing.T) {
	t.Parallel()
	orig := checklistdomain.TaskChecklistItem{
		ID: "item-1", TaskID: "task-1", SortOrder: 1, Text: "criterion",
	}
	m := FromDomainTaskChecklistItem(orig)
	back := ToDomainTaskChecklistItem(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestTaskChecklistItemCommand_roundTrip(t *testing.T) {
	t.Parallel()
	timeout := 300
	orig := checklistdomain.TaskChecklistItemCommand{
		ID: "cmd-1", ItemID: "item-1", SortOrder: 0,
		Command: "go test ./...", ExpectedOutcome: "pass",
		TimeoutSeconds: &timeout,
	}
	m := FromDomainTaskChecklistItemCommand(orig)
	back := ToDomainTaskChecklistItemCommand(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
	if back.TimeoutSeconds == &timeout {
		t.Fatal("TimeoutSeconds must be cloned, not aliased")
	}
}

func TestTaskChecklistCompletion_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := checklistdomain.TaskChecklistCompletion{
		TaskID: "task-1", ItemID: "item-1", At: now,
		By: "agent", Evidence: "ok",
		VerifiedBy:        checklistdomain.VerifierExecuteAgent,
		VerifierReasoning: "tests pass", CycleID: "cyc-1",
	}
	m := FromDomainTaskChecklistCompletion(orig)
	back := ToDomainTaskChecklistCompletion(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}
