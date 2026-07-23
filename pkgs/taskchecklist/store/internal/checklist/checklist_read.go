package checklist

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/taskload"
	"gorm.io/gorm"
)

// ItemView is one definition row plus completion for a subject task.
// Re-aliased by the store facade so the
// JSON field tags stay stable on the wire.
type ItemView = contract.ChecklistItemView

// DefinitionSourceTaskID returns the task id that owns checklist item
// definitions for taskID. Walks the ParentID chain through any
// ChecklistInherit-true ancestors. Errors:
//   - ErrNotFound when the task or an ancestor is missing.
//   - ErrInvalidInput when an inherit-true task has no parent, or a
//     cycle in the parent chain is detected.
func DefinitionSourceTaskID(ctx context.Context, db *gorm.DB, taskID string) (string, error) {
	defer storekernel.DeferLatency(storekernel.OpDefinitionSourceTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.DefinitionSourceTaskID")
	return DefinitionSourceTaskIDInTx(db.WithContext(ctx), taskID)
}

// DefinitionSourceTaskIDInTx is the in-transaction variant used by
// other internal store packages that already hold a *gorm.DB tx
// handle.
func DefinitionSourceTaskIDInTx(tx *gorm.DB, taskID string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.DefinitionSourceTaskIDInTx")
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return "", fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	var n int64
	if err := tx.Model(&taskmodel.Task{}).Where("id = ?", taskID).Count(&n).Error; err != nil {
		return "", fmt.Errorf("load task: %w", err)
	}
	if n == 0 {
		return "", taskcoredomain.ErrNotFound
	}
	return taskID, nil
}

// List returns definition items for taskID with done flags for that
// same task. The taskID must exist; otherwise ErrNotFound.
func List(ctx context.Context, db *gorm.DB, taskID string) ([]ItemView, error) {
	defer storekernel.DeferLatency(storekernel.OpListChecklist)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.List")
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	var out []ItemView
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := taskload.LoadTask(tx, taskID); err != nil {
			return err
		}
		defID, err := DefinitionSourceTaskIDInTx(tx, taskID)
		if err != nil {
			return err
		}
		items, err := itemsForDefinitionInTx(tx, defID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			out = []ItemView{}
			return nil
		}
		ids := make([]string, len(items))
		for i := range items {
			ids[i] = items[i].ID
		}
		var doneRows []checklistmodel.TaskChecklistCompletion
		if err := tx.Where("task_id = ? AND item_id IN ?", taskID, ids).Find(&doneRows).Error; err != nil {
			return fmt.Errorf("list checklist completions: %w", err)
		}
		doneByItem := make(map[string]checklistdomain.TaskChecklistCompletion, len(doneRows))
		for _, d := range doneRows {
			dd := checklistmodel.ToDomainTaskChecklistCompletion(d)
			doneByItem[dd.ItemID] = dd
		}
		out = make([]ItemView, 0, len(items))
		cmdsByItem, err := commandsForItemsInTx(tx, ids)
		if err != nil {
			return err
		}
		for _, it := range items {
			v := ItemView{
				ID:             it.ID,
				SortOrder:      it.SortOrder,
				Text:           it.Text,
				VerifyCommands: cmdsByItem[it.ID],
			}
			if d, ok := doneByItem[it.ID]; ok {
				v.Done = true
				v.Evidence = d.Evidence
				v.VerifiedBy = string(d.VerifiedBy)
				v.VerifierReasoning = d.VerifierReasoning
				v.CycleID = d.CycleID
			}
			out = append(out, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
