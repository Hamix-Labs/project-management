package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// processState records what the worker has written so far for a single
// task. Nested bags are stage-scoped where possible (B-34):
//   - cycle: cross-iteration identity (id, started, effective model)
//   - phase: execute/verify phase seqs and correlation for the current loop
//   - verify: verify-attempt scratch (feedback, prior passes); reset on new cycle
//   - git: execute-outcome anchors handed to verify/commit ingest
//   - resume: entry mirrors for crash/operator resume (not live mid-execute)
//
// Prefer passing stage values as arguments when adding new control flow;
// only put fields here when panic/shutdown cleanup or multi-iteration
// verify retry must consult them.
type cycleLifecycleState struct {
	cycleID        string
	cycleStarted   bool
	startedAt      time.Time
	effectiveModel string
}

type phaseLifecycleState struct {
	runningPhase                 cyclesdomain.Phase
	runningPhaseSeq              int64
	runCorrelationID             string
	executeReachedVerify         bool
	lastCompletedExecutePhaseSeq int64
	lastVerifyAfterExecuteSeq    int64
}

type verifyLifecycleState struct {
	verifySnap     verificationSnapshot
	verifyAttempt  int
	verifyFeedback string
	// previouslyPassed accumulates criterion verdicts that earlier
	// retry attempts proved passed. Keyed by criterion ID; carried in
	// memory across the retry loop so the next execute prompt can list
	// these items as "already verified, do not re-do" and the next
	// verify pass can short-circuit them. The atomic-decision contract
	// (docs/data-model.md "Worker verification loop") is preserved
	// because nothing here is committed to task_checklist_completions
	// until the cycle succeeds and applyVerifiedCompletions is called
	// with the union. On terminal failure the map is discarded.
	previouslyPassed   map[string]criterionVerdict
	lastFailedVerdicts []criterionVerdict
	reportParseErr     string
	reportTampered     bool
	mirrorDegraded     bool
}

type gitLifecycleState struct {
	gitSnap            git.PhaseSnapshot
	postExecuteHeadSHA string
	lastCommitIngestOK bool
}

type resumeMirrorState struct {
	continuation         *ContinuationBundle
	resumeNotice         bool
	interruptedPhase     cyclesdomain.Phase
	lastCursorResumeMode CursorResumeMode
}

type processState struct {
	cycle  cycleLifecycleState
	phase  phaseLifecycleState
	verify verifyLifecycleState
	git    gitLifecycleState
	resume resumeMirrorState
}

// Run drives the harness cycle body for one task already in StatusRunning.
// The worker owns queue admission (reload, readiness, ready→running) and
// ack ordering before calling Run.
func (h *Harness) Run(parentCtx context.Context, task *taskcoredomain.Task) {
	h.RunWithRetry(parentCtx, task, nil)
}

// transitionTask flips the task to next; returns false on any store
// error (including ErrNotFound when the task was deleted mid-cycle).
func (h *Harness) transitionTask(ctx context.Context, taskID string, next taskcoredomain.Status, op string) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.transitionTask",
		"task_id", taskID, "next", string(next), "op", op)
	if _, err := h.store.Update(ctx, taskID, taskcorestore.UpdateTaskInput{Status: &next}, taskcoredomain.ActorAgent); err != nil {
		level := slog.LevelWarn
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			level = slog.LevelInfo
		}
		slog.Log(ctx, level, "agent harness task transition failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.transitionTask.err",
			"task_id", taskID, "next", string(next), "op", op, "err", err)
		return false
	}
	h.publishTaskUpdated(taskID)
	return true
}

