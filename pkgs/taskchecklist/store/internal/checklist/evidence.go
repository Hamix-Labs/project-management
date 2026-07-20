package checklist

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/taskload"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxEvidenceBytes = 16 * 1024
const maxReasoningBytes = 16 * 1024

// SetDoneWithEvidenceInTx records completion with proof metadata inside
// an existing transaction. Only taskcoredomain.ActorAgent may write.
func SetDoneWithEvidenceInTx(
	tx *gorm.DB,
	subjectTaskID, itemID string,
	evidence string,
	verifier checklistdomain.VerifierKind,
	reasoning, cycleID string,
	by taskcoredomain.Actor,
) (CriteriaFlagChange, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.SetDoneWithEvidenceInTx")
	if err := taskcoredomain.ValidateActor(by); err != nil {
		return CriteriaFlagChange{}, err
	}
	if by != taskcoredomain.ActorAgent {
		return CriteriaFlagChange{}, fmt.Errorf("%w: only the agent may mark checklist items done or undone", taskcoredomain.ErrInvalidInput)
	}
	if err := validateEvidencePayload(evidence, verifier, reasoning); err != nil {
		return CriteriaFlagChange{}, err
	}
	subjectTaskID = strings.TrimSpace(subjectTaskID)
	itemID = strings.TrimSpace(itemID)
	if subjectTaskID == "" || itemID == "" {
		return CriteriaFlagChange{}, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if _, err := taskload.LoadTask(tx, subjectTaskID); err != nil {
		return CriteriaFlagChange{}, err
	}
	defOwner, err := DefinitionSourceTaskIDInTx(tx, subjectTaskID)
	if err != nil {
		return CriteriaFlagChange{}, err
	}
	var it checklistmodel.TaskChecklistItem
	if err := tx.Where("id = ? AND task_id = ?", itemID, defOwner).First(&it).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CriteriaFlagChange{}, taskcoredomain.ErrNotFound
		}
		return CriteriaFlagChange{}, fmt.Errorf("load checklist item: %w", err)
	}
	drow := checklistdomain.TaskChecklistCompletion{
		TaskID:            subjectTaskID,
		ItemID:            itemID,
		At:                time.Now().UTC(),
		By:                string(by),
		Evidence:          evidence,
		VerifiedBy:        verifier,
		VerifierReasoning: reasoning,
		CycleID:           strings.TrimSpace(cycleID),
	}
	row := checklistmodel.FromDomainTaskChecklistCompletion(drow)
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "task_id"}, {Name: "item_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"at", "done_by", "evidence", "verified_by", "verifier_reasoning", "cycle_id",
		}),
	}).Create(&row).Error; err != nil {
		return CriteriaFlagChange{}, fmt.Errorf("save completion: %w", err)
	}
	seq, err := eventsaudit.NextEventSeq(tx, subjectTaskID)
	if err != nil {
		return CriteriaFlagChange{}, err
	}
	b, _ := json.Marshal(map[string]any{
		"item_id": itemID, "done": true,
		"verified_by": string(verifier), "cycle_id": drow.CycleID,
	})
	if err := eventsaudit.AppendEvent(tx, subjectTaskID, seq, taskeventsdomain.EventChecklistItemToggled, by, b); err != nil {
		return CriteriaFlagChange{}, err
	}
	return syncCriteriaSatisfiedAtInTx(tx, subjectTaskID, by)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateEvidencePayload(evidence string, verifier checklistdomain.VerifierKind, reasoning string) error {
	if !checklistdomain.ValidVerifierKind(verifier) {
		return fmt.Errorf("%w: invalid verified_by", taskcoredomain.ErrInvalidInput)
	}
	if verifier != checklistdomain.VerifierLegacy {
		if strings.TrimSpace(evidence) == "" {
			return fmt.Errorf("%w: evidence required", taskcoredomain.ErrInvalidInput)
		}
	}
	if len(evidence) > maxEvidenceBytes {
		return fmt.Errorf("%w: evidence too long", taskcoredomain.ErrInvalidInput)
	}
	if len(reasoning) > maxReasoningBytes {
		return fmt.Errorf("%w: verifier_reasoning too long", taskcoredomain.ErrInvalidInput)
	}
	return nil
}

// SetDoneWithEvidence is the non-transactional wrapper.
func SetDoneWithEvidence(
	ctx context.Context,
	db *gorm.DB,
	subjectTaskID, itemID string,
	evidence string,
	verifier checklistdomain.VerifierKind,
	reasoning, cycleID string,
	by taskcoredomain.Actor,
) (CriteriaFlagChange, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.checklist.SetDoneWithEvidence")
	defer storekernel.DeferLatency(storekernel.OpSetChecklistItemDone)()
	var flag CriteriaFlagChange
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		flag, err = SetDoneWithEvidenceInTx(tx, subjectTaskID, itemID, evidence, verifier, reasoning, cycleID, by)
		return err
	})
	return flag, err
}
