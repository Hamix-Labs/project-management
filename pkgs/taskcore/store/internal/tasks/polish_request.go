package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RequestPolishInput is the store payload for operator polish from review.
type RequestPolishInput = contract.RequestPolishInput

// RequestTaskPolish sets pending_retry (kind=polish) and status=ready from review.
// When flags/new criteria are present, clears flagged completions and appends
// new definition rows in the same transaction.
func RequestTaskPolish(ctx context.Context, db *gorm.DB, in RequestPolishInput, by domain.Actor) (*domain.Task, domain.Status, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.RequestTaskPolish", "task_id", in.TaskID)
	if err := domain.ValidateActor(by); err != nil {
		return nil, "", err
	}
	if by != domain.ActorUser {
		return nil, "", fmt.Errorf("%w: polish requires user actor", domain.ErrInvalidInput)
	}
	taskID := strings.TrimSpace(in.TaskID)
	if taskID == "" {
		return nil, "", fmt.Errorf("%w: id", domain.ErrInvalidInput)
	}
	instructions := strings.TrimSpace(in.Instructions)
	flagged := normalizePolishIDList(in.FlaggedCriterionIDs)
	newItems := normalizePolishNewCriteria(in.NewCriteria)
	intent := domain.PendingRetry{
		Kind:                domain.PendingKindPolish,
		Mode:                domain.RetryResume,
		ParentCycleID:       strings.TrimSpace(in.ParentCycleID),
		Instructions:        instructions,
		FlaggedCriterionIDs: flagged,
	}
	var updated *domain.Task
	var origStatus domain.Status
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur model.Task
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", taskID).First(&cur).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return fmt.Errorf("load task: %w", err)
		}
		dcur := model.ToDomainTask(cur)
		origStatus = dcur.Status
		parentID, err := resolvePolishParentCycleInTx(tx, taskID, intent.ParentCycleID)
		if err != nil {
			return err
		}
		intent.ParentCycleID = parentID

		if dcur.Status == domain.StatusReady && dcur.PendingRetry != nil {
			if polishIntentAlreadyQueued(dcur.PendingRetry, intent, flagged, newItems) {
				updated = &dcur
				return nil
			}
			return fmt.Errorf("%w: task already queued with different retry intent", domain.ErrConflict)
		}
		if dcur.Status != domain.StatusReview {
			return fmt.Errorf("%w: task status is %q, want review", domain.ErrInvalidInput, dcur.Status)
		}

		if len(flagged) > 0 {
			if err := checkliststore.ClearCompletionsForPolishInTx(tx, taskID, flagged, by); err != nil {
				return err
			}
		}
		newIDs, err := checkliststore.AddItemsInTx(tx, taskID, newItems, by)
		if err != nil {
			return err
		}
		intent.FlaggedCriterionIDs = flagged
		intent.NewCriterionIDs = newIDs
		if err := intent.Validate(); err != nil {
			return err
		}

		nextSeq, err := eventsaudit.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"kind":                  string(intent.Kind),
			"mode":                  string(intent.Mode),
			"parent_cycle_id":       intent.ParentCycleID,
			"instructions":          intent.Instructions,
			"flagged_criterion_ids": intent.FlaggedCriterionIDs,
			"new_criterion_ids":     intent.NewCriterionIDs,
			"skip_verify":           intent.SkipVerify,
		})
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, nextSeq, taskeventsdomain.EventTaskPolishRequested, by, payload); err != nil {
			return err
		}
		nextSeq++
		dcur.PendingRetry = &intent
		ready := domain.StatusReady
		if err := applyStatusPatch(tx, taskID, &dcur, &ready, by, &nextSeq); err != nil {
			return err
		}
		cur = model.FromDomainTask(dcur)
		if err := tx.Save(&cur).Error; err != nil {
			return fmt.Errorf("save task: %w", err)
		}
		if err := hydrateDependsOn(ctx, tx, &dcur); err != nil {
			return err
		}
		updated = &dcur
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrNotFound
		}
		return nil, "", err
	}
	return updated, origStatus, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func resolvePolishParentCycleInTx(tx *gorm.DB, taskID, explicit string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		var c cyclesmodel.TaskCycle
		if err := tx.Where("id = ?", explicit).First(&c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", domain.ErrNotFound
			}
			return "", fmt.Errorf("load parent cycle: %w", err)
		}
		dc := cyclesmodel.ToDomainTaskCycle(c)
		if dc.TaskID != taskID {
			return "", fmt.Errorf("%w: parent_cycle_id does not belong to this task", domain.ErrInvalidInput)
		}
		if dc.Status != cyclesdomain.CycleStatusSucceeded {
			return "", fmt.Errorf("%w: polish parent cycle must be succeeded", domain.ErrInvalidInput)
		}
		return dc.ID, nil
	}
	var cycles []cyclesmodel.TaskCycle
	if err := tx.Where("task_id = ?", taskID).Order("attempt_seq DESC").Limit(50).Find(&cycles).Error; err != nil {
		return "", fmt.Errorf("list cycles: %w", err)
	}
	for i := range cycles {
		dc := cyclesmodel.ToDomainTaskCycle(cycles[i])
		if dc.Status == cyclesdomain.CycleStatusSucceeded {
			return dc.ID, nil
		}
	}
	return "", fmt.Errorf("%w: no succeeded cycle to polish from", domain.ErrInvalidInput)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func normalizePolishNewCriteria(items []contract.CreateChecklistItemInput) []contract.CreateChecklistItemInput {
	if len(items) == 0 {
		return nil
	}
	out := make([]contract.CreateChecklistItemInput, 0, len(items))
	for _, raw := range items {
		t := strings.TrimSpace(raw.Text)
		if t == "" {
			continue
		}
		out = append(out, contract.CreateChecklistItemInput{
			Text:           t,
			VerifyCommands: raw.VerifyCommands,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func normalizePolishIDList(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// polishIntentAlreadyQueued reports whether a queued polish pending_retry matches
// this request. New criterion texts cannot be re-inserted; match by instructions,
// parent, flags, and new-item count when items are present.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func polishIntentAlreadyQueued(queued *domain.PendingRetry, intent domain.PendingRetry, flagged []string, newItems []contract.CreateChecklistItemInput) bool {
	if queued == nil || queued.NormalizeKind() != domain.PendingKindPolish {
		return false
	}
	probe := intent
	probe.FlaggedCriterionIDs = flagged
	if len(newItems) == 0 {
		probe.NewCriterionIDs = nil
		if err := probe.Validate(); err != nil {
			return false
		}
		return queued.Equal(&probe)
	}
	if queued.ParentCycleID != intent.ParentCycleID ||
		strings.TrimSpace(queued.Instructions) != strings.TrimSpace(intent.Instructions) ||
		!stringSlicesEqual(queued.FlaggedCriterionIDs, flagged) ||
		len(queued.NewCriterionIDs) != len(newItems) {
		return false
	}
	return true
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
