// Package runnerfake provides a deterministic in-memory implementation of
// runner.Runner used by every V1 worker test (contract:
// docs/architecture.md). Scripts are keyed on (TaskID, Phase, AttemptSeq)
// with AttemptSeq 0 as a wildcard so tests can pin multi-attempt outcomes
// or script a phase for any attempt.
//
// The fake is exported (capital R Runner) so test files in other packages
// can construct it directly: runnerfake.New() returns a *Runner that
// satisfies runner.Runner.
package runnerfake

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// anyAttemptSeq is the wildcard AttemptSeq for Script/Fail/ScriptProgress.
// Exact ScriptAttempt keys take precedence over the wildcard on lookup.
const anyAttemptSeq int64 = 0

// Runner is a deterministic fake implementation of runner.Runner.
//
// Tests register expected outcomes with Script (success) or Fail (error
// path), then assert against Calls() afterwards. Run lookups that have no
// matching script return runner.ErrInvalidOutput so missing scripts surface
// as test failures rather than silently passing.
type Runner struct {
	name         string
	version      string
	defaultModel string
	autoSession  bool

	mu      sync.Mutex
	scripts map[scriptKey]scripted
	calls   []runner.Request
}

type scriptKey struct {
	taskID     string
	phase      cyclesdomain.Phase
	attemptSeq int64
}

type scripted struct {
	result     runner.Result
	err        error
	progress   []runner.ProgressEvent
	hasOutcome bool // true after Script/Fail*; progress-only entries are not runnable
}

// New returns a fake runner with default name "fake" and version "v0".
// Override either via WithName / WithVersion before scripting if a test
// needs to assert on Runner.Name / Runner.Version (the worker records
// these in TaskCyclePhase.MetaJSON).
func New() *Runner {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.New")
	return &Runner{
		name:        "fake",
		version:     "v0",
		scripts:     make(map[scriptKey]scripted),
		autoSession: true,
	}
}

// WithoutAutoSessionID disables the default injection of details_json.session_id
// on successful Runs (needed for same-chat hard-fail tests).
//
//funclogmeasure:skip category=hot-path reason="Test helper setter; Run emits operation traces."
func (r *Runner) WithoutAutoSessionID() *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoSession = false
	return r
}

// WithName overrides the value returned by Name().
func (r *Runner) WithName(name string) *Runner {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.WithName", "name", name)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.name = name
	return r
}

// WithVersion overrides the value returned by Version().
func (r *Runner) WithVersion(version string) *Runner {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.WithVersion", "version", version)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.version = version
	return r
}

// WithDefaultModel sets the model that EffectiveModel returns when
// req.CursorModel is empty. Mirrors the cursor adapter's
// DefaultCursorModel option so worker tests can pin the
// cursor_model_effective audit value end-to-end.
func (r *Runner) WithDefaultModel(model string) *Runner {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.WithDefaultModel", "model", model)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultModel = model
	return r
}

// Script registers result as the value Run will return for (taskID, phase)
// on any AttemptSeq. Last write wins for the wildcard key; exact
// ScriptAttempt entries still take precedence at lookup time.
// Result is stored as-is; tests should typically build it via
// runner.NewResult so caps are applied.
func (r *Runner) Script(taskID string, phase cyclesdomain.Phase, result runner.Result) {
	r.ScriptAttempt(taskID, phase, anyAttemptSeq, result)
}

// ScriptAttempt registers result for an exact (taskID, phase, attemptSeq).
func (r *Runner) ScriptAttempt(taskID string, phase cyclesdomain.Phase, attemptSeq int64, result runner.Result) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.ScriptAttempt",
		"task_id", taskID, "phase", string(phase), "attempt_seq", attemptSeq)
	r.mu.Lock()
	defer r.mu.Unlock()
	key := scriptKey{taskID: taskID, phase: phase, attemptSeq: attemptSeq}
	entry := r.scripts[key]
	entry.result = result
	entry.err = nil
	entry.hasOutcome = true
	r.scripts[key] = entry
}

