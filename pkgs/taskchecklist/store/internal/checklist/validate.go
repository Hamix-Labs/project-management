package checklist

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// itemsForDefinitionInTx returns the canonical-ordered definition rows
// for the task that owns them (must already be the resolved definition
// owner; not the inherit-true subject).
func itemsForDefinitionInTx(tx *gorm.DB, defTaskID string) ([]checklistdomain.TaskChecklistItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.itemsForDefinitionInTx")
	var items []checklistmodel.TaskChecklistItem
	if err := tx.Where("task_id = ?", defTaskID).Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return checklistmodel.ToDomainTaskChecklistItems(items), nil
}

// validateChecklistCompleteInTx asserts that every definition item
// inherited by subjectTaskID has a matching task_checklist_completions
// row for the same subject task. Empty checklist == OK. Surfaces
// ErrInvalidInput when at least one item is unchecked, so the caller
// can surface a 400 to the API client.
func validateChecklistCompleteInTx(tx *gorm.DB, subjectTaskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.validateChecklistCompleteInTx")
	defID, err := DefinitionSourceTaskIDInTx(tx, subjectTaskID)
	if err != nil {
		return err
	}
	items, err := itemsForDefinitionInTx(tx, defID)
	if err != nil {
		return fmt.Errorf("checklist: %w", err)
	}
	if len(items) == 0 {
		return nil
	}
	for _, it := range items {
		var comp checklistmodel.TaskChecklistCompletion
		err := tx.Where("task_id = ? AND item_id = ?", subjectTaskID, it.ID).First(&comp).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: all checklist items must be done before marking this task done", taskcoredomain.ErrInvalidInput)
		}
		if err != nil {
			return fmt.Errorf("checklist completion: %w", err)
		}
		dcomp := checklistmodel.ToDomainTaskChecklistCompletion(comp)
		if !checklistdomain.ValidVerifierKind(dcomp.VerifiedBy) {
			return fmt.Errorf("%w: checklist completion missing verified_by", taskcoredomain.ErrInvalidInput)
		}
		if dcomp.VerifiedBy != checklistdomain.VerifierLegacy && strings.TrimSpace(dcomp.Evidence) == "" {
			return fmt.Errorf("%w: checklist completion missing evidence", taskcoredomain.ErrInvalidInput)
		}
	}
	return nil
}

// ValidateCanMarkDoneInTx is the cross-domain guard that the task
// CRUD/update/devmirror code calls before transitioning a task to
// status=done. Requires checklist complete.
func ValidateCanMarkDoneInTx(tx *gorm.DB, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.ValidateCanMarkDoneInTx")
	return validateChecklistCompleteInTx(tx, taskID)
}

// ValidateCanAddCriterionInTx rejects appending definition rows while the
// agent is actively working the task (status=running).
func ValidateCanAddCriterionInTx(tx *gorm.DB, t *taskcoredomain.Task) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.ValidateCanAddCriterionInTx")
	return validateCriteriaMutable(t)
}

// ValidateCriteriaMutable rejects user-driven checklist mutations while the
// task is in progress. Done tasks remain editable for post-completion tweaks.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidateCriteriaMutable(t *taskcoredomain.Task) error {
	return validateCriteriaMutable(t)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateCriteriaMutable(t *taskcoredomain.Task) error {
	if t == nil {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if t.Status == taskcoredomain.StatusRunning {
		return fmt.Errorf("%w: cannot change criteria while task is running", taskcoredomain.ErrConflict)
	}
	return nil
}

// criterionLockedByCompletion reports whether existing completion rows block
// definition edits. Satisfied criteria stay locked while the task is still
// active; once status=done operators may revise definitions.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func criterionLockedByCompletion(taskStatus taskcoredomain.Status, doneCount int64) bool {
	if doneCount == 0 {
		return false
	}
	return taskStatus != taskcoredomain.StatusDone
}

// DeleteOwnedItemsInTx removes every checklist definition row owned
// by taskID and the per-subject completion rows that point at those
// items. Exported so the task update/delete paths can drop a task's
// checklist atomically alongside the parent row.
func DeleteOwnedItemsInTx(tx *gorm.DB, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.DeleteOwnedItemsInTx")
	var ids []string
	if err := tx.Model(&checklistmodel.TaskChecklistItem{}).Where("task_id = ?", taskID).Pluck("id", &ids).Error; err != nil {
		return fmt.Errorf("list checklist items: %w", err)
	}
	if len(ids) == 0 {
		return nil
	}
	if err := tx.Where("item_id IN ?", ids).Delete(&checklistmodel.TaskChecklistCompletion{}).Error; err != nil {
		return fmt.Errorf("delete completions: %w", err)
	}
	if err := tx.Where("task_id = ?", taskID).Delete(&checklistmodel.TaskChecklistItem{}).Error; err != nil {
		return fmt.Errorf("delete checklist items: %w", err)
	}
	return nil
}
