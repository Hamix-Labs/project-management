package contract

import cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"

type (
	// StartCycleInput captures everything needed to begin a new execution attempt.
	StartCycleInput = cyclescontract.StartCycleInput
	// CompletePhaseInput captures the terminal transition for a phase row.
	CompletePhaseInput = cyclescontract.CompletePhaseInput
)
