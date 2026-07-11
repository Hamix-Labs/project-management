package orchestration

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// DecideVerifyDisabledLegacy maps legacy checklist completion outcome to effects.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideVerifyDisabledLegacy(checklistErr error) VerifyEffects {
	if checklistErr != nil {
		return VerifyEffects{TerminalFailure: true}
	}
	return VerifyEffects{}
}

// DecideFinalizeSuccess maps completion ledger outcome to terminal cycle/task status.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideFinalizeSuccess(completionErr error) FinalizeEffects {
	if completionErr != nil {
		return FinalizeEffects{
			CycleStatus: cyclesdomain.CycleStatusFailed,
			TaskStatus:  taskcoredomain.StatusFailed,
			Reason:      ReasonChecklistCompletionFailed,
		}
	}
	return FinalizeEffects{
		CycleStatus: cyclesdomain.CycleStatusSucceeded,
		TaskStatus:  taskcoredomain.StatusDone,
	}
}
