package checklist

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDefinitionItemsAtCreateInTx inserts definition rows during POST /tasks
// inside the create transaction. Unlike Add, it does not re-check
// ValidateCanAddCriterionInTx because the row was just inserted.
func SeedDefinitionItemsAtCreateInTx(tx *gorm.DB, taskID string, items []CreateChecklistItemInput, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.SeedDefinitionItemsAtCreateInTx")
	if err := taskcoredomain.ValidateActor(by); err != nil {
		return err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if len(items) == 0 {
		return nil
	}
	// Caller just inserted the task in this tx — skip LoadTask / MAX(sort_order).
	maxOrder := 0
	seq := int64(2) // task_created used seq 1
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
		if err := eventsaudit.AppendEvent(tx, taskID, seq, taskeventsdomain.EventChecklistItemAdded, by, b); err != nil {
			return err
		}
		seq++
	}
	return nil
}
