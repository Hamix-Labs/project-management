package checklist

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/taskload"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddTextsInTx appends definition rows inside an outer transaction (e.g. polish).
// Returns created item IDs in input order (empty texts skipped).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AddTextsInTx(tx *gorm.DB, taskID string, texts []string, by taskcoredomain.Actor) ([]string, error) {
	items := make([]CreateChecklistItemInput, 0, len(texts))
	for _, t := range texts {
		items = append(items, CreateChecklistItemInput{Text: t})
	}
	return AddItemsInTx(tx, taskID, items, by)
}

// AddItemsInTx appends definition rows (optional verify commands) inside an outer TX.
// Returns created item IDs in input order (empty texts skipped).
func AddItemsInTx(tx *gorm.DB, taskID string, items []CreateChecklistItemInput, by taskcoredomain.Actor) ([]string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.AddItemsInTx")
	if err := taskcoredomain.ValidateActor(by); err != nil {
		return nil, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	t, err := taskload.LoadTask(tx, taskID)
	if err != nil {
		return nil, err
	}
	if err := ValidateCriteriaMutable(t); err != nil {
		return nil, err
	}
	var maxOrder int
	row := tx.Model(&checklistmodel.TaskChecklistItem{}).Select("COALESCE(MAX(sort_order), 0)").Where("task_id = ?", taskID)
	if err := row.Scan(&maxOrder).Error; err != nil {
		return nil, fmt.Errorf("checklist order: %w", err)
	}
	seq, err := eventsaudit.NextEventSeq(tx, taskID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		text := strings.TrimSpace(raw.Text)
		if text == "" {
			continue
		}
		cmds, err := NormalizeVerifyCommandInputs(raw.VerifyCommands)
		if err != nil {
			return nil, err
		}
		maxOrder++
		dit := checklistdomain.TaskChecklistItem{
			ID:        uuid.NewString(),
			TaskID:    taskID,
			SortOrder: maxOrder,
			Text:      text,
		}
		if err := tx.Create(checklistmodel.FromDomainTaskChecklistItemPtr(&dit)).Error; err != nil {
			return nil, fmt.Errorf("insert checklist item: %w", err)
		}
		if len(cmds) > 0 {
			if err := replaceCommandsInTx(tx, dit.ID, cmds); err != nil {
				return nil, err
			}
		}
		b, _ := json.Marshal(map[string]string{"item_id": dit.ID, "text": dit.Text})
		if err := eventsaudit.AppendEvent(tx, taskID, seq, taskeventsdomain.EventChecklistItemAdded, by, b); err != nil {
			return nil, err
		}
		seq++
		out = append(out, dit.ID)
	}
	return out, nil
}

// ClearCompletionsForPolishInTx deletes completion rows for flagged criteria so
// the next polish cycle must re-verify them. Allowed for ActorUser (polish only).
func ClearCompletionsForPolishInTx(tx *gorm.DB, subjectTaskID string, itemIDs []string, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.ClearCompletionsForPolishInTx")
	if err := taskcoredomain.ValidateActor(by); err != nil {
		return err
	}
	if by != taskcoredomain.ActorUser {
		return fmt.Errorf("%w: polish reopen requires user actor", taskcoredomain.ErrInvalidInput)
	}
	subjectTaskID = strings.TrimSpace(subjectTaskID)
	if subjectTaskID == "" {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if len(itemIDs) == 0 {
		return nil
	}
	if _, err := taskload.LoadTask(tx, subjectTaskID); err != nil {
		return err
	}
	defOwner, err := DefinitionSourceTaskIDInTx(tx, subjectTaskID)
	if err != nil {
		return err
	}
	seq, err := eventsaudit.NextEventSeq(tx, subjectTaskID)
	if err != nil {
		return err
	}
	for _, raw := range itemIDs {
		itemID := strings.TrimSpace(raw)
		if itemID == "" {
			continue
		}
		var it checklistmodel.TaskChecklistItem
		if err := tx.Where("id = ? AND task_id = ?", itemID, defOwner).First(&it).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: flagged criterion %s", taskcoredomain.ErrNotFound, itemID)
			}
			return fmt.Errorf("load checklist item: %w", err)
		}
		res := tx.Where("task_id = ? AND item_id = ?", subjectTaskID, itemID).Delete(&checklistmodel.TaskChecklistCompletion{})
		if res.Error != nil {
			return fmt.Errorf("delete completion: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			continue
		}
		b, _ := json.Marshal(map[string]any{"item_id": itemID, "done": false, "reason": "polish_reopen"})
		if err := eventsaudit.AppendEvent(tx, subjectTaskID, seq, taskeventsdomain.EventChecklistItemToggled, by, b); err != nil {
			return err
		}
		seq++
	}
	_, err = syncCriteriaSatisfiedAtInTx(tx, subjectTaskID, by)
	return err
}
