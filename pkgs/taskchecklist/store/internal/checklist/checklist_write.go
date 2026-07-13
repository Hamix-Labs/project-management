package checklist

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/taskload"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Add appends a definition row; rejected while status=running. Appends
// EventChecklistItemAdded in the same TX.
func Add(ctx context.Context, db *gorm.DB, taskID, text string, verifyCommands []VerifyCommandInput, by taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error) {
	defer storekernel.DeferLatency(storekernel.OpAddChecklistItem)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.Add")
	if err := storekernel.ValidateActor(by); err != nil {
		return nil, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("%w: checklist text required", taskcoredomain.ErrInvalidInput)
	}
	cmds, err := NormalizeVerifyCommandInputs(verifyCommands)
	if err != nil {
		return nil, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	var created *checklistdomain.TaskChecklistItem
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		t, err := taskload.LoadTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ValidateCriteriaMutable(t); err != nil {
			return err
		}
		var maxOrder int
		row := tx.Model(&checklistmodel.TaskChecklistItem{}).Select("COALESCE(MAX(sort_order), 0)").Where("task_id = ?", taskID)
		if err := row.Scan(&maxOrder).Error; err != nil {
			return fmt.Errorf("checklist order: %w", err)
		}
		dit := checklistdomain.TaskChecklistItem{
			ID:        uuid.NewString(),
			TaskID:    taskID,
			SortOrder: maxOrder + 1,
			Text:      text,
		}
		if err := tx.Create(checklistmodel.FromDomainTaskChecklistItemPtr(&dit)).Error; err != nil {
			return fmt.Errorf("insert checklist item: %w", err)
		}
		if err := replaceCommandsInTx(tx, dit.ID, cmds); err != nil {
			return err
		}
		seq, err := storekernel.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(map[string]string{"item_id": dit.ID, "text": dit.Text})
		if err := storekernel.AppendEvent(tx, taskID, seq, taskeventsdomain.EventChecklistItemAdded, by, b); err != nil {
			return err
		}
		created = &dit
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("add checklist item: %w", err)
	}
	return created, nil
}

// Delete removes a definition row owned by taskID. Cascades to the
// per-subject completion rows for that item. Appends
// EventChecklistItemRemoved in the same TX.
func Delete(ctx context.Context, db *gorm.DB, taskID, itemID string, by taskcoredomain.Actor) error {
	defer storekernel.DeferLatency(storekernel.OpDeleteChecklistItem)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.Delete")
	if err := storekernel.ValidateActor(by); err != nil {
		return err
	}
	taskID = strings.TrimSpace(taskID)
	itemID = strings.TrimSpace(itemID)
	if taskID == "" || itemID == "" {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		t, err := taskload.LoadTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ValidateCriteriaMutable(t); err != nil {
			return err
		}
		var it checklistmodel.TaskChecklistItem
		if err := tx.Where("id = ? AND task_id = ?", itemID, taskID).First(&it).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return taskcoredomain.ErrNotFound
			}
			return fmt.Errorf("load checklist item: %w", err)
		}
		dit := checklistmodel.ToDomainTaskChecklistItem(it)
		// Symmetric with UpdateText above: a done criterion records
		// what was actually accepted as satisfied, and the
		// EventChecklistItemToggled (done=true) audit row already
		// references this item id. Removing the definition would
		// orphan that toggle event — anyone replaying the timeline
		// would see "checklist item X marked done" with no record of
		// what X was, and the per-subject completion row would have
		// to be silently cascaded away, erasing the historical fact
		// that the subject task ever satisfied this requirement.
		// Reject at the source of truth so every client (UI, CLI,
		// future API consumers) gets the same answer; the SPA also
		// disables its Remove button for done items, but that's
		// defence-in-depth.
		//
		// As with UpdateText, completion lives on
		// TaskChecklistCompletion (per-subject), not on the item
		// itself, so we count across *all* subjects: if any
		// subject — including an inheriting child — has marked the
		// criterion done, the audit-orphaning concern applies.
		var doneCount int64
		if err := tx.Model(&checklistmodel.TaskChecklistCompletion{}).
			Where("item_id = ?", itemID).
			Count(&doneCount).Error; err != nil {
			return fmt.Errorf("count completions: %w", err)
		}
		if criterionLockedByCompletion(t.Status, doneCount) {
			return fmt.Errorf("%w: cannot remove a criterion that has already been marked done", taskcoredomain.ErrInvalidInput)
		}
		if doneCount > 0 {
			if err := tx.Where("item_id = ?", itemID).Delete(&checklistmodel.TaskChecklistCompletion{}).Error; err != nil {
				return fmt.Errorf("delete checklist completions: %w", err)
			}
		}
		seq, err := storekernel.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(map[string]string{"item_id": itemID, "text": dit.Text})
		if err := storekernel.AppendEvent(tx, taskID, seq, taskeventsdomain.EventChecklistItemRemoved, by, b); err != nil {
			return err
		}
		if err := tx.Delete(&it).Error; err != nil {
			return fmt.Errorf("delete checklist item: %w", err)
		}
		return nil
	})
}

