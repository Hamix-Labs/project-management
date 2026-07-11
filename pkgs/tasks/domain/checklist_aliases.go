package domain

import checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"

type (
	// TaskChecklistItem aliases taskchecklist/domain until taskcore (#186).
	TaskChecklistItem = checklistdomain.TaskChecklistItem
	// TaskChecklistItemCommand aliases taskchecklist/domain until taskcore (#186).
	TaskChecklistItemCommand = checklistdomain.TaskChecklistItemCommand
	// TaskChecklistCompletion aliases taskchecklist/domain until taskcore (#186).
	TaskChecklistCompletion = checklistdomain.TaskChecklistCompletion
	// VerifierKind aliases taskchecklist/domain until taskcore (#186).
	VerifierKind = checklistdomain.VerifierKind
)

const (
	MaxVerifyCommandsPerItem    = checklistdomain.MaxVerifyCommandsPerItem
	MaxVerifyCommandLen         = checklistdomain.MaxVerifyCommandLen
	MaxVerifyExpectedOutcomeLen = checklistdomain.MaxVerifyExpectedOutcomeLen

	VerifierAgentSelf          = checklistdomain.VerifierAgentSelf
	VerifierVerifyAgent        = checklistdomain.VerifierVerifyAgent
	VerifierDeterministicCheck = checklistdomain.VerifierDeterministicCheck
	VerifierHumanOverride      = checklistdomain.VerifierHumanOverride
	VerifierLegacy             = checklistdomain.VerifierLegacy
)

// ValidVerifierKind forwards to taskchecklist/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskchecklist/domain; operation trace is emitted by the BC chokepoint."
func ValidVerifierKind(k VerifierKind) bool {
	return checklistdomain.ValidVerifierKind(k)
}