// Fail registers err as the error Run will return for (taskID, phase) on
// any AttemptSeq. The accompanying result is the zero Result (mirroring
// the contract of runner.ErrInvalidOutput).
func (r *Runner) Fail(taskID string, phase cyclesdomain.Phase, err error) {
	r.FailAttempt(taskID, phase, anyAttemptSeq, err)
}

// FailAttempt registers err for an exact (taskID, phase, attemptSeq).
func (r *Runner) FailAttempt(taskID string, phase cyclesdomain.Phase, attemptSeq int64, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.FailAttempt",
		"task_id", taskID, "phase", string(phase), "attempt_seq", attemptSeq, "err", err)
	r.mu.Lock()
	defer r.mu.Unlock()
	key := scriptKey{taskID: taskID, phase: phase, attemptSeq: attemptSeq}
	entry := r.scripts[key]
	entry.err = err
	entry.result = runner.Result{}
	entry.hasOutcome = true
	r.scripts[key] = entry
}

// FailWithResult registers (result, err) as the pair Run will return for
// (taskID, phase) on any AttemptSeq. Used when the adapter contract
// requires both a partial Result and a typed error (e.g. ErrNonZeroExit
// with the captured RawOutput).
func (r *Runner) FailWithResult(taskID string, phase cyclesdomain.Phase, result runner.Result, err error) {
	r.FailWithResultAttempt(taskID, phase, anyAttemptSeq, result, err)
}

// FailWithResultAttempt registers (result, err) for an exact attempt.
func (r *Runner) FailWithResultAttempt(taskID string, phase cyclesdomain.Phase, attemptSeq int64, result runner.Result, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.FailWithResultAttempt",
		"task_id", taskID, "phase", string(phase), "attempt_seq", attemptSeq, "err", err)
	r.mu.Lock()
	defer r.mu.Unlock()
	key := scriptKey{taskID: taskID, phase: phase, attemptSeq: attemptSeq}
	entry := r.scripts[key]
	entry.result = result
	entry.err = err
	entry.hasOutcome = true
	r.scripts[key] = entry
}

// ScriptProgress registers progress events that Run will invoke via
// req.OnProgress (when non-nil) before returning the scripted outcome
// for (taskID, phase) on any AttemptSeq. Events are stored even when no
// result has been scripted yet; Run still requires a Script/Fail entry.
func (r *Runner) ScriptProgress(taskID string, phase cyclesdomain.Phase, events ...runner.ProgressEvent) {
	r.ScriptProgressAttempt(taskID, phase, anyAttemptSeq, events...)
}

// ScriptProgressAttempt registers progress events for an exact attempt.
func (r *Runner) ScriptProgressAttempt(taskID string, phase cyclesdomain.Phase, attemptSeq int64, events ...runner.ProgressEvent) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.ScriptProgressAttempt",
		"task_id", taskID, "phase", string(phase), "attempt_seq", attemptSeq, "events", len(events))
	r.mu.Lock()
	defer r.mu.Unlock()
	key := scriptKey{taskID: taskID, phase: phase, attemptSeq: attemptSeq}
	entry := r.scripts[key]
	entry.progress = append([]runner.ProgressEvent(nil), events...)
	r.scripts[key] = entry
}

