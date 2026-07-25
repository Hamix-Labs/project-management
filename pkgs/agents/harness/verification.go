package harness

import (
	"context"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

type criterionVerdict = verify.Verdict
type verificationSnapshot = verify.Snapshot

const verificationFailedReason = verify.FailedReasonPrefix

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) verifySvc() *verify.Service {
	if h.verify == nil {
		h.verify = verify.NewService(verify.Deps{
			Store:      h.store,
			Runner:     h.runner,
			ReportDir:  h.opts.ReportDir,
			WorkingDir: h.opts.WorkingDir,
			Git:        h.gitSvc(),
			Clock:      h.opts.Clock,
			Hooks: verify.Hooks{
				Publish: h.publish,
				PersistProgress: func(ctx context.Context, taskID, cycleID string, phaseSeq int64, ev runner.ProgressEvent) {
					h.persistProgress(ctx, taskID, cycleID, phaseSeq, ev)
					h.publishProgress(taskID, cycleID, phaseSeq, h.phaseRunCorrelationID(), ev)
				},
				RecordVerdict:   h.recordVerifyVerdict,
				ObserveDuration: h.observeVerifyDuration,
				SetRunCancel: func(cancel context.CancelFunc) {
					h.setCurrentRunCancel(cancel)
				},
			},
		})
	}
	h.verify.SetReportDir(h.opts.ReportDir)
	h.verify.SetWorkingDir(h.opts.WorkingDir)
	return h.verify
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) loadVerificationSnapshot(ctx context.Context, taskID string) (verificationSnapshot, error) {
	return h.verifySvc().LoadSnapshot(ctx, taskID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) completeChecklistLegacy(ctx context.Context, taskID string) error {
	return h.verifySvc().CompleteChecklistLegacy(ctx, taskID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) applyVerifiedCompletions(ctx context.Context, taskID, cycleID string, verdicts []criterionVerdict) error {
	if err := h.verifySvc().ApplyVerifiedCompletions(ctx, taskID, cycleID, verdicts); err != nil {
		return err
	}
	// Align with HTTP checklist coherence (ADR-0026): checklist writes must
	// publish task_updated so the SPA does not rely on cycle-frame side effects.
	h.publishTaskUpdated(taskID)
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) runVerificationPipeline(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	snap verificationSnapshot,
	feedback string,
) ([]criterionVerdict, string, error) {
	svc := h.verifySvc()
	svc.SetPlanVerifyRun(func(ctx context.Context, in verify.PlanVerifyRunInput) (verify.VerifyRunPlan, error) {
		return h.planVerifyRun(ctx, in.Task, in.Cycle, state, in.Snap, in.VerifyAttempt, in.Feedback, in.CmdEvidence, in.SelfReport)
	})
	return svc.RunPipeline(parentCtx, task, cycle, snap, state.verify.verifyAttempt, state.verify.previouslyPassed, feedback, state.verify.mirrorDegraded, verify.PhaseCallbacks{
		OnStarted: func(phase *cyclesdomain.TaskCyclePhase) {
			state.phase.runningPhase = cyclesdomain.PhaseVerify
			state.phase.runningPhaseSeq = phase.PhaseSeq
			id := cyclesdomain.RunCorrelationIDFromDetailsJSON(phase.DetailsJSON)
			state.phase.runCorrelationID = id
			h.setPhaseRunCorrelationID(id)
		},
		OnEnded: func() {
			state.phase.lastVerifyAfterExecuteSeq = state.phase.lastCompletedExecutePhaseSeq
			state.phase.runningPhase = ""
			state.phase.runningPhaseSeq = 0
			state.phase.runCorrelationID = ""
			h.setPhaseRunCorrelationID("")
		},
	})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func formatVerificationFailedReason(finalVerdicts []criterionVerdict, lockedPasses map[string]criterionVerdict) string {
	return verify.FormatFailedReason(finalVerdicts, lockedPasses)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func verifyDiffSection(workingDir string) string {
	return verify.DiffSection(workingDir)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) persistCriteriaReports(
	ctx context.Context,
	cycleID string,
	attemptSeq int64,
	criteria []checklistcontract.ChecklistVerifyItem,
	previouslyPassed map[string]criterionVerdict,
	selfReport map[string]reports.CriteriaEntry,
) error {
	return h.verifySvc().PersistCriteriaReports(ctx, cycleID, attemptSeq, criteria, previouslyPassed, selfReport)
}
