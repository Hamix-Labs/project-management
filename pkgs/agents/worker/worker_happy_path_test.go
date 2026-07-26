package worker_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

func TestWorker_HappyPath_writesOnePhaseAndFourMirrors(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsk := h.createReadyTask(ctx, "happy")

	r := runnerfake.New()
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "all green",
		json.RawMessage(`{"ok":true}`), "",
	))

	_, done := h.startWorker(ctx, r, worker.Options{})
	final := h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("worker exit err: %v", err)
	}

	bg := context.Background()
	if final.Status != taskcoredomain.StatusReview {
		t.Fatalf("task status = %q, want done", final.Status)
	}

	cycle := assertCycleStatus(t, h.store, tsk.ID, 1, cyclesdomain.CycleStatusSucceeded)

	var meta map[string]string
	if err := json.Unmarshal(cycle.MetaJSON, &meta); err != nil {
		t.Fatalf("unmarshal meta: %v (raw=%s)", err, cycle.MetaJSON)
	}
	if meta["runner"] != "fake" || meta["runner_version"] != "v0" || len(meta["prompt_hash"]) != 64 {
		t.Fatalf("meta json shape = %+v", meta)
	}
	// cursor_model_intent + cursor_model_effective MUST be present
	// even when empty so the audit trail can distinguish "no model
	// configured anywhere" from "key was never recorded". The harness
	// task has no CursorModel and the runnerfake has no default model,
	// so both should be "".
	if _, ok := meta["cursor_model_intent"]; !ok {
		t.Fatalf("meta missing cursor_model_intent (must be present, even empty): %+v", meta)
	}
	if _, ok := meta["cursor_model_effective"]; !ok {
		t.Fatalf("meta missing cursor_model_effective (must be present, even empty): %+v", meta)
	}
	if meta["cursor_model_intent"] != "" || meta["cursor_model_effective"] != "" {
		t.Fatalf("meta cursor_model_intent/cursor_model_effective expected empty for default-only harness task: %+v", meta)
	}

	phases, err := h.store.ListPhasesForCycle(bg, cycle.ID)
	if err != nil {
		t.Fatalf("list phases: %v", err)
	}
	if len(phases) != 1 {
		t.Fatalf("phase count = %d, want 1", len(phases))
	}
	if phases[0].Phase != cyclesdomain.PhaseExecute || phases[0].Status != cyclesdomain.PhaseStatusSucceeded {
		t.Fatalf("phase[0] = %q/%q, want execute/succeeded", phases[0].Phase, phases[0].Status)
	}

	events, err := h.store.ListTaskEvents(bg, tsk.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	counts := eventTypeCounts(events)
	if counts[taskeventsdomain.EventCycleStarted] != 1 {
		t.Fatalf("cycle_started count = %d, want 1", counts[taskeventsdomain.EventCycleStarted])
	}
	if counts[taskeventsdomain.EventCycleCompleted] != 1 {
		t.Fatalf("cycle_completed count = %d, want 1", counts[taskeventsdomain.EventCycleCompleted])
	}
	if counts[taskeventsdomain.EventPhaseStarted] != 1 {
		t.Fatalf("phase_started count = %d, want 1", counts[taskeventsdomain.EventPhaseStarted])
	}
	if counts[taskeventsdomain.EventPhaseCompleted] != 1 {
		t.Fatalf("phase_completed count = %d, want 1", counts[taskeventsdomain.EventPhaseCompleted])
	}

	calls := h.notifier.snapshot()
	// 4 publishes from cycle/phase row writes (cycle start, execute
	// start, execute complete, cycle terminate) + 1 trailing publish
	// after the final transitionTask succeeds (see process.go: that
	// trailing publish is the cure for the "task stuck in running on
	// the open detail page until refresh" race; the SPA's debounced
	// invalidation needs it to refetch *after* the status row flips).
	if len(calls) != 5 {
		t.Fatalf("notifier publish count = %d, want 5 (calls=%+v)", len(calls), calls)
	}
	for i, c := range calls {
		if c.TaskID != tsk.ID || c.CycleID != cycle.ID {
			t.Fatalf("publish[%d] = %+v, want task=%s cycle=%s", i, c, tsk.ID, cycle.ID)
		}
	}
	runnerCalls := r.Calls()
	if len(runnerCalls) != 1 {
		t.Fatalf("runner call count = %d, want 1 (%#v)", len(runnerCalls), runnerCalls)
	}
	if !strings.Contains(runnerCalls[0].Prompt, "do the thing") {
		t.Fatalf("runner prompt missing task text: %#v", runnerCalls)
	}
}

