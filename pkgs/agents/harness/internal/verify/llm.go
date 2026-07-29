package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func (s *Service) runLLMVerify(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	phaseSeq int64,
	runCorrelationID string,
	snap Snapshot,
	previouslyPassed map[string]Verdict,
	selfReport map[string]reports.CriteriaEntry,
	feedback string,
	cmdEvidence []CommandEvidence,
	verifyAttempt int,
) (cyclesdomain.TokenUsage, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runLLMVerify",
		"task_id", task.ID, "cycle_id", cycle.ID, "locked_passes", len(previouslyPassed))
	s.emitSetupProgress(ctx, task.ID, cycle.ID, phaseSeq,
		runner.SetupProgressEvent(runner.ProgressRunStateSetupPrompt, "Preparing verify…"))
	promptText := buildVerifyPrompt(ctx, s, task.ID, snap, cycle.ID, previouslyPassed, selfReport, feedback, cmdEvidence)
	resumeSessionID := ""
	if s.hooks.PlanVerifyRun != nil {
		plan, err := s.hooks.PlanVerifyRun(ctx, PlanVerifyRunInput{
			Task:             task,
			Cycle:            cycle,
			Snap:             snap,
			VerifyAttempt:    verifyAttempt,
			Feedback:         feedback,
			CmdEvidence:      cmdEvidence,
			SelfReport:       selfReport,
			PreviouslyPassed: previouslyPassed,
		})
		if err != nil {
			return cyclesdomain.TokenUsage{}, false, err
		}
		promptText = plan.Prompt
		resumeSessionID = plan.ResumeSessionID
	}
	var total cyclesdomain.TokenUsage
	var usagePresent bool
	res, err := s.runVerifyCursor(ctx, task, cycle, phaseSeq, runCorrelationID, promptText, resumeSessionID, snap)
	if u, ok := cyclesdomain.TokenUsageFromDetailsJSON(res.Details); ok {
		total = cyclesdomain.AddTokenUsage(total, u)
		usagePresent = true
	}
	if errors.Is(err, runner.ErrResumeSession) {
		// same_chat: hard-fail (ADR-0085). different_chat: also hard-fail on
		// mid-chain resume miss — first verify is already forced fresh.
		err = cursorresume.ResumeSessionFailed(err)
	}
	return total, usagePresent, err
}

// BuildVerifyPrompt exports the full verify prompt composer for harness fallback paths.
func (s *Service) BuildVerifyPrompt(
	ctx context.Context,
	taskID string,
	snap Snapshot,
	cycleID string,
	previouslyPassed map[string]Verdict,
	selfReport map[string]reports.CriteriaEntry,
	feedback string,
	cmdEvidence []CommandEvidence,
) string {
	return buildVerifyPrompt(ctx, s, taskID, snap, cycleID, previouslyPassed, selfReport, feedback, cmdEvidence)
}

func buildVerifyPrompt(
	ctx context.Context,
	s *Service,
	taskID string,
	snap Snapshot,
	cycleID string,
	previouslyPassed map[string]Verdict,
	selfReport map[string]reports.CriteriaEntry,
	feedback string,
	cmdEvidence []CommandEvidence,
) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.buildVerifyPrompt",
		"task_id", taskID, "cycle_id", cycleID, "locked_passes", len(previouslyPassed))
	var b strings.Builder
	b.WriteString("You implemented this task. Now verify each criterion below.\n")
	b.WriteString("Do not modify source files.\n")
	b.WriteString(prompt.FormatVerifyReportContract(
		s.BuildVerifyReportContract(ctx, taskID, snap, cycleID, previouslyPassed, selfReport, feedback, cmdEvidence),
	))
	return b.String()
}

// BuildVerifyReportContract assembles the shared verify-report artifact body
// used by fresh prompts and same-chat resume deltas.
func (s *Service) BuildVerifyReportContract(
	ctx context.Context,
	taskID string,
	snap Snapshot,
	cycleID string,
	previouslyPassed map[string]Verdict,
	selfReport map[string]reports.CriteriaEntry,
	feedback string,
	cmdEvidence []CommandEvidence,
) prompt.VerifyReportContract {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.BuildVerifyReportContract",
		"task_id", taskID, "cycle_id", cycleID, "locked_passes", len(previouslyPassed))
	commits := s.loadTaskCommits(ctx, taskID)
	locked := make([]string, 0, len(previouslyPassed))
	for id := range previouslyPassed {
		locked = append(locked, id)
	}
	sort.Strings(locked)
	criteria := make([]prompt.VerifyCriterionLine, 0, len(snap.Criteria))
	for _, it := range snap.Criteria {
		if _, ok := previouslyPassed[it.ID]; ok {
			continue
		}
		e, ok := selfReport[it.ID]
		if !ok || !e.ClaimedDone {
			continue
		}
		criteria = append(criteria, prompt.VerifyCriterionLine{
			ID:       it.ID,
			Text:     it.Text,
			Evidence: e.Evidence,
		})
	}
	return prompt.VerifyReportContract{
		ReportPath:             reports.VerifyReportPath(s.reportDir, cycleID),
		LockedIDs:              locked,
		Criteria:               criteria,
		CommandEvidenceSection: FormatCommandEvidenceSection(cmdEvidence),
		GitContext:             git.FormatGitContextForPrompt(commits),
		DiffSection:            DiffSection(s.workingDir),
		Feedback:               feedback,
		ToolOnly:               s.toolOnlyReports,
	}
}

