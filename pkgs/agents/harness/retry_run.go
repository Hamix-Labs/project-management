package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

type startCycleOpts struct {
	parentCycleID *string
	retryMode     taskcoredomain.RetryMode
	runKind       taskcoredomain.PendingRunKind
	instructions  string
	flaggedIDs    []string
	newIDs        []string
	skipVerify    bool
}

// RunWithRetry starts a new cycle. intent==nil is the existing first-run path.
func (h *Harness) RunWithRetry(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	if intent == nil {
		h.runFreshCycle(parentCtx, task, startCycleOpts{})
		return
	}
	if err := intent.Validate(); err != nil {
		slog.Warn("agent harness retry intent invalid", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.RunWithRetry.invalid_intent",
			"task_id", task.ID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_invalid_intent")
		return
	}
	if intent.NormalizeKind() == taskcoredomain.PendingKindPolish {
		h.runPolish(parentCtx, task, intent)
		return
	}
	switch intent.Mode {
	case taskcoredomain.RetryFresh:
		h.runFreshRetry(parentCtx, task, intent)
	case taskcoredomain.RetryResume:
		h.runResumeRetry(parentCtx, task, intent)
	default:
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_invalid_intent")
	}
}

func (h *Harness) runFreshRetry(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	if _, err := h.gitResetForFreshRetry(parentCtx, intent.ParentCycleID); err != nil {
		reason := retryGitResetFailed
		if errors.Is(err, git.ErrRetryResetAnchorMissing) {
			reason = retryResetAnchorMissing
		}
		slog.Warn("agent harness fresh retry git reset failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runFreshRetry.reset_err",
			"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, reason)
		return
	}
	parentID := intent.ParentCycleID
	h.runFreshCycle(parentCtx, task, startCycleOpts{
		parentCycleID: &parentID,
		retryMode:     taskcoredomain.RetryFresh,
	})
}

func (h *Harness) runResumeRetry(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	cp, err := h.loadCheckpointFromParent(parentCtx, intent.ParentCycleID)
	if err != nil {
		slog.Warn("agent harness resume retry checkpoint failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runResumeRetry.checkpoint_err",
			"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_checkpoint_failed")
		return
	}
	startedAt := h.opts.Clock()
	state := processState{
		cycle:  cycleLifecycleState{startedAt: startedAt},
		verify: verifyLifecycleFromCheckpoint(cp, true),
	}
	defer h.recoverFromPanic(&state, *task)

	parentID := intent.ParentCycleID
	cycle, ok := h.startCycle(parentCtx, task, &state, startCycleOpts{
		parentCycleID: &parentID,
		retryMode:     taskcoredomain.RetryResume,
	})
	if !ok {
		h.bestEffortFailTask(parentCtx, task.ID)
		return
	}
	if cp.Entry == resumeEntryVerifyOnly {
		if err := h.resumeSvc().SeedCrossCycleExecuteFromParent(parentCtx, cycle, intent.ParentCycleID); err != nil {
			slog.Warn("agent harness verify-only resume seed execute failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runResumeRetry.seed_execute_err",
				"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
			h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_verify_only_seed_failed")
			return
		}
		if err := h.resumeSvc().MirrorParentCriteriaForVerifyOnly(parentCtx, cycle.ID, intent.ParentCycleID); err != nil {
			slog.Warn("agent harness verify-only resume mirror criteria failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runResumeRetry.mirror_criteria_err",
				"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
			h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_verify_only_mirror_failed")
			return
		}
	}
	h.enterCycleLoopFromCheckpoint(parentCtx, task, cycle, &state, cp, cycleLoopEntryOperatorRetry)
}

// runPolish starts a new attempt from a succeeded parent: always execute (never
// verify-only). Cursor session resume via retry_mode=resume. Always seed locked
// previouslyPassed so InjectCriteria does not re-open accepted criteria (including
// instructions-only / skip_verify). When the operator flagged or added criteria,
// those IDs stay unlocked and verify runs; otherwise skip verify.
func (h *Harness) runPolish(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	cp, err := h.loadCheckpointFromParent(parentCtx, intent.ParentCycleID)
	if err != nil {
		slog.Warn("agent harness polish checkpoint failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runPolish.checkpoint_err",
			"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_checkpoint_failed")
		return
	}
	startedAt := h.opts.Clock()
	previouslyPassed, err := h.seedPolishPreviouslyPassed(parentCtx, task.ID, intent.FlaggedCriterionIDs, intent.NewCriterionIDs)
	if err != nil {
		slog.Warn("agent harness polish seed previouslyPassed failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runPolish.seed_err",
			"task_id", task.ID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_checkpoint_failed")
		return
	}
	state := processState{
		cycle:  cycleLifecycleState{startedAt: startedAt},
		verify: verifyLifecycleState{previouslyPassed: previouslyPassed},
	}
	defer h.recoverFromPanic(&state, *task)

	parentID := intent.ParentCycleID
	cycle, ok := h.startCycle(parentCtx, task, &state, startCycleOpts{
		parentCycleID: &parentID,
		retryMode:     taskcoredomain.RetryResume,
		runKind:       taskcoredomain.PendingKindPolish,
		instructions:  intent.Instructions,
		flaggedIDs:    intent.FlaggedCriterionIDs,
		newIDs:        intent.NewCriterionIDs,
		skipVerify:    intent.SkipVerify,
	})
	if !ok {
		h.bestEffortFailTask(parentCtx, task.ID)
		return
	}
	snap, err := h.loadVerificationSnapshot(parentCtx, task)
	if err != nil {
		slog.Error("agent harness polish verification snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runPolish.verify_snap_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, &state, task.ID, cyclesdomain.CycleStatusFailed, "verification_snapshot_load_failed")
		return
	}
	state.verify.verifySnap = snap
	h.runCycleLoop(parentCtx, task, cycle, &state, cycleLoopOpts{
		resumeNotice: true,
		knownCommits: cp.KnownCommits,
		skipVerify:   intent.SkipVerify,
	})
}

func (h *Harness) seedPolishPreviouslyPassed(ctx context.Context, taskID string, flagged, newIDs []string) (map[string]criterionVerdict, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.seedPolishPreviouslyPassed", "task_id", taskID)
	reopen := make(map[string]struct{}, len(flagged)+len(newIDs))
	for _, id := range flagged {
		reopen[id] = struct{}{}
	}
	for _, id := range newIDs {
		reopen[id] = struct{}{}
	}
	items, err := h.store.ListChecklistForSubject(ctx, taskID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]criterionVerdict, len(items))
	for _, it := range items {
		if _, ok := reopen[it.ID]; ok {
			continue
		}
		if !it.Done {
			continue
		}
		out[it.ID] = criterionVerdict{
			ID:        it.ID,
			Passed:    true,
			Evidence:  it.Evidence,
			Reasoning: it.VerifierReasoning,
		}
	}
	return out, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) runFreshCycle(parentCtx context.Context, task *taskcoredomain.Task, opts startCycleOpts) {
	startedAt := h.opts.Clock()
	state := processState{
		cycle:  cycleLifecycleState{startedAt: startedAt},
		verify: verifyLifecycleState{previouslyPassed: map[string]criterionVerdict{}},
	}
	defer h.recoverFromPanic(&state, *task)

	cycle, ok := h.startCycle(parentCtx, task, &state, opts)
	if !ok {
		h.bestEffortFailTask(parentCtx, task.ID)
		return
	}
	snap, err := h.loadVerificationSnapshot(parentCtx, task)
	if err != nil {
		slog.Error("agent harness fresh-cycle verification snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runFreshCycle.verify_snap_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, &state, task.ID, cyclesdomain.CycleStatusFailed, "verification_snapshot_load_failed")
		return
	}
	state.verify.verifySnap = snap
	h.runCycleLoop(parentCtx, task, cycle, &state, cycleLoopOpts{})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) reconstructCheckpoint(ctx context.Context, cycle *cyclesdomain.TaskCycle) (resumeCheckpoint, error) {
	return h.resumeSvc().ReconstructCheckpoint(ctx, cycle)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) loadCheckpointFromParent(ctx context.Context, parentCycleID string) (resumeCheckpoint, error) {
	return h.resumeSvc().LoadCheckpointFromParent(ctx, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) loadContinuationBundle(ctx context.Context, parentCycleID string) (ContinuationBundle, error) {
	return h.resumeSvc().LoadContinuationBundle(ctx, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) seedCrossCycleExecuteFromParent(ctx context.Context, cycle *cyclesdomain.TaskCycle, parentCycleID string) error {
	return h.resumeSvc().SeedCrossCycleExecuteFromParent(ctx, cycle, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) mirrorParentCriteriaForVerifyOnly(ctx context.Context, childCycleID, parentCycleID string) error {
	return h.resumeSvc().MirrorParentCriteriaForVerifyOnly(ctx, childCycleID, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) failTaskAfterRetryPrep(ctx context.Context, taskID, reason string) {
	h.resumeSvc().FailTaskAfterRetryPrep(ctx, taskID, reason)
}
