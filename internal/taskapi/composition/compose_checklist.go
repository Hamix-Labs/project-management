package composition

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

// DefinitionSourceTaskID returns the task id that owns checklist item definitions.
func (a *API) DefinitionSourceTaskID(ctx context.Context, taskID string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DefinitionSourceTaskID")
	return a.checklist.DefinitionSourceTaskID(ctx, taskID)
}

// ListChecklistForSubject returns definition items for taskID with done flags for that same task.
func (a *API) ListChecklistForSubject(ctx context.Context, taskID string) ([]checkliststore.ChecklistItemView, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListChecklistForSubject")
	return a.checklist.ListChecklistForSubject(ctx, taskID)
}

// AddChecklistItem appends a definition row when the task is not running or done.
func (a *API) AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []checklistcontract.VerifyCommandInput, by taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AddChecklistItem")
	return a.checklist.AddChecklistItem(ctx, taskID, text, verifyCommands, by)
}

// ReplaceChecklistVerifyCommands replaces optional verify commands on a criterion.
func (a *API) ReplaceChecklistVerifyCommands(ctx context.Context, taskID, itemID string, cmds []checklistcontract.VerifyCommandInput, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReplaceChecklistVerifyCommands")
	return a.checklist.ReplaceChecklistVerifyCommands(ctx, taskID, itemID, cmds, by)
}

// NormalizeVerifyCommands validates optional verify command inputs.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NormalizeVerifyCommands(in []checklistcontract.VerifyCommandInput) ([]checklistcontract.VerifyCommandInput, error) {
	return checkliststore.NormalizeVerifyCommands(in)
}

// ListChecklistForVerify returns criteria rows for worker verification.
func (a *API) ListChecklistForVerify(ctx context.Context, taskID string) ([]checkliststore.ChecklistVerifyItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListChecklistForVerify")
	return a.checklist.ListChecklistForVerify(ctx, taskID)
}

// IsTaskCycleRunning reports whether the task or an inherit ancestor has a running cycle.
func (a *API) IsTaskCycleRunning(ctx context.Context, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.IsTaskCycleRunning")
	return a.checklist.IsTaskCycleRunning(ctx, taskID)
}

// SetChecklistItemDoneWithEvidence records agent completion with proof metadata.
func (a *API) SetChecklistItemDoneWithEvidence(
	ctx context.Context,
	subjectTaskID, itemID string,
	evidence string,
	verifier checklistdomain.VerifierKind,
	reasoning, cycleID string,
	by taskcoredomain.Actor,
) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetChecklistItemDoneWithEvidence")
	flag, err := a.checklist.SetDoneWithEvidence(ctx, subjectTaskID, itemID, evidence, verifier, reasoning, cycleID, by)
	if err != nil {
		return err
	}
	if flag.BecameComplete {
		a.notifyUnblockedDependents(ctx, subjectTaskID)
	}
	return nil
}

// DeleteChecklistItem removes a definition row owned by taskID.
func (a *API) DeleteChecklistItem(ctx context.Context, taskID, itemID string, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteChecklistItem")
	return a.checklist.DeleteChecklistItem(ctx, taskID, itemID, by)
}

// UpdateChecklistItemText updates the definition text for an item owned by taskID.
func (a *API) UpdateChecklistItemText(ctx context.Context, taskID, itemID, text string, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateChecklistItemText")
	return a.checklist.UpdateChecklistItemText(ctx, taskID, itemID, text, by)
}

// SetChecklistItemDone sets or clears completion for subjectTaskID on an item from its definition source.
func (a *API) SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetChecklistItemDone")
	before, _ := a.taskcore.Get(ctx, subjectTaskID)
	if err := a.checklist.SetChecklistItemDone(ctx, subjectTaskID, itemID, done, by); err != nil {
		return err
	}
	after, _ := a.taskcore.Get(ctx, subjectTaskID)
	if before != nil && after != nil && before.CriteriaSatisfiedAt == nil && after.CriteriaSatisfiedAt != nil {
		a.notifyUnblockedDependents(ctx, subjectTaskID)
	}
	return nil
}
