package contract

import "context"

// ChecklistVerifyItem is a criterion row for worker verification.
type ChecklistVerifyItem struct {
	ID             string
	Text           string
	SourceTaskID   string
	DefinitionTask string
	VerifyCommands []VerifyCommandView
}

// ChecklistHarnessStore extends ChecklistStore with harness verify reads.
type ChecklistHarnessStore interface {
	ChecklistStore
	ListChecklistForVerify(ctx context.Context, taskID string) ([]ChecklistVerifyItem, error)
}
