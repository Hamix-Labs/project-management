package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"log/slog"
	"strings"
)

func (s *Service) runVerifyChecks(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	phaseSeq int64,
	runCorrelationID string,
	attemptSeq int64,
	snap Snapshot,
	previouslyPassed map[string]Verdict,
	feedback string,
	mirrorDegradedIn bool,
) ([]Verdict, string, bool, cyclesdomain.TokenUsage, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyChecks",
		"task_id", task.ID, "cycle_id", cycle.ID,
		"run_correlation_id", runCorrelationID,
		"criteria_count", len(snap.Criteria), "previously_passed", len(previouslyPassed))
	mirrorDegraded := mirrorDegradedIn
	expected := make(map[string]struct{}, len(snap.Criteria))
	for _, it := range snap.Criteria {
		if _, locked := previouslyPassed[it.ID]; locked {
			continue
		}
		expected[it.ID] = struct{}{}
	}

	selfReport, err := s.loadCriteriaSelfReport(parentCtx, cycle.ID, attemptSeq, expected)
	if err != nil {
		return nil, "", mirrorDegraded, cyclesdomain.TokenUsage{}, false, err
	}

	if uerr := s.PersistCriteriaReports(parentCtx, cycle.ID, attemptSeq, snap.Criteria, previouslyPassed, selfReport); uerr != nil {
		mirrorDegraded = true
		slog.Warn("agent harness UpsertCriteriaReports failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyChecks.upsert_criteria_err",
			"cycle_id", cycle.ID, "attempt_seq", attemptSeq, "err", uerr)
	}

	verdicts := make([]Verdict, 0, len(snap.Criteria))
	needLLMVerify := false
	llmExpected := make(map[string]struct{})
	var usage cyclesdomain.TokenUsage
	var usagePresent bool

	for _, it := range snap.Criteria {
		if locked, ok := previouslyPassed[it.ID]; ok {
			verdicts = append(verdicts, locked)
			continue
		}
		entry := selfReport[it.ID]
		v := Verdict{
			ID:       it.ID,
			Evidence: entry.Evidence,
		}
		if !entry.ClaimedDone {
			v.Passed = false
			v.Verifier = checklistdomain.VerifierAgentSelf
			v.Reasoning = "execute agent did not claim criterion done"
			verdicts = append(verdicts, v)
			s.recordVerdict(checklistdomain.VerifierAgentSelf, false)
			continue
		}
		if len(it.VerifyCommands) == 0 {
			// ADR-0090: claim-only criteria accept from execute submit.
			v.Passed = true
			v.Verifier = checklistdomain.VerifierExecuteClaim
			v.Reasoning = "accepted execute claim (no verify commands)"
			verdicts = append(verdicts, v)
			s.recordVerdict(checklistdomain.VerifierExecuteClaim, true)
			continue
		}
		needLLMVerify = true
		llmExpected[it.ID] = struct{}{}
		verdicts = append(verdicts, v)
	}

	if needLLMVerify {
		cmdEvidence, cmdErr := s.RunCriterionCommands(parentCtx, task.ID, cycle.ID, phaseSeq, attemptSeq, snap, selfReport, nil)
		if cmdErr != nil {
			return nil, "", mirrorDegraded, cyclesdomain.TokenUsage{}, false, cmdErr
		}
		runUsage, runUsagePresent, runErr := s.runLLMVerify(parentCtx, task, cycle, phaseSeq, runCorrelationID, snap, previouslyPassed, selfReport, feedback, cmdEvidence, int(attemptSeq)-1)
		if runUsagePresent {
			usage = cyclesdomain.AddTokenUsage(usage, runUsage)
			usagePresent = true
		}
		nextVerdicts, parseErr := s.assembleVerdictsFromVerifyReport(cycle.ID, llmExpected, verdicts, selfReport, previouslyPassed)
		if err := verifyLLMRunError(runErr, parseErr); err != nil {
			return nil, "", mirrorDegraded, usage, usagePresent, err
		}
		verdicts = nextVerdicts
	}

	if uerr := s.persistVerifyReports(parentCtx, cycle.ID, attemptSeq, verdicts, previouslyPassed); uerr != nil {
		mirrorDegraded = true
		slog.Warn("agent harness UpsertVerifyReports failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyChecks.upsert_verify_err",
			"cycle_id", cycle.ID, "attempt_seq", attemptSeq, "err", uerr)
	}

	var failures []string
	for _, v := range verdicts {
		if !v.Passed {
			failures = append(failures, fmt.Sprintf("%s: %s", v.ID, v.Reasoning))
		}
	}
	if len(failures) > 0 {
		return verdicts, strings.Join(failures, "; "), mirrorDegraded, usage, usagePresent, fmt.Errorf("verification failed")
	}
	return verdicts, "", mirrorDegraded, usage, usagePresent, nil
}
