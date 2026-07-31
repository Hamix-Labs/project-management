package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"log/slog"
)

func (s *Service) runVerifyChecks(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	phaseSeq int64,
	runCorrelationID string,
	attemptSeq int64,
	snap Snapshot,
	lockedPasses map[string]Verdict,
	mirrorDegradedIn bool,
) ([]Verdict, bool, cyclesdomain.TokenUsage, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyChecks",
		"task_id", task.ID, "cycle_id", cycle.ID,
		"run_correlation_id", runCorrelationID,
		"criteria_count", len(snap.Criteria), "locked_passes", len(lockedPasses))
	_ = phaseSeq
	mirrorDegraded := mirrorDegradedIn
	expected := make(map[string]struct{}, len(snap.Criteria))
	for _, it := range snap.Criteria {
		if _, locked := lockedPasses[it.ID]; locked {
			continue
		}
		expected[it.ID] = struct{}{}
	}

	selfReport, err := s.loadCriteriaSelfReport(parentCtx, cycle.ID, attemptSeq, expected)
	if err != nil {
		return nil, mirrorDegraded, cyclesdomain.TokenUsage{}, false, err
	}

	if uerr := s.PersistCriteriaReports(parentCtx, cycle.ID, attemptSeq, snap.Criteria, lockedPasses, selfReport); uerr != nil {
		mirrorDegraded = true
		slog.Warn("agent harness UpsertCriteriaReports failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyChecks.upsert_criteria_err",
			"cycle_id", cycle.ID, "attempt_seq", attemptSeq, "err", uerr)
	}

	verdicts := make([]Verdict, 0, len(snap.Criteria))
	for _, it := range snap.Criteria {
		if locked, ok := lockedPasses[it.ID]; ok {
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
		v.Passed = true
		v.Verifier = checklistdomain.VerifierExecuteClaim
		if len(it.VerifyCommands) > 0 {
			v.Reasoning = "accepted execute claim (agent self-checked verify commands)"
		} else {
			v.Reasoning = "accepted execute claim (no verify commands)"
		}
		verdicts = append(verdicts, v)
		s.recordVerdict(checklistdomain.VerifierExecuteClaim, true)
	}

	if uerr := s.persistVerifyReports(parentCtx, cycle.ID, attemptSeq, verdicts, lockedPasses); uerr != nil {
		mirrorDegraded = true
		slog.Warn("agent harness UpsertVerifyReports failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.runVerifyChecks.upsert_verify_err",
			"cycle_id", cycle.ID, "attempt_seq", attemptSeq, "err", uerr)
	}

	for _, v := range verdicts {
		if !v.Passed {
			return verdicts, mirrorDegraded, cyclesdomain.TokenUsage{}, false, fmt.Errorf("verification failed")
		}
	}
	return verdicts, mirrorDegraded, cyclesdomain.TokenUsage{}, false, nil
}
