package model

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestTask_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	pid := "proj-1"
	wb := "wb-1"
	cm := "opus"
	gate := &domain.TaskGate{Status: domain.GateStatusPendingRelease, Hold: true}
	retry := &domain.PendingRetry{Mode: domain.RetryResume, ParentCycleID: "cyc-1"}
	orig := domain.Task{
		ID:              "task-1",
		Title:           "Ship it",
		Status:          domain.StatusReady,
		Priority:        domain.PriorityHigh,
		InitialPrompt:   "do the thing",
		ProjectID:       &pid,
		Number:          intPtr(7),
		WorktreeID:      &wb,
		CursorModel:     cm,
		PickupNotBefore: &now,
		PendingRetry:    retry,
		Gate:            gate,
		Tags:            []string{"a", "b"},
		Milestone:       strPtr("m1"),
		RunnerConfig:    json.RawMessage("{}"),
	}
	m := FromDomainTask(orig)
	back := ToDomainTask(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch:\norig=%+v\nback=%+v", orig, back)
	}
}

func TestFromDomainTask_emptyTagsDriverValue(t *testing.T) {
	t.Parallel()
	m := FromDomainTask(domain.Task{
		ID:            "task-empty-tags",
		Title:         "t",
		InitialPrompt: "p",
		Status:        domain.StatusRunning,
		Priority:      domain.PriorityMedium,
		Runner:        "cursor",
		RunnerConfig:  json.RawMessage("{}"),
	})
	for _, tc := range []struct {
		name string
		val  interface{ Value() (driver.Value, error) }
	}{
		{name: "tags", val: m.Tags},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v, err := tc.val.Value()
			if err != nil {
				t.Fatal(err)
			}
			s, ok := v.(string)
			if !ok {
				if b, ok := v.([]byte); ok {
					s = string(b)
				} else {
					t.Fatalf("Value type %T = %#v", v, v)
				}
			}
			if s != "[]" {
				t.Fatalf("Value = %q, want []", s)
			}
		})
	}
}

func TestTaskDependency_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	orig := domain.TaskDependency{
		TaskID:          "t1",
		DependsOnTaskID: "t0",
		Satisfies:       domain.DependencySatisfiesDone,
		CreatedAt:       now,
	}
	m := FromDomainTaskDependency(orig)
	back := ToDomainTaskDependency(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func strPtr(s string) *string { return &s }

func intPtr(n int) *int { return &n }
