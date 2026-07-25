package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// startExecutePhase opens the execute phase row that wraps runner.Run.
// state is updated so the panic-recovery and shutdown branches can find
// the phase to close out.
func (h *Harness) startExecutePhase(ctx context.Context, cycle *cyclesdomain.TaskCycle, state *processState) (*cyclesdomain.TaskCyclePhase, bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.startExecutePhase",
		"cycle_id", cycle.ID)
	exec, err := h.store.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		slog.Warn("agent harness StartPhase(execute) failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.startExecutePhase.err",
			"cycle_id", cycle.ID, "err", err)
		return nil, false
	}
	state.phase.runningPhase = cyclesdomain.PhaseExecute
	state.phase.runningPhaseSeq = exec.PhaseSeq
	state.phase.runCorrelationID = cyclesdomain.RunCorrelationIDFromDetailsJSON(exec.DetailsJSON)
	h.setPhaseRunCorrelationID(state.phase.runCorrelationID)
	h.publish(cycle.TaskID, cycle.ID)
	started := runner.SetupProgressEvent(runner.ProgressRunStateSetupStarted, "Preparing execute…")
	h.persistProgress(ctx, cycle.TaskID, cycle.ID, exec.PhaseSeq, started)
	h.publishProgress(cycle.TaskID, cycle.ID, exec.PhaseSeq, state.phase.runCorrelationID, started)
	return exec, true
}

func (h *Harness) invokeRunnerWithTask(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	exec *cyclesdomain.TaskCyclePhase,
	decision CursorResumeDecision,
) (runner.Result, error) {
	return h.invokeRunnerWithDecision(parentCtx, task, cycle, exec, cyclesdomain.PhaseExecute, task.CursorModel, decision)
}

// invokeRunnerWithDecision runs the runner with a pre-built resume decision.
func (h *Harness) invokeRunnerWithDecision(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	phaseRow *cyclesdomain.TaskCyclePhase,
	phase cyclesdomain.Phase,
	cursorModel string,
	decision CursorResumeDecision,
) (runner.Result, error) {
	runCorrelationID := cyclesdomain.RunCorrelationIDFromDetailsJSON(phaseRow.DetailsJSON)
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.invokeRunnerWithDecision",
		"task_id", task.ID, "cycle_id", cycle.ID, "phase_seq", phaseRow.PhaseSeq,
		"run_correlation_id", runCorrelationID,
		"cursor_resume_mode", string(decision.Mode),
		"recovery_hint_kind", string(decision.RecoveryKind),
		"recovery_hint_bytes", len(decision.Prompt),
		"run_timeout_ns", int64(h.opts.RunTimeout))
	runCtx, cancelCause := context.WithCancelCause(parentCtx)
	if h.opts.RunTimeout > 0 {
		var timeoutCancel context.CancelFunc
		runCtx, timeoutCancel = withRunTimeout(runCtx, h.opts.RunTimeout)
		defer timeoutCancel()
	}
	cancel := func() { cancelCause(context.Canceled) }
	defer cancel()
	projectContext, err := h.selectedProjectContext(runCtx, task, cycle)
	if err != nil {
		details, _ := json.Marshal(map[string]string{"error": err.Error()})
		return runner.NewResult(cyclesdomain.PhaseStatusFailed, "project context selection failed", details, ""), fmt.Errorf("project context: %w: %v", runner.ErrInvalidOutput, err)
	}
	h.setCurrentRunCancel(cancel)
	defer h.setCurrentRunCancel(nil)
	onProgress := func(ev runner.ProgressEvent) {
		h.persistProgress(runCtx, task.ID, cycle.ID, phaseRow.PhaseSeq, ev)
		h.publishProgress(task.ID, cycle.ID, phaseRow.PhaseSeq, runCorrelationID, ev)
	}
	promptText := prompt.WrapWithProjectContext(decision.Prompt, projectContext.Text)
	onProgress(setupInvokeProgress(decision))
	return h.runner.Run(runCtx, runner.Request{
		TaskID:           task.ID,
		AttemptSeq:       cycle.AttemptSeq,
		Phase:            phase,
		Prompt:           promptText,
		WorkingDir:       h.opts.WorkingDir,
		Timeout:          h.opts.RunTimeout,
		CursorModel:      cursorModel,
		RunCorrelationID: runCorrelationID,
		ResumeSessionID:  decision.ResumeSessionID,
		OnProgress:       onProgress,
	})
}

// invokeRunner builds the Request, applies the per-run timeout (if any),
// publishes the cancel func so an operator can interrupt the run, and
// returns whatever the runner produced. <=0 RunTimeout means "no cap":
// the run can only be interrupted by the parent ctx (process shutdown)
// or CancelCurrentRun (operator-initiated). The returned error is the
// raw adapter error (typed via runner.Err* sentinels); classification
// is done by the caller so the shutdown branch can short-circuit it.
// invokeRunner is retained for tests that build task.InitialPrompt directly.
//
//funclogmeasure:skip category=hot-path reason="Test shim; invokeRunnerWithDecision emits trace logs."
func (h *Harness) invokeRunner(parentCtx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle, exec *cyclesdomain.TaskCyclePhase) (runner.Result, error) {
	decision := CursorResumeDecision{Mode: CursorResumeFresh, Prompt: task.InitialPrompt}
	return h.invokeRunnerWithDecision(parentCtx, task, cycle, exec, cyclesdomain.PhaseExecute, task.CursorModel, decision)
}

