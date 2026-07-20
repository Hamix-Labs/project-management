package contract

import (
	"context"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// SetDoneWithEvidenceInput records checklist completion with proof metadata.
type SetDoneWithEvidenceInput struct {
	TaskID    string
	ItemID    string
	Evidence  string
	Verifier  checklistdomain.VerifierKind
	Reasoning string
	CycleID   string
	By        taskcoredomain.Actor
}

// ChecklistStore covers checklist CRUD and completion for a task subject.
type ChecklistStore interface {
	ListChecklistForSubject(ctx context.Context, taskID string) ([]ChecklistItemView, error)
	AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []VerifyCommandInput, by taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error)
	UpdateChecklistItemText(ctx context.Context, taskID, itemID, text string, by taskcoredomain.Actor) error
	ReplaceChecklistVerifyCommands(ctx context.Context, taskID, itemID string, cmds []VerifyCommandInput, by taskcoredomain.Actor) error
	SetChecklistItemDoneWithEvidence(ctx context.Context, in SetDoneWithEvidenceInput) error
	SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by taskcoredomain.Actor) error
	DeleteChecklistItem(ctx context.Context, taskID, itemID string, by taskcoredomain.Actor) error
}