func (s *Service) runVerifyCursor(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	phaseSeq int64,
	runCorrelationID string,
	promptText string,
	resumeSessionID string,
	snap Snapshot,
) (runner.Result, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyCursor",
		"task_id", task.ID, "cycle_id", cycle.ID, "phase_seq", phaseSeq,
		"run_correlation_id", runCorrelationID)
	runCtx, cancelCause := context.WithCancelCause(ctx)
	cancel := func() { cancelCause(context.Canceled) }
	if s.hooks.SetRunCancel != nil {
		s.hooks.SetRunCancel(cancel, task.ID)
		defer s.hooks.SetRunCancel(nil, "")
	}
	onProgress := func(ev runner.ProgressEvent) {
		if s.hooks.PersistProgress != nil {
			s.hooks.PersistProgress(ctx, task.ID, cycle.ID, phaseSeq, ev)
		}
	}
	onSessionID := func(sessionID string) {
		if s.hooks.PersistSessionID != nil {
			s.hooks.PersistSessionID(ctx, cycle.ID, phaseSeq, sessionID)
		}
	}
	invokeMsg := "Starting Cursor CLI…"
	if strings.TrimSpace(resumeSessionID) != "" {
		invokeMsg = "Resuming Cursor session…"
	}
	onProgress(runner.SetupProgressEvent(runner.ProgressRunStateSetupInvoke, invokeMsg))
	req := runner.Request{
		TaskID:           task.ID,
		AttemptSeq:       cycle.AttemptSeq,
		Phase:            cyclesdomain.PhaseVerify,
		Prompt:           promptText,
		WorkingDir:       s.workingDir,
		CursorModel:      EffectiveVerifyModel(task, snap),
		RunCorrelationID: runCorrelationID,
		ResumeSessionID:  resumeSessionID,
		OnProgress:       onProgress,
		OnSessionID:      onSessionID,
	}
	if s.hooks.PrepareRunnerRequest != nil {
		if err := s.hooks.PrepareRunnerRequest(ctx, &req, task, cycle); err != nil {
			return runner.NewResult(cyclesdomain.PhaseStatusFailed, "agent MCP prepare failed: "+err.Error(), nil, ""), err
		}
	}
	return s.runner.Run(runCtx, req)
}

func (s *Service) assembleVerdictsFromVerifyReport(
	cycleID string,
	expected map[string]struct{},
	verdicts []Verdict,
	selfReport map[string]reports.CriteriaEntry,
	previouslyPassed map[string]Verdict,
) ([]Verdict, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.assembleVerdictsFromVerifyReport",
		"cycle_id", cycleID, "expected", len(expected))
	if s.hooks.RequireVerifySubmitReceipt != nil {
		if err := s.hooks.RequireVerifySubmitReceipt(cycleID); err != nil {
			return nil, err
		}
	}
	vrep, err := reports.ParseVerifyReport(s.reportDir, cycleID, expected)
	if err != nil {
		return nil, err
	}
	next := make([]Verdict, 0, len(verdicts))
	for _, v := range verdicts {
		if _, locked := previouslyPassed[v.ID]; locked {
			next = append(next, v)
			continue
		}
		if v.Verifier == checklistdomain.VerifierAgentSelf {
			next = append(next, v)
			continue
		}
		entry := selfReport[v.ID]
		vr := vrep[v.ID]
		nv := Verdict{ID: v.ID, Evidence: entry.Evidence}
		if vr.Verified {
			nv.Passed = true
			nv.Verifier = checklistdomain.VerifierExecuteAgent
			nv.Reasoning = vr.Reasoning
		} else {
			nv.Passed = false
			nv.Verifier = checklistdomain.VerifierExecuteAgent
			nv.Reasoning = vr.Reasoning
		}
		next = append(next, nv)
		s.recordVerdict(checklistdomain.VerifierExecuteAgent, nv.Passed)
	}
	return next, nil
}

func verifyLLMRunError(runErr error, parseErr error) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.verifyLLMRunError",
		"has_run_err", runErr != nil, "has_parse_err", parseErr != nil)
	if runErr != nil {
		return runErr
	}
	if parseErr != nil {
		return parseErr
	}
	return nil
}
