package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// ChecklistStore covers checklist CRUD and completion for a task subject.
type ChecklistStore interface {
	ListChecklistForSubject(ctx context.Context, taskID string) ([]ChecklistItemView, error)
	IsTaskCycleRunning(ctx context.Context, taskID string) (bool, error)
	AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []VerifyCommandInput, by domain.Actor) (*domain.TaskChecklistItem, error)
	UpdateChecklistItemText(ctx context.Context, taskID, itemID, text string, by domain.Actor) error
	ReplaceChecklistVerifyCommands(ctx context.Context, taskID, itemID string, cmds []VerifyCommandInput, by domain.Actor) error
	SetChecklistItemDoneWithEvidence(ctx context.Context, subjectTaskID, itemID string, evidence string, verifier domain.VerifierKind, reasoning, cycleID string, by domain.Actor) error
	SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by domain.Actor) error
	DeleteChecklistItem(ctx context.Context, taskID, itemID string, by domain.Actor) error
}