// UpdateText updates the definition text for an item owned by taskID.
// No-op (no event emitted) when the new text matches the existing
// row, so idempotent UI saves do not pollute the audit log. Appends
// EventChecklistItemUpdated in the same TX otherwise.
func UpdateText(ctx context.Context, db *gorm.DB, taskID, itemID, text string, by taskcoredomain.Actor) error {
	defer storekernel.DeferLatency(storekernel.OpUpdateChecklistItemText)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.UpdateText")
	if err := storekernel.ValidateActor(by); err != nil {
		return err
	}
	taskID = strings.TrimSpace(taskID)
	itemID = strings.TrimSpace(itemID)
	text = strings.TrimSpace(text)
	if taskID == "" || itemID == "" || text == "" {
		return fmt.Errorf("%w: text", taskcoredomain.ErrInvalidInput)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		t, err := taskload.LoadTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ValidateCriteriaMutable(t); err != nil {
			return err
		}
		var it checklistmodel.TaskChecklistItem
		if err := tx.Where("id = ? AND task_id = ?", itemID, taskID).First(&it).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return taskcoredomain.ErrNotFound
			}
			return fmt.Errorf("load checklist item: %w", err)
		}
		dit := checklistmodel.ToDomainTaskChecklistItem(it)
		// A done criterion records what was actually accepted as
		// satisfied. Letting the text be retroactively rewritten
		// would silently change the meaning of the already-emitted
		// EventChecklistItemToggled (done=true) audit row — anyone
		// replaying the timeline would see "done" against definition
		// text that never existed at completion time. Reject the
		// edit at the source of truth so any client (UI, CLI, etc.)
		// gets the same answer; the SPA also disables its Edit
		// button for done items, but that's defence-in-depth.
		//
		// `Done` lives on the per-subject TaskChecklistCompletion
		// row, not on the item itself (this is what lets inheriting
		// child tasks track completion independently while sharing
		// one definition). We reject if *any* subject has marked
		// the criterion done — even the inherit case where the
		// definition is owned here but the completion lives on a
		// child — because the audit-rewriting concern is symmetric
		// across every subject that already accepted the text.
		var doneCount int64
		if err := tx.Model(&checklistmodel.TaskChecklistCompletion{}).
			Where("item_id = ?", itemID).
			Count(&doneCount).Error; err != nil {
			return fmt.Errorf("count completions: %w", err)
		}
		if criterionLockedByCompletion(t.Status, doneCount) {
			return fmt.Errorf("%w: cannot edit a criterion that has already been marked done", taskcoredomain.ErrInvalidInput)
		}
		if dit.Text == text {
			return nil
		}
		if err := tx.Model(&it).Update("text", text).Error; err != nil {
			return fmt.Errorf("update checklist item: %w", err)
		}
		seq, err := storekernel.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(map[string]any{"item_id": itemID, "text": text})
		return storekernel.AppendEvent(tx, taskID, seq, taskeventsdomain.EventChecklistItemUpdated, by, b)
	})
}

// SetDone sets or clears completion for subjectTaskID on an item
// resolved through DefinitionSourceTaskIDInTx. Only taskcoredomain.ActorAgent
// may change completion; the human user records criteria via Add but
// does not toggle done. Appends EventChecklistItemToggled in the same
// TX.
func SetDone(ctx context.Context, db *gorm.DB, subjectTaskID, itemID string, done bool, by taskcoredomain.Actor) error {
	defer storekernel.DeferLatency(storekernel.OpSetChecklistItemDone)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.SetDone")
	if err := storekernel.ValidateActor(by); err != nil {
		return err
	}
	if by != taskcoredomain.ActorAgent {
		return fmt.Errorf("%w: only the agent may mark checklist items done or undone", taskcoredomain.ErrInvalidInput)
	}
	subjectTaskID = strings.TrimSpace(subjectTaskID)
	itemID = strings.TrimSpace(itemID)
	if subjectTaskID == "" || itemID == "" {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := taskload.LoadTask(tx, subjectTaskID); err != nil {
			return err
		}
		defOwner, err := DefinitionSourceTaskIDInTx(tx, subjectTaskID)
		if err != nil {
			return err
		}
		var it checklistmodel.TaskChecklistItem
		if err := tx.Where("id = ? AND task_id = ?", itemID, defOwner).First(&it).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return taskcoredomain.ErrNotFound
			}
			return fmt.Errorf("load checklist item: %w", err)
		}
		var existing checklistmodel.TaskChecklistCompletion
		err = tx.Where("task_id = ? AND item_id = ?", subjectTaskID, itemID).First(&existing).Error
		switch {
		case err == nil:
			if done {
				return nil
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if !done {
				return nil
			}
		default:
			return fmt.Errorf("load completion: %w", err)
		}
		if done {
			drow := checklistdomain.TaskChecklistCompletion{
				TaskID:     subjectTaskID,
				ItemID:     itemID,
				At:         time.Now().UTC(),
				By:         string(by),
				VerifiedBy: checklistdomain.VerifierLegacy,
			}
			row := checklistmodel.FromDomainTaskChecklistCompletion(drow)
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "task_id"}, {Name: "item_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"at", "done_by", "verified_by"}),
			}).Create(&row).Error; err != nil {
				return fmt.Errorf("save completion: %w", err)
			}
		} else {
			res := tx.Where("task_id = ? AND item_id = ?", subjectTaskID, itemID).Delete(&checklistmodel.TaskChecklistCompletion{})
			if res.Error != nil {
				return fmt.Errorf("delete completion: %w", res.Error)
			}
		}
		seq, err := storekernel.NextEventSeq(tx, subjectTaskID)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(map[string]any{"item_id": itemID, "done": done})
		if err := storekernel.AppendEvent(tx, subjectTaskID, seq, taskeventsdomain.EventChecklistItemToggled, by, b); err != nil {
			return err
		}
		_, err = syncCriteriaSatisfiedAtInTx(tx, subjectTaskID, by)
		return err
	})
}
