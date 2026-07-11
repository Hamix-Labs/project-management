package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// BackfillCriteriaSatisfiedAt sets criteria_satisfied_at for tasks whose
// checklist is already complete. Idempotent migration helper.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func BackfillCriteriaSatisfiedAt(ctx context.Context, db *gorm.DB) error {
	return checkliststore.BackfillCriteriaSatisfiedAt(ctx, db)
}

type (
	// ChecklistItemView is the public re-export of the per-task checklist row shape.
	ChecklistItemView = checkliststore.ChecklistItemView
	// ChecklistVerifyItem is a criterion row for worker verification.
	ChecklistVerifyItem = checkliststore.ChecklistVerifyItem
	// CreateChecklistItemInput is the public re-export for task-create checklist rows.
	CreateChecklistItemInput = checklistcontract.CreateChecklistItemInput
	// VerifyCommandInput is the public re-export for checklist verify command wire shape.
	VerifyCommandInput = checklistcontract.VerifyCommandInput
)

// DefinitionSourceTaskID returns the task id that owns checklist item definitions.
func (s *Store) DefinitionSourceTaskID(ctx context.Context, taskID string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DefinitionSourceTaskID")
	return s.checklist.DefinitionSourceTaskID(ctx, taskID)
}

// ListChecklistForSubject returns definition items for taskID with done flags for that same task.
func (s *Store) ListChecklistForSubject(ctx context.Context, taskID string) ([]ChecklistItemView, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListChecklistForSubject")
	return s.checklist.ListChecklistForSubject(ctx, taskID)
}

// AddChecklistItem appends a definition row when the task is not running or done.
func (s *Store) AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []VerifyCommandInput, by taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AddChecklistItem")
	return s.checklist.AddChecklistItem(ctx, taskID, text, verifyCommands, by)
}

// ReplaceChecklistVerifyCommands replaces optional verify commands on a criterion.
func (s *Store) ReplaceChecklistVerifyCommands(ctx context.Context, taskID, itemID string, cmds []VerifyCommandInput, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReplaceChecklistVerifyCommands")
	return s.checklist.ReplaceChecklistVerifyCommands(ctx, taskID, itemID, cmds, by)
}

// NormalizeVerifyCommands validates optional verify command inputs.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NormalizeVerifyCommands(in []VerifyCommandInput) ([]VerifyCommandInput, error) {
	return checkliststore.NormalizeVerifyCommands(in)
}

// ListChecklistForVerify returns criteria rows for worker verification.
func (s *Store) ListChecklistForVerify(ctx context.Context, taskID string) ([]ChecklistVerifyItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListChecklistForVerify")
	return s.checklist.ListChecklistForVerify(ctx, taskID)
}

// IsTaskCycleRunning reports whether the task or an inherit ancestor has a running cycle.
func (s *Store) IsTaskCycleRunning(ctx context.Context, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.IsTaskCycleRunning")
	return s.checklist.IsTaskCycleRunning(ctx, taskID)
}

// SetChecklistItemDoneWithEvidence records agent completion with proof metadata.
func (s *Store) SetChecklistItemDoneWithEvidence(
	ctx context.Context,
	subjectTaskID, itemID string,
	evidence string,
	verifier checklistdomain.VerifierKind,
	reasoning, cycleID string,
	by taskcoredomain.Actor,
) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetChecklistItemDoneWithEvidence")
	flag, err := s.checklist.SetDoneWithEvidence(ctx, subjectTaskID, itemID, evidence, verifier, reasoning, cycleID, by)
	if err != nil {
		return err
	}
	if flag.BecameComplete {
		s.notifyUnblockedDependents(ctx, subjectTaskID)
	}
	return nil
}

// DeleteChecklistItem removes a definition row owned by taskID.
func (s *Store) DeleteChecklistItem(ctx context.Context, taskID, itemID string, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteChecklistItem")
	return s.checklist.DeleteChecklistItem(ctx, taskID, itemID, by)
}

// UpdateChecklistItemText updates the definition text for an item owned by taskID.
func (s *Store) UpdateChecklistItemText(ctx context.Context, taskID, itemID, text string, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateChecklistItemText")
	return s.checklist.UpdateChecklistItemText(ctx, taskID, itemID, text, by)
}

// SetChecklistItemDone sets or clears completion for subjectTaskID on an item from its definition source.
func (s *Store) SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetChecklistItemDone")
	before, _ := s.Get(ctx, subjectTaskID)
	if err := s.checklist.SetChecklistItemDone(ctx, subjectTaskID, itemID, done, by); err != nil {
		return err
	}
	after, _ := s.Get(ctx, subjectTaskID)
	if before != nil && after != nil && before.CriteriaSatisfiedAt == nil && after.CriteriaSatisfiedAt != nil {
		s.notifyUnblockedDependents(ctx, subjectTaskID)
	}
	return nil
}
