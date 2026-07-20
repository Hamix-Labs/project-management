package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestErrSentinels_distinct(t *testing.T) {
	t.Parallel()

	all := []error{runner.ErrTimeout, runner.ErrNonZeroExit, runner.ErrInvalidOutput}
	for i, a := range all {
		for j, b := range all {
			if i == j {
				continue
			}
			if errors.Is(a, b) {
				t.Errorf("sentinel %d (%v) is wrongly matched by %d (%v)", i, a, j, b)
			}
		}
	}
}

// TestRunnerInterface_compileTime asserts at least one type satisfies the
// interface so signature drift surfaces here, not in callers.
func TestRunnerInterface_compileTime(t *testing.T) {
	t.Parallel()
	var _ runner.Runner = runnerfake.New()
}

// TestRunnerFake_EffectiveModel pins the contract every adapter must
// satisfy: trim req.CursorModel and return it when non-empty,
// otherwise fall back to the adapter's configured default. The empty
// return value is intentional and means "no model configured anywhere"
// â€” callers (worker recordRun, buildCycleMeta) MUST persist that
// empty string verbatim into TaskCycle.MetaJSON / Prometheus labels so
// the audit trail can distinguish pre-feature cycles from explicit
// "use the global default" rows.
func TestRunnerFake_EffectiveModel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		defaultModel string
		reqModel     string
		want         string
	}{
		{"both empty", "", "", ""},
		{"default only", "opus", "", "opus"},
		{"request overrides default", "opus", "sonnet-4.5", "sonnet-4.5"},
		{"request whitespace falls back", "opus", "   ", "opus"},
		{"request trimmed", "opus", "  sonnet-4.5  ", "sonnet-4.5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := runnerfake.New().WithDefaultModel(tc.defaultModel)
			got := r.EffectiveModel(runner.Request{CursorModel: tc.reqModel})
			if got != tc.want {
				t.Errorf("EffectiveModel(req.CursorModel=%q) with default %q: got %q want %q",
					tc.reqModel, tc.defaultModel, got, tc.want)
			}
		})
	}
}

// TestRunnerFake_returnsScriptedResult covers the success path of the fake
// (used by every later test in the worker plan).
func TestRunnerFake_returnsScriptedResult(t *testing.T) {
	t.Parallel()

	want := runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "ok", nil, "")
	r := runnerfake.New()
	r.Script("task-A", cyclesdomain.PhaseExecute, want)

	got, err := r.Run(context.Background(), runner.Request{
		TaskID:     "task-A",
		AttemptSeq: 1,
		Phase:      cyclesdomain.PhaseExecute,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got.Status != want.Status || got.Summary != want.Summary {
		t.Errorf("got %+v want %+v", got, want)
	}
	if calls := r.Calls(); len(calls) != 1 {
		t.Fatalf("Calls len: got %d want 1", len(calls))
	}
}

// TestRunnerFake_ScriptAttempt_prefersExactOverWildcard pins multi-attempt
// scripting: an exact AttemptSeq entry wins over the wildcard Script.
func TestRunnerFake_ScriptAttempt_prefersExactOverWildcard(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.Script("task-A", cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "any", nil, ""))
	r.ScriptAttempt("task-A", cyclesdomain.PhaseExecute, 2, runner.NewResult(
		cyclesdomain.PhaseStatusFailed, "attempt-2", nil, ""))

	got1, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-A", AttemptSeq: 1, Phase: cyclesdomain.PhaseExecute,
	})
	if err != nil {
		t.Fatalf("attempt 1: %v", err)
	}
	if got1.Summary != "any" {
		t.Fatalf("attempt 1 summary = %q, want any", got1.Summary)
	}

	got2, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-A", AttemptSeq: 2, Phase: cyclesdomain.PhaseExecute,
	})
	if err != nil {
		t.Fatalf("attempt 2: %v", err)
	}
	if got2.Summary != "attempt-2" || got2.Status != cyclesdomain.PhaseStatusFailed {
		t.Fatalf("attempt 2 = %+v, want failed/attempt-2", got2)
	}
}

// TestRunnerFake_ScriptProgress_invokesOnProgress ensures mid-run progress
// callbacks fire before the scripted result is returned.
func TestRunnerFake_ScriptProgress_invokesOnProgress(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.Script("task-P", cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "done", nil, ""))
	r.ScriptProgress("task-P", cyclesdomain.PhaseExecute,
		runner.ProgressEvent{Kind: "tool_call", Subtype: "started", Tool: "Read", Message: "reading"},
		runner.ProgressEvent{Kind: "assistant", Message: "ok"},
	)

	var seen []runner.ProgressEvent
	got, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-P", AttemptSeq: 1, Phase: cyclesdomain.PhaseExecute,
		OnProgress: func(ev runner.ProgressEvent) {
			seen = append(seen, ev)
		},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got.Summary != "done" {
		t.Fatalf("summary = %q, want done", got.Summary)
	}
	if len(seen) != 2 {
		t.Fatalf("progress events = %d, want 2 (%+v)", len(seen), seen)
	}
	if seen[0].Tool != "Read" || seen[1].Kind != "assistant" {
		t.Fatalf("progress = %+v", seen)
	}
}

