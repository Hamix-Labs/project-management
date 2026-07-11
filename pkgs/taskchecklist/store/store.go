// Package store implements GORM persistence for task checklists.
package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/internal/checklist"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for task checklists.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.NewStore")
	return &Store{db: db}
}

type (
	// ChecklistItemView is the per-task checklist row shape returned by List.
	ChecklistItemView = checklist.ItemView
	// ChecklistVerifyItem is a criterion row for worker verification.
	ChecklistVerifyItem = checklist.VerifyItem
	// CreateChecklistItemInput is one criterion seeded at task create.
	CreateChecklistItemInput = contract.CreateChecklistItemInput
	// VerifyCommandInput is the checklist verify command wire shape.
	VerifyCommandInput = contract.VerifyCommandInput
	// CriteriaFlagChange reports criteria_satisfied_at transitions inside a TX.
	CriteriaFlagChange = checklist.CriteriaFlagChange
)

// ValidateCanMarkDoneInTx rejects transitioning a task to done when checklist items remain open.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; checklist.ValidateCanMarkDoneInTx emits trace at the store chokepoint."
func ValidateCanMarkDoneInTx(tx *gorm.DB, taskID string) error {
	return checklist.ValidateCanMarkDoneInTx(tx, taskID)
}

// ValidateCanAddCriterionInTx rejects appending definition rows while status=running.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; checklist.ValidateCanAddCriterionInTx emits trace at the store chokepoint."
func ValidateCanAddCriterionInTx(tx *gorm.DB, t *domain.Task) error {
	return checklist.ValidateCanAddCriterionInTx(tx, t)
}

// DeleteOwnedItemsInTx removes checklist definition and completion rows owned by taskID.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; checklist.DeleteOwnedItemsInTx emits trace at the store chokepoint."
func DeleteOwnedItemsInTx(tx *gorm.DB, taskID string) error {
	return checklist.DeleteOwnedItemsInTx(tx, taskID)
}

// SeedDefinitionItemsAtCreateInTx inserts checklist definition rows during POST /tasks.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; checklist.SeedDefinitionItemsAtCreateInTx emits trace at the store chokepoint."
func SeedDefinitionItemsAtCreateInTx(tx *gorm.DB, taskID string, items []CreateChecklistItemInput, by domain.Actor) error {
	return checklist.SeedDefinitionItemsAtCreateInTx(tx, taskID, items, by)
}

// BackfillCriteriaSatisfiedAt sets criteria_satisfied_at for tasks whose checklist is already complete.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; checklist.BackfillCriteriaSatisfiedAt emits trace at the store chokepoint."
func BackfillCriteriaSatisfiedAt(ctx context.Context, db *gorm.DB) error {
	return checklist.BackfillCriteriaSatisfiedAt(ctx, db)
}

// NormalizeVerifyCommands validates optional verify command inputs.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NormalizeVerifyCommands(in []VerifyCommandInput) ([]VerifyCommandInput, error) {
	return checklist.NormalizeVerifyCommandInputs(in)
}

func (s *Store) DefinitionSourceTaskID(ctx context.Context, taskID string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.DefinitionSourceTaskID")
	return checklist.DefinitionSourceTaskID(ctx, s.db, taskID)
}

func (s *Store) ListChecklistForSubject(ctx context.Context, taskID string) ([]ChecklistItemView, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.ListChecklistForSubject")
	return checklist.List(ctx, s.db, taskID)
}

func (s *Store) AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []VerifyCommandInput, by domain.Actor) (*checklistdomain.TaskChecklistItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.AddChecklistItem")
	return checklist.Add(ctx, s.db, taskID, text, verifyCommands, by)
}

func (s *Store) ReplaceChecklistVerifyCommands(ctx context.Context, taskID, itemID string, cmds []VerifyCommandInput, by domain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.ReplaceChecklistVerifyCommands")
	return checklist.ReplaceVerifyCommands(ctx, s.db, taskID, itemID, cmds, by)
}

func (s *Store) ListChecklistForVerify(ctx context.Context, taskID string) ([]ChecklistVerifyItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.ListChecklistForVerify")
	return checklist.ListForVerify(ctx, s.db, taskID)
}

func (s *Store) IsTaskCycleRunning(ctx context.Context, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.IsTaskCycleRunning")
	return checklist.IsTaskCycleRunning(ctx, s.db, taskID)
}

func (s *Store) SetChecklistItemDoneWithEvidence(
	ctx context.Context,
	subjectTaskID, itemID string,
	evidence string,
	verifier checklistdomain.VerifierKind,
	reasoning, cycleID string,
	by domain.Actor,
) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.SetChecklistItemDoneWithEvidence")
	_, err := checklist.SetDoneWithEvidence(ctx, s.db, subjectTaskID, itemID, evidence, verifier, reasoning, cycleID, by)
	return err
}

// SetDoneWithEvidence records agent completion with proof metadata and reports
// whether criteria_satisfied_at transitioned inside the same transaction.
func (s *Store) SetDoneWithEvidence(
	ctx context.Context,
	subjectTaskID, itemID string,
	evidence string,
	verifier checklistdomain.VerifierKind,
	reasoning, cycleID string,
	by domain.Actor,
) (CriteriaFlagChange, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.SetDoneWithEvidence")
	return checklist.SetDoneWithEvidence(ctx, s.db, subjectTaskID, itemID, evidence, verifier, reasoning, cycleID, by)
}

func (s *Store) DeleteChecklistItem(ctx context.Context, taskID, itemID string, by domain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.DeleteChecklistItem")
	return checklist.Delete(ctx, s.db, taskID, itemID, by)
}

func (s *Store) UpdateChecklistItemText(ctx context.Context, taskID, itemID, text string, by domain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.UpdateChecklistItemText")
	return checklist.UpdateText(ctx, s.db, taskID, itemID, text, by)
}

func (s *Store) SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by domain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.store.SetChecklistItemDone")
	return checklist.SetDone(ctx, s.db, subjectTaskID, itemID, done, by)
}
