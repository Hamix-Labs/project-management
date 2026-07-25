package worker_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// TestIntegration_ReadyWorkerCycleEventsStream is the pkgs-level
// agent→task→cycle seam (B-29): store creates a ready task, the worker
// picks it up via runnerfake (with ScriptProgress), and durable cycle /
// phase / event / stream artifacts are asserted together.
func TestIntegration_ReadyWorkerCycleEventsStream(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "integration-seam")

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "integration ok",
		json.RawMessage(`{"ok":true}`), "",
	))
	r.ScriptProgress(tsk.ID, cyclesdomain.PhaseExecute, runner.ProgressEvent{
		Kind:    "tool_call",
		Subtype: "started",
		Tool:    "ReadFile",
		Message: "Started ReadFile",
		Payload: json.RawMessage(`{"type":"tool_call","name":"ReadFile"}`),
	})

	_, done := h.startWorker(ctx, r, worker.Options{})
	final := h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}
	if final.Status != taskcoredomain.StatusReview {
		t.Fatalf("task status = %q, want done", final.Status)
	}

	bg := context.Background()
	cycle := assertCycleStatus(t, h.store, tsk.ID, 1, cyclesdomain.CycleStatusSucceeded)

	phases, err := h.store.ListPhasesForCycle(bg, cycle.ID)
	if err != nil {
		t.Fatalf("list phases: %v", err)
	}
	if len(phases) != 1 || phases[0].Phase != cyclesdomain.PhaseExecute || phases[0].Status != cyclesdomain.PhaseStatusSucceeded {
		t.Fatalf("phases = %+v, want one execute/succeeded", phases)
	}

	events, err := h.store.ListTaskEvents(bg, tsk.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	counts := eventTypeCounts(events)
	for _, typ := range []taskeventsdomain.EventType{
		taskeventsdomain.EventCycleStarted,
		taskeventsdomain.EventCycleCompleted,
		taskeventsdomain.EventPhaseStarted,
		taskeventsdomain.EventPhaseCompleted,
	} {
		if counts[typ] != 1 {
			t.Fatalf("%s count = %d, want 1 (counts=%v)", typ, counts[typ], counts)
		}
	}

	stream, err := h.store.ListCycleStreamEvents(bg, cycle.ID, 0, 50)
	if err != nil {
		t.Fatalf("list stream: %v", err)
	}
	if len(stream) < 2 {
		t.Fatalf("stream events = %d, want at least setup + tool_call (%+v)", len(stream), stream)
	}
	foundSetup := false
	foundTool := false
	for _, ev := range stream {
		if ev.Kind == runner.ProgressRunStateKind && ev.Tool == runner.ProgressToolHarnessSetup {
			foundSetup = true
		}
		if ev.Kind == "tool_call" && ev.Tool == "ReadFile" {
			foundTool = true
		}
	}
	if !foundSetup {
		t.Fatalf("missing harness setup stream event among %+v", stream)
	}
	if !foundTool {
		t.Fatalf("missing tool_call/ReadFile among %+v", stream)
	}

	if calls := r.Calls(); len(calls) != 1 || calls[0].AttemptSeq != 1 {
		t.Fatalf("runner calls = %#v, want one AttemptSeq=1", calls)
	}
}