// startCycle writes the StartCycle row and updates state on success.
// MetaJSON carries runner identity, prompt hash, AND the operator's
// model intent + the runner's resolved effective model (Phase 1a-ii of
// the per-task runner/model attribution plan) so the audit trail and
// observability slice-and-dice can distinguish runs by adapter
// version, intent, and effective model — without depending on runtime
// metric scrapes.
//
// The Request is the same shape invokeRunner builds later (sans
// per-run timeout, which is irrelevant to attribution). Intent is
// Runner-specific metadata (e.g. model intent/effective) comes from
// the Attributor interface (metrics labels + cycle meta). Both may produce
// "" and that empty string is the truth, not a placeholder.
func (h *Harness) startCycle(ctx context.Context, task *taskcoredomain.Task, state *processState, opts startCycleOpts) (*cyclesdomain.TaskCycle, bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.startCycle",
		"task_id", task.ID)
	req := runner.Request{
		TaskID:      task.ID,
		Phase:       cyclesdomain.PhaseExecute,
		Prompt:      task.InitialPrompt,
		WorkingDir:  h.opts.WorkingDir,
		CursorModel: task.CursorModel,
	}
	meta := buildCycleMeta(h.runner, task.InitialPrompt, req)
	if opts.retryMode != "" {
		meta = mergeCycleMetaBytes(meta, map[string]any{"retry_mode": string(opts.retryMode)})
	}
	if opts.runKind != "" {
		meta = mergeCycleMetaBytes(meta, map[string]any{"run_kind": string(opts.runKind)})
	}
	if strings.TrimSpace(opts.instructions) != "" {
		meta = mergeCycleMetaBytes(meta, map[string]any{"polish_instructions": strings.TrimSpace(opts.instructions)})
	}
	in := cyclescontract.StartCycleInput{
		TaskID:        task.ID,
		TriggeredBy:   taskcoredomain.ActorAgent,
		ParentCycleID: opts.parentCycleID,
		Meta:          meta,
	}
	cycle, err := h.store.StartCycle(ctx, in)
	if err != nil {
		slog.Warn("agent harness StartCycle failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.startCycle.err", "task_id", task.ID, "err", err)
		return nil, false
	}
	state.cycle.cycleID = cycle.ID
	state.cycle.cycleStarted = true
	if attr, ok := h.runner.(runner.Attributor); ok {
		state.cycle.effectiveModel = attr.MetricsLabels(req)["model"]
	} else if ml, ok := h.runner.(runner.MetricsLabeler); ok {
		state.cycle.effectiveModel = ml.MetricsLabels(req)["model"]
	} else {
		state.cycle.effectiveModel = h.runner.EffectiveModel(req)
	}
	h.publish(task.ID, cycle.ID)
	return cycle, true
}

// terminateCycle closes the cycle row and clears state so the recovery
// path is a no-op for already-terminal cycles. Records one metrics
// observation on success so cmd/taskapi's Prometheus counter +
// histogram see the happy-path attempt outcome.
func (h *Harness) terminateCycle(ctx context.Context, state *processState, taskID string, status cyclesdomain.CycleStatus, reason string) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.terminateCycle",
		"cycle_id", state.cycle.cycleID, "status", string(status), "reason", reason)
	if state.cycle.cycleID == "" {
		return true
	}
	if _, err := h.store.TerminateCycle(ctx, state.cycle.cycleID, status, reason, taskcoredomain.ActorAgent); err != nil {
		level := slog.LevelWarn
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			level = slog.LevelInfo
		}
		slog.Log(ctx, level, "agent harness TerminateCycle failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.terminateCycle.err",
			"cycle_id", state.cycle.cycleID, "err", err)
		state.cycle.cycleStarted = false
		return false
	}
	state.cycle.cycleStarted = false
	h.publish(taskID, state.cycle.cycleID)
	h.recordRun(string(status), h.runner.Name(), state.cycle.effectiveModel, state.cycle.startedAt)
	h.observeVerifyRetries(state.verify.verifyAttempt)
	// GC the worker-managed scratch dir for this cycle. Idempotent
	// against a missing dir; logged at Debug because operators rarely
	// care unless cleanup itself errors. Closes the unbounded-disk-
	// growth gap that existed when files were written under RepoRoot/.legacy-scratch.
	if err := reports.CleanupReportDir(h.opts.ReportDir, state.cycle.cycleID); err != nil {
		slog.Warn("agent harness cleanupReportDir failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.terminateCycle.cleanup_err",
			"cycle_id", state.cycle.cycleID, "report_dir", h.opts.ReportDir, "err", err)
	}
	return true
}
