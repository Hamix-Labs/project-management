package contract

import (
	"context"

	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// ChecklistStore covers harness checklist mutations and verify reads.
type ChecklistStore interface {
	AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []checklistcontract.VerifyCommandInput, by taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error)
	ListChecklistForVerify(ctx context.Context, taskID string) ([]checkliststore.ChecklistVerifyItem, error)
	ListChecklistForSubject(ctx context.Context, taskID string) ([]checklistcontract.ChecklistItemView, error)
	SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by taskcoredomain.Actor) error
	SetChecklistItemDoneWithEvidence(ctx context.Context, subjectTaskID, itemID string, evidence string, verifier checklistdomain.VerifierKind, reasoning, cycleID string, by taskcoredomain.Actor) error
}
