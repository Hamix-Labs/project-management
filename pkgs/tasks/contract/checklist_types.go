package contract

import checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"

type (
	// ChecklistItemView is one definition row plus completion for a subject task.
	ChecklistItemView = checklistcontract.ChecklistItemView
	// CreateChecklistItemInput is one criterion seeded at task create.
	CreateChecklistItemInput = checklistcontract.CreateChecklistItemInput
	// VerifyCommandInput is a verify command on checklist create/update.
	VerifyCommandInput = checklistcontract.VerifyCommandInput
	// VerifyCommandView is a persisted command row on checklist API responses.
	VerifyCommandView = checklistcontract.VerifyCommandView
)
