package resume

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// LoadContinuationBundle rehydrates cross-cycle resume context from a parent attempt.
func (s *Service) LoadContinuationBundle(ctx context.Context, parentCycleID string) (ContinuationBundle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.LoadContinuationBundle",
		"parent_cycle_id", parentCycleID)
	var bundle ContinuationBundle
	bundle.PreviouslyPassed = map[string]CriterionVerdict{}
	parentCycleID = strings.TrimSpace(parentCycleID)
	if parentCycleID == "" {
		return bundle, fmt.Errorf("continuation: empty parent cycle id")
	}
	cycle, err := s.store.GetCycle(ctx, parentCycleID)
	if err != nil {
		return bundle, err
	}
	if !cyclesdomain.TerminalCycleStatus(cycle.Status) {
		return bundle, fmt.Errorf("continuation: parent cycle %q is not terminal", cycle.Status)
	}
	bundle.ParentCycleID = parentCycleID
	bundle.LineageAttempt = cycle.AttemptSeq

	phases, err := s.store.ListPhasesForCycle(ctx, parentCycleID)
	if err != nil {
		return bundle, err
	}
	bundle.FailureReason = parentFailureReason(phases, cycle)

	commits, err := s.loadKnownCommitsForTask(ctx, cycle.TaskID)
	if err != nil {
		return bundle, err
	}
	bundle.Commits = commits

	if len(phases) == 0 {
		bundle.Warnings = append(bundle.Warnings, "parent cycle has no phases")
		bundle.Entry = EntryExecute
	} else {
		lastPhase := phases[len(phases)-1]
		bundle.FailurePhase = lastPhase.Phase
		bundle.FailureClass = classifyParentFailure(phases, cycle, lastPhase)
		lastExecute := lastExecutePhase(phases)
		if lastExecute != nil {
			bundle.ScopeFiles = git.ScopeFilesFromPhaseDetails(ctx, s.gitRepo(), s.opts.WorkingDir, lastExecute.DetailsJSON)
			bundle.RunnerFeedback = runnerFeedbackFromPhase(lastExecute)
			bundle.CriteriaReportProbeErr = git.CriteriaReportProbeErrFromPhaseDetails(lastExecute.DetailsJSON)
			if lastExecute.Status == cyclesdomain.PhaseStatusFailed {
				if summary := phaseSummary(*lastExecute); summary != "" {
					bundle.ExecuteFeedback = "Prior execute attempt failed: " + summary
				}
			}
		}
		bundle.Entry = routeResumeEntry(phases, lastExecute, lastPhase, cycle, len(bundle.Commits) > 0)
	}

	previouslyPassed, _, verifyFeedback, _, err := s.loadVerifyCheckpointData(ctx, parentCycleID, phases)
	if err != nil {
		return bundle, err
	}
	bundle.PreviouslyPassed = previouslyPassed
	bundle.VerifyFeedback = verifyFeedback

	criteriaRows, err := s.store.ListCriteriaReportsForCycle(ctx, parentCycleID)
	if err != nil {
		return bundle, err
	}
	for i := range criteriaRows {
		if criteriaRows[i].AttemptSeq == cyclesdomain.ExecuteCriteriaReportAttemptSeq {
			bundle.CriteriaEvidence = append(bundle.CriteriaEvidence, criteriaRows[i])
		}
	}

	bundle.Sufficient = continuationSufficient(bundle, cycle)
	if !bundle.Sufficient {
		bundle.Warnings = append(bundle.Warnings, "insufficient continuation data for parent attempt")
	}
	return bundle, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func classifyParentFailure(phases []cyclesdomain.TaskCyclePhase, cycle *cyclesdomain.TaskCycle, lastPhase cyclesdomain.TaskCyclePhase) FailureClass {
	reason := parentFailureReason(phases, cycle)
	if reason == "" {
		reason = phaseSummary(lastPhase)
	}
	if reason == cancelledByOperatorReason {
		return FailureClassOperator
	}
	if strings.HasPrefix(reason, verificationFailedReason) || lastPhase.Phase == cyclesdomain.PhaseVerify {
		return FailureClassVerify
	}
	if strings.HasPrefix(reason, "runner_") || strings.Contains(reason, "runner_") {
		return FailureClassRunner
	}
	if lastPhase.Phase == cyclesdomain.PhaseExecute && lastPhase.Status == cyclesdomain.PhaseStatusFailed {
		return FailureClassRunner
	}
	if reason == "shutdown" || reason == "panic" || strings.HasSuffix(reason, "_failed") {
		return FailureClassInfrastructure
	}
	return FailureClassInfrastructure
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func routeResumeEntry(phases []cyclesdomain.TaskCyclePhase, lastExecute *cyclesdomain.TaskCyclePhase, lastPhase cyclesdomain.TaskCyclePhase, cycle *cyclesdomain.TaskCycle, hasCommits bool) Entry {
	reason := parentFailureReason(phases, cycle)
	if reason == "" {
		reason = phaseSummary(lastPhase)
	}
	if lastExecute != nil &&
		lastExecute.Status == cyclesdomain.PhaseStatusSucceeded &&
		cycle.Status == cyclesdomain.CycleStatusFailed &&
		(lastPhase.Phase == cyclesdomain.PhaseVerify || strings.HasPrefix(reason, verificationFailedReason)) &&
		hasCommits {
		return EntryVerifyOnly
	}
	return EntryExecute
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parentFailureReason(phases []cyclesdomain.TaskCyclePhase, cycle *cyclesdomain.TaskCycle) string {
	if len(phases) > 0 {
		last := phases[len(phases)-1]
		if last.Status == cyclesdomain.PhaseStatusFailed {
			if s := phaseSummary(last); s != "" {
				return s
			}
		}
		for i := len(phases) - 1; i >= 0; i-- {
			if phases[i].Status == cyclesdomain.PhaseStatusFailed {
				if s := phaseSummary(phases[i]); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func continuationSufficient(bundle ContinuationBundle, cycle *cyclesdomain.TaskCycle) bool {
	if cycle == nil || cycle.ID == "" {
		return false
	}
	if len(bundle.PreviouslyPassed) > 0 || len(bundle.Commits) > 0 || len(bundle.CriteriaEvidence) > 0 {
		return true
	}
	if bundle.FailureReason != "" || bundle.FailurePhase != "" {
		return true
	}
	return cyclesdomain.TerminalCycleStatus(cycle.Status)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func lastExecutePhase(phases []cyclesdomain.TaskCyclePhase) *cyclesdomain.TaskCyclePhase {
	var last *cyclesdomain.TaskCyclePhase
	for i := range phases {
		p := &phases[i]
		if p.Phase != cyclesdomain.PhaseExecute {
			continue
		}
		if last == nil || p.PhaseSeq > last.PhaseSeq {
			last = p
		}
	}
	return last
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func phaseSummary(p cyclesdomain.TaskCyclePhase) string {
	if p.Summary == nil {
		return ""
	}
	return strings.TrimSpace(*p.Summary)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func runnerFeedbackFromPhase(p *cyclesdomain.TaskCyclePhase) string {
	if p == nil {
		return ""
	}
	summary := phaseSummary(*p)
	if summary == "" {
		return ""
	}
	if len(summary) > 512 {
		summary = summary[:512] + "…"
	}
	return summary
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func runnerDetailsExcerpt(details []byte) string {
	if len(details) == 0 {
		return ""
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(details, &root); err != nil {
		return ""
	}
	if raw, ok := root["summary"]; ok {
		var s string
		if json.Unmarshal(raw, &s) == nil && s != "" {
			if len(s) > 256 {
				s = s[:256] + "…"
			}
			return s
		}
	}
	return ""
}