// Run looks up the scripted outcome for (req.TaskID, req.Phase, req.AttemptSeq),
// falling back to the AttemptSeq-wildcard script when no exact match exists.
// Scripted progress events are delivered via req.OnProgress before the
// result is returned. When no script is registered it returns
// runner.ErrInvalidOutput so missing expectations fail tests loudly.
// Run honours ctx cancellation: a cancelled context returns ctx.Err()
// wrapped with runner.ErrTimeout so callers can errors.Is against the
// typed-error contract.
func (r *Runner) Run(ctx context.Context, req runner.Request) (runner.Result, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.Run",
		"task_id", req.TaskID, "phase", string(req.Phase), "attempt_seq", req.AttemptSeq)

	if err := ctx.Err(); err != nil {
		return runner.Result{}, fmt.Errorf("runnerfake: %w: %v", runner.ErrTimeout, err)
	}

	r.mu.Lock()
	r.calls = append(r.calls, req)
	entry, ok := r.lookupLocked(req.TaskID, req.Phase, req.AttemptSeq)
	r.mu.Unlock()

	if !ok {
		return runner.Result{}, fmt.Errorf("runnerfake: %w: no script for (task_id=%s, phase=%s, attempt_seq=%d)",
			runner.ErrInvalidOutput, req.TaskID, req.Phase, req.AttemptSeq)
	}
	if req.OnProgress != nil {
		for _, ev := range entry.progress {
			req.OnProgress(ev)
		}
	}
	result := entry.result
	r.mu.Lock()
	autoSession := r.autoSession
	r.mu.Unlock()
	if entry.err == nil && autoSession {
		result = ensureFakeSessionID(result, req)
	}
	if entry.err == nil && req.Phase == cyclesdomain.PhaseExecute {
		if seedErr := seedCommitRegisterForTests(req); seedErr != nil {
			return runner.Result{}, fmt.Errorf("runnerfake: seed commit register: %w", seedErr)
		}
	}
	return result, entry.err
}

func ensureFakeSessionID(result runner.Result, req runner.Request) runner.Result {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.ensureFakeSessionID",
		"task_id", req.TaskID, "phase", string(req.Phase))
	if cyclesdomain.SessionIDFromDetailsJSON(result.Details) != "" {
		return result
	}
	id := fmt.Sprintf("fake-sess-%s-%s-%d", req.TaskID, req.Phase, req.AttemptSeq)
	var m map[string]any
	if len(result.Details) > 0 {
		_ = json.Unmarshal(result.Details, &m)
	}
	if m == nil {
		m = map[string]any{}
	}
	m[cyclesdomain.PhaseDetailsSessionID] = id
	b, err := json.Marshal(m)
	if err != nil {
		return result
	}
	result.Details = b
	return result
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (r *Runner) lookupLocked(taskID string, phase cyclesdomain.Phase, attemptSeq int64) (scripted, bool) {
	if entry, ok := r.scripts[scriptKey{taskID: taskID, phase: phase, attemptSeq: attemptSeq}]; ok && entry.hasOutcome {
		return entry, true
	}
	if attemptSeq != anyAttemptSeq {
		if entry, ok := r.scripts[scriptKey{taskID: taskID, phase: phase, attemptSeq: anyAttemptSeq}]; ok && entry.hasOutcome {
			return entry, true
		}
	}
	return scripted{}, false
}

// Name returns the configured runner name (default "fake").
func (r *Runner) Name() string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.Name")
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.name
}

// Version returns the configured runner version (default "v0").
func (r *Runner) Version() string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.Version")
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.version
}

// EffectiveModel implements runner.Runner. Mirrors the cursor adapter
// fallback: trim req.CursorModel and use it when non-empty; otherwise
// fall back to the value set via WithDefaultModel (default "").
func (r *Runner) EffectiveModel(req runner.Request) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.EffectiveModel",
		"task_id", req.TaskID)
	m := strings.TrimSpace(req.CursorModel)
	if m != "" {
		return m
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.defaultModel
}

// Calls returns a copy of every Request seen by Run, in invocation order.
// Tests use this to assert on what the worker sent to the runner.
func (r *Runner) Calls() []runner.Request {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.Calls")
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]runner.Request, len(r.calls))
	copy(out, r.calls)
	return out
}

// Reset clears recorded calls and registered scripts. Useful in
// table-driven tests that share one *Runner across subtests.
func (r *Runner) Reset() {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runnerfake.Runner.Reset")
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scripts = make(map[scriptKey]scripted)
	r.calls = nil
}

// Compile-time assertion that *Runner implements runner.Runner.
var _ runner.Runner = (*Runner)(nil)
