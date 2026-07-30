package harness

import (
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func checklistItemsForPrompt(items []checklistcontract.ChecklistVerifyItem) []prompt.ChecklistItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]prompt.ChecklistItem, len(items))
	for i, it := range items {
		cmds := make([]prompt.ChecklistCommand, 0, len(it.VerifyCommands))
		for _, c := range it.VerifyCommands {
			cmds = append(cmds, prompt.ChecklistCommand{
				Command:         c.Command,
				ExpectedOutcome: c.ExpectedOutcome,
			})
		}
		out[i] = prompt.ChecklistItem{ID: it.ID, Text: it.Text, Commands: cmds}
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func lockedPassIDSet(lockedPasses map[string]criterionVerdict) map[string]struct{} {
	if len(lockedPasses) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(lockedPasses))
	for id := range lockedPasses {
		out[id] = struct{}{}
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func continuationInputFromBundle(cycle *cyclesdomain.TaskCycle, bundle *ContinuationBundle) prompt.ContinuationInput {
	if bundle == nil {
		return prompt.ContinuationInput{Cycle: cycle}
	}
	return prompt.ContinuationInput{
		LineageAttempt:  bundle.LineageAttempt,
		Cycle:           cycle,
		FailureClass:    string(bundle.FailureClass),
		FailureReason:   bundle.FailureReason,
		FailurePhase:    bundle.FailurePhase,
		ScopeFiles:      bundle.ScopeFiles,
		Commits:         bundle.Commits,
		ExecuteFeedback: bundle.ExecuteFeedback,
		RunnerFeedback:  bundle.RunnerFeedback,
		GitDiagnostics:  bundle.GitDiagnostics,
		Warnings:        bundle.Warnings,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func cycleIDOrEmpty(cycle *cyclesdomain.TaskCycle) string {
	if cycle == nil {
		return ""
	}
	return cycle.ID
}