// progressPersistTimeout bounds AppendCycleStreamEvent after detaching the
// caller's cancel/deadline so a wedged DB cannot hang the harness on shutdown.
const progressPersistTimeout = 5 * time.Second

func (h *Harness) persistProgress(ctx context.Context, taskID, cycleID string, phaseSeq int64, ev runner.ProgressEvent) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.persistProgress",
		"task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq,
		"kind", ev.Kind, "subtype", ev.Subtype)
	if ev.Kind == "" {
		return
	}
	payload := ev.Payload
	if len(payload) == 0 {
		var err error
		payload, err = json.Marshal(ev)
		if err != nil {
			slog.Warn("agent harness progress payload marshal failed",
				"cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.persistProgress.marshal_err",
				"task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq, "err", err)
			payload = []byte("{}")
		}
	}
	// Progress is best-effort observability and must outlive the work context
	// that produced it (kill timers, run timeouts). Detach cancellation but
	// keep values (trace/correlation); bound the write so shutdown cannot leak.
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), progressPersistTimeout)
	defer cancel()
	if _, err := h.store.AppendCycleStreamEvent(writeCtx, cyclesstore.AppendCycleStreamEventInput{
		TaskID:   taskID,
		CycleID:  cycleID,
		PhaseSeq: phaseSeq,
		Source:   progressStreamSource(ev),
		Kind:     ev.Kind,
		Subtype:  ev.Subtype,
		Message:  ev.Message,
		Tool:     ev.Tool,
		Payload:  payload,
	}); err != nil {
		slog.Warn("agent harness progress persistence failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.persistProgress.err",
			"task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq,
			"kind", ev.Kind, "err", err)
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by persistProgress."
func progressStreamSource(ev runner.ProgressEvent) string {
	if ev.Tool == verify.ProgressToolVerifyCommand || ev.Tool == runner.ProgressToolHarnessSetup {
		return "worker"
	}
	return "cursor"
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O."
func setupInvokeProgress(decision CursorResumeDecision) runner.ProgressEvent {
	msg := "Starting Cursor CLI…"
	if decision.Mode == CursorResumeContinue && strings.TrimSpace(decision.ResumeSessionID) != "" {
		msg = "Resuming Cursor session…"
	}
	return runner.SetupProgressEvent(runner.ProgressRunStateSetupInvoke, msg)
}

// withRunTimeout returns parent unchanged when d <= 0; otherwise wraps with WithTimeout.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by invokeRunnerWithDecision."
func withRunTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, d)
}

// completeExecutePhase persists the runner's outcome on the execute
// phase row. Errors from the store are logged and reported back so the
// caller can stop the pipeline (a missing row usually means the task
// was deleted mid-cycle).
func (h *Harness) completeExecutePhase(ctx context.Context, state *processState, cycle *cyclesdomain.TaskCycle, exec *cyclesdomain.TaskCyclePhase, status cyclesdomain.PhaseStatus, result runner.Result, phaseDetails []byte) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.completeExecutePhase",
		"cycle_id", cycle.ID, "phase_seq", exec.PhaseSeq, "status", string(status))
	details := phaseDetails
	if details == nil {
		details = detailsBytes(result)
	}
	in := cyclescontract.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: exec.PhaseSeq,
		Status:   status,
		Details:  details,
		By:       taskcoredomain.ActorAgent,
	}
	if result.Summary != "" {
		s := result.Summary
		in.Summary = &s
	}
	if _, err := h.store.CompletePhase(ctx, in); err != nil {
		level := slog.LevelWarn
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			level = slog.LevelInfo
		}
		slog.Log(ctx, level, "agent harness CompletePhase(execute) failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.completeExecutePhase.err",
			"cycle_id", cycle.ID, "phase_seq", exec.PhaseSeq, "err", err)
		// The phase row is in an indeterminate state (either still
		// running, already terminal, or vanished). Clear the phase
		// pointer so bestEffortTerminate's CompletePhase retry is
		// skipped — but leave cycleStarted=true so the cycle row
		// itself still gets terminated, otherwise the cycle row is
		// orphaned in `running` and the task row is orphaned in
		// `running`, requiring the startup orphan sweep to clean up
		// (see meta.go::detailsBytes for the historical context). The
		// deferred recoverFromPanic only acts on actual panics, so
		// leaving cycleStarted=true here cannot cause a double
		// TerminateCycle on the happy-error path.
		state.phase.runningPhase = ""
		state.phase.runningPhaseSeq = 0
		state.phase.runCorrelationID = ""
		h.setPhaseRunCorrelationID("")
		return false
	}
	state.phase.runningPhase = ""
	state.phase.runningPhaseSeq = 0
	state.phase.runCorrelationID = ""
	h.setPhaseRunCorrelationID("")
	h.publish(cycle.TaskID, cycle.ID)
	return true
}