// TestRunnerFake_ScriptProgress_nilOnProgressIsSafe covers runs without a
// progress callback (worker paths that omit live updates).
func TestRunnerFake_ScriptProgress_nilOnProgressIsSafe(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.Script("task-N", cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "ok", nil, ""))
	r.ScriptProgress("task-N", cyclesdomain.PhaseExecute,
		runner.ProgressEvent{Kind: "stream", Message: "ignored"})

	if _, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-N", Phase: cyclesdomain.PhaseExecute,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// TestRunnerFake_ScriptProgressAlone_notRunnable ensures progress without a
// Script/Fail outcome still fails loudly.
func TestRunnerFake_ScriptProgressAlone_notRunnable(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.ScriptProgress("task-X", cyclesdomain.PhaseExecute,
		runner.ProgressEvent{Kind: "stream", Message: "orphan"})

	_, err := r.Run(context.Background(), runner.Request{
		TaskID: "task-X", Phase: cyclesdomain.PhaseExecute,
	})
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Fatalf("got %v, want ErrInvalidOutput", err)
	}
}

// TestRunnerFake_unknownScriptReturnsErrInvalidOutput keeps the fake honest:
// missing scripts must be loud failures so worker tests don't pass on the
// wrong code path.
func TestRunnerFake_unknownScriptReturnsErrInvalidOutput(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	_, err := r.Run(context.Background(), runner.Request{
		TaskID: "missing", Phase: cyclesdomain.PhaseExecute,
	})
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Errorf("got %v want errors.Is(_, ErrInvalidOutput)", err)
	}
}

// TestRunnerFake_failWithErr propagates a typed error to the caller.
func TestRunnerFake_failWithErr(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.Fail("task-B", cyclesdomain.PhaseVerify, runner.ErrNonZeroExit)
	_, err := r.Run(context.Background(), runner.Request{TaskID: "task-B", Phase: cyclesdomain.PhaseVerify})
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Errorf("got %v want errors.Is(_, ErrNonZeroExit)", err)
	}
}

// TestRunnerFake_failWithResult covers the worker's "partial result + typed
// error" contract (e.g. ErrNonZeroExit alongside captured RawOutput).
func TestRunnerFake_failWithResult(t *testing.T) {
	t.Parallel()

	partial := runner.NewResult(cyclesdomain.PhaseStatusFailed, "exit 2", nil, "stderr blob")
	r := runnerfake.New()
	r.FailWithResult("task-C", cyclesdomain.PhaseExecute, partial, runner.ErrNonZeroExit)

	got, err := r.Run(context.Background(), runner.Request{TaskID: "task-C", Phase: cyclesdomain.PhaseExecute})
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Fatalf("err: got %v want ErrNonZeroExit", err)
	}
	if got.Status != cyclesdomain.PhaseStatusFailed || got.RawOutput != "stderr blob" {
		t.Errorf("partial Result lost: got %+v", got)
	}
}

// TestRunnerFake_cancelledContextReturnsErrTimeout proves the fake honours
// context cancellation through the typed-error channel, so worker tests can
// exercise timeout paths without spinning a real CLI.
func TestRunnerFake_cancelledContextReturnsErrTimeout(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.Script("task-D", cyclesdomain.PhaseExecute, runner.Result{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Run(ctx, runner.Request{TaskID: "task-D", Phase: cyclesdomain.PhaseExecute})
	if !errors.Is(err, runner.ErrTimeout) {
		t.Errorf("got %v want errors.Is(_, ErrTimeout)", err)
	}
}

// TestRunnerFake_NameVersionDefaultsAndOverrides documents that Name and
// Version can be customised so adapter-conformance tests in later stages
// can pin MetaJSON values.
func TestRunnerFake_NameVersionDefaultsAndOverrides(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	if r.Name() != "fake" || r.Version() != "v0" {
		t.Errorf("defaults: name=%q version=%q", r.Name(), r.Version())
	}
	r.WithName("cursor-fake").WithVersion("0.42.0")
	if r.Name() != "cursor-fake" || r.Version() != "0.42.0" {
		t.Errorf("overrides not applied: name=%q version=%q", r.Name(), r.Version())
	}
}

// TestRunnerFake_Reset clears recorded calls and scripts.
func TestRunnerFake_Reset(t *testing.T) {
	t.Parallel()

	r := runnerfake.New()
	r.Script("task-E", cyclesdomain.PhaseExecute, runner.Result{Status: cyclesdomain.PhaseStatusSucceeded})
	if _, err := r.Run(context.Background(), runner.Request{TaskID: "task-E", Phase: cyclesdomain.PhaseExecute}); err != nil {
		t.Fatalf("seed Run: %v", err)
	}
	if len(r.Calls()) != 1 {
		t.Fatalf("seed call count")
	}

	r.Reset()
	if len(r.Calls()) != 0 {
		t.Errorf("Calls not cleared by Reset")
	}
	_, err := r.Run(context.Background(), runner.Request{TaskID: "task-E", Phase: cyclesdomain.PhaseExecute})
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Errorf("script not cleared by Reset: got %v", err)
	}
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