// TestWorker_StartCycle_recordsRunnerModelAttribution covers the
// per-task runner/model attribution contract: every cycle MUST persist
// both the operator's intent and the runner's resolved effective model
// into TaskCycle.MetaJSON via the CycleMetaProvider interface. The
// audit trail relies on having BOTH so callers can answer "operator
// asked for X but adapter ran Y" without a separate join.
//
// The matrix covers the four meaningful combinations of (operator
// intent, adapter default). The fifth row (both empty) is already
// pinned by TestWorker_HappyPath_writesTwoPhasesAndSixMirrors.
func TestWorker_StartCycle_recordsRunnerModelAttribution(t *testing.T) {
	// Not parallel: four sequential subtests each run worker + in-memory SQLite.
	cases := []struct {
		name          string
		taskModel     string
		runnerDefault string
		wantIntent    string
		wantEffective string
	}{
		{
			name:          "intent_overrides_default",
			taskModel:     "sonnet-4.5",
			runnerDefault: "opus",
			wantIntent:    "sonnet-4.5",
			wantEffective: "sonnet-4.5",
		},
		{
			name:          "intent_only_no_default",
			taskModel:     "sonnet-4.5",
			runnerDefault: "",
			wantIntent:    "sonnet-4.5",
			wantEffective: "sonnet-4.5",
		},
		{
			name:          "default_fills_when_intent_empty",
			taskModel:     "",
			runnerDefault: "opus",
			wantIntent:    "",
			wantEffective: "opus",
		},
		{
			name:          "intent_with_whitespace_falls_back_to_default",
			taskModel:     "   ",
			runnerDefault: "opus",
			// CycleMetaProvider trims the intent string so
			// whitespace-only is persisted as "".
			wantIntent:    "",
			wantEffective: "opus",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Subtests run sequentially: each starts a worker + git prep on
			// in-memory SQLite; nested t.Parallel here races sibling DBs.
			h := newHarness(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tsk := h.createReadyTaskWithModel(ctx, "model-attr-"+tc.name, tc.taskModel)

			r := runnerfake.New().WithDefaultModel(tc.runnerDefault)
			r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
				cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))

			_, done := h.startWorker(ctx, r, worker.Options{})
			h.waitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
			cancel()
			if err := <-done; err != nil {
				t.Fatalf("worker exit err: %v", err)
			}

			cycle := assertCycleStatus(t, h.store, tsk.ID, 1, cyclesdomain.CycleStatusSucceeded)
			var meta map[string]string
			if err := json.Unmarshal(cycle.MetaJSON, &meta); err != nil {
				t.Fatalf("unmarshal meta: %v (raw=%s)", err, cycle.MetaJSON)
			}
			if got := meta["cursor_model_intent"]; got != tc.wantIntent {
				t.Errorf("meta.cursor_model_intent = %q, want %q (full meta=%+v)", got, tc.wantIntent, meta)
			}
			if got := meta["cursor_model_effective"]; got != tc.wantEffective {
				t.Errorf("meta.cursor_model_effective = %q, want %q (full meta=%+v)", got, tc.wantEffective, meta)
			}
			// Pre-existing keys MUST stay populated; this test
			// also pins the no-regression invariant.
			if meta["runner"] != "fake" || meta["runner_version"] != "v0" || len(meta["prompt_hash"]) != 64 {
				t.Errorf("meta legacy keys regressed: %+v", meta)
			}
		})
	}
}
