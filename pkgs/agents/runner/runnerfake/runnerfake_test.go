package runnerfake

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestRunner_autoInjectsSessionID(t *testing.T) {
	t.Parallel()
	r := New()
	r.Script("task-1", cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok", []byte(`{}`), ""))
	res, err := r.Run(context.Background(), runner.Request{
		TaskID:     "task-1",
		Phase:      cyclesdomain.PhaseExecute,
		AttemptSeq: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := cyclesdomain.SessionIDFromDetailsJSON(res.Details)
	if id == "" {
		t.Fatalf("expected auto session_id, details=%s", res.Details)
	}
}

func TestRunner_WithoutAutoSessionID(t *testing.T) {
	t.Parallel()
	r := New().WithoutAutoSessionID()
	r.Script("task-1", cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok", []byte(`{}`), ""))
	res, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-1",
		Phase:  cyclesdomain.PhaseExecute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id := cyclesdomain.SessionIDFromDetailsJSON(res.Details); id != "" {
		t.Fatalf("expected no session_id, got %q", id)
	}
}

func TestRunner_preservesExistingSessionID(t *testing.T) {
	t.Parallel()
	r := New()
	details, err := json.Marshal(map[string]string{"session_id": "keep-me"})
	if err != nil {
		t.Fatal(err)
	}
	r.Script("task-1", cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok", details, ""))
	res, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-1",
		Phase:  cyclesdomain.PhaseExecute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := cyclesdomain.SessionIDFromDetailsJSON(res.Details); got != "keep-me" {
		t.Fatalf("session_id = %q, want keep-me", got)
	}
}
