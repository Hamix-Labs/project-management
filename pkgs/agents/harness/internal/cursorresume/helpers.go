package cursorresume

import (
	"sort"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// FirstVerifyAfterNewExecute reports whether verify should force fresh after a
// newly completed execute phase (ADR-0031 / different_chat).
//
//funclogmeasure:skip category=hot-path reason="Pure state comparison for verify fresh-after-execute deny."
func FirstVerifyAfterNewExecute(lastVerifyAfterExecuteSeq, lastCompletedExecutePhaseSeq int64) bool {
	return lastVerifyAfterExecuteSeq < lastCompletedExecutePhaseSeq
}

// RecoveryKindInput is the pure input to SelectRecoveryKind.
type RecoveryKindInput struct {
	Phase             cyclesdomain.Phase
	ReportParseErr    string
	RetryMode         taskcoredomain.RetryMode
	RunKind           taskcoredomain.PendingRunKind
	HasContinuation   bool
	ResumeNotice      bool
	HasFailedVerdicts bool
}

// SelectRecoveryKind chooses the structured delta template for a resumed session.
//
//funclogmeasure:skip category=hot-path reason="Pure kind selection from pre-resolved facts."
func SelectRecoveryKind(in RecoveryKindInput) prompt.RecoveryKind {
	if in.Phase == cyclesdomain.PhaseVerify {
		if in.HasFailedVerdicts {
			return prompt.RecoveryVerifyImplementation
		}
		return prompt.RecoveryVerifyInfra
	}
	if in.ReportParseErr != "" {
		if strings.Contains(strings.ToLower(in.ReportParseErr), "missing") {
			return prompt.RecoveryCriteriaReportMissing
		}
		return prompt.RecoveryCriteriaReportInvalid
	}
	if in.RetryMode == taskcoredomain.RetryResume && in.HasContinuation {
		return prompt.RecoveryOperatorRetryResume
	}
	if in.RunKind == taskcoredomain.PendingKindPolish {
		return prompt.RecoveryHumanPolish
	}
	if in.ResumeNotice {
		return prompt.RecoveryProcessRestart
	}
	if in.HasFailedVerdicts {
		return prompt.RecoveryVerifyImplementation
	}
	return prompt.RecoveryVerifyImplementation
}

// CriterionFailureInput is one failed criterion for recovery DTO mapping.
type CriterionFailureInput struct {
	ID        string
	Passed    bool
	Reasoning string
	Verifier  string
}

// FailedCriteriaFromInputs maps failed criteria into prompt recovery DTOs.
//
//funclogmeasure:skip category=hot-path reason="Pure verdict to DTO mapping."
func FailedCriteriaFromInputs(verdicts []CriterionFailureInput) []prompt.CriterionFailure {
	var out []prompt.CriterionFailure
	for _, v := range verdicts {
		if v.Passed {
			continue
		}
		out = append(out, prompt.CriterionFailure{
			ID:        v.ID,
			Reasoning: v.Reasoning,
			Verifier:  v.Verifier,
		})
	}
	return out
}

// LockedCriterionIDs returns sorted IDs from a locked-pass set.
//
//funclogmeasure:skip category=hot-path reason="Pure id extraction from locked verdict map."
func LockedCriterionIDs(locked map[string]struct{}) []string {
	if len(locked) == 0 {
		return nil
	}
	ids := make([]string, 0, len(locked))
	for id := range locked {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ActiveCriterionIDs returns sorted checklist IDs not yet locked as passed.
//
//funclogmeasure:skip category=hot-path reason="Pure active checklist id list from facts."
func ActiveCriterionIDs(allIDs []string, lockedPasses map[string]struct{}) []string {
	expected := make([]string, 0)
	for _, id := range allIDs {
		if _, ok := lockedPasses[id]; ok {
			continue
		}
		expected = append(expected, id)
	}
	sort.Strings(expected)
	return expected
}
