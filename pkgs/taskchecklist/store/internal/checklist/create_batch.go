package checklist

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/taskload"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDefinitionItemsAtCreateInTx inserts definition rows during POST /tasks
// inside the create transaction. Unlike Add, it does not re-check
// ValidateCanAddCriterionInTx because the row was just inserted.
func SeedDefinitionItemsAtCreateInTx(tx *gorm.DB, taskID string, items []CreateChecklistItemInput, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.SeedDefinitionItemsAtCreateInTx")
	if err := storekernel.ValidateActor(by); err != nil {
		return err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if len(items) == 0 {
		return nil
	}
	if _, err := taskload.LoadTask(tx, taskID); err != nil {
		return err
	}
	var maxOrder int
	row := tx.Model(&checklistmodel.TaskChecklistItem{}).Select("COALESCE(MAX(sort_order), 0)").Where("task_id = ?", taskID)
	if err := row.Scan(&maxOrder).Error; err != nil {
		return fmt.Errorf("checklist order: %w", err)
	}
	seq, err := storekernel.NextEventSeq(tx, taskID)
	if err != nil {
		return err
	}
	for _, raw := range items {
		text := strings.TrimSpace(raw.Text)
		if text == "" {
			continue
		}
		cmds, err := NormalizeVerifyCommandInputs(raw.VerifyCommands)
		if err != nil {
			return err
		}
		maxOrder++
		it := checklistdomain.TaskChecklistItem{
			ID:        uuid.NewString(),
			TaskID:    taskID,
			SortOrder: maxOrder,
			Text:      text,
		}
		if err := tx.Create(checklistmodel.FromDomainTaskChecklistItemPtr(&it)).Error; err != nil {
			return fmt.Errorf("insert checklist item: %w", err)
		}
		if len(cmds) > 0 {
			if err := replaceCommandsInTx(tx, it.ID, cmds); err != nil {
				return err
			}
		}
		b, _ := json.Marshal(map[string]string{"item_id": it.ID, "text": it.Text})
		if err := storekernel.AppendEvent(tx, taskID, seq, taskeventsdomain.EventChecklistItemAdded, by, b); err != nil {
			return err
		}
		seq++
	}
	return nil
}
