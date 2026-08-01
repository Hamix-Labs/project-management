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

// RequestOpenPRInput is the store payload for operator open-pr from review.
type RequestOpenPRInput = contract.RequestOpenPRInput

type openPRApprovalPayload struct {
	From    string `json:"from"`
	CycleID string `json:"cycle_id,omitempty"`
}

// RequestTaskOpenPR sets pending_retry (kind=open_pr) and status=ready from review.
// Emits approval_granted (human approved work) and open_pr_requested.
func RequestTaskOpenPR(ctx context.Context, db *gorm.DB, in RequestOpenPRInput, by domain.Actor) (*domain.Task, domain.Status, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.RequestTaskOpenPR", "task_id", in.TaskID)
	if err := domain.ValidateActor(by); err != nil {
		return nil, "", err
	}
	if by != domain.ActorUser {
		return nil, "", fmt.Errorf("%w: open-pr requires user actor", domain.ErrInvalidInput)
	}
	taskID := strings.TrimSpace(in.TaskID)
	if taskID == "" {
		return nil, "", fmt.Errorf("%w: id", domain.ErrInvalidInput)
	}
	intent := domain.PendingRetry{
		Kind:          domain.PendingKindOpenPR,
		Mode:          domain.RetryResume,
		ParentCycleID: strings.TrimSpace(in.ParentCycleID),
		SkipVerify:    true,
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
		parentID, err := resolveOpenPRParentCycleInTx(tx, taskID, intent.ParentCycleID)
		if err != nil {
			return err
		}
		intent.ParentCycleID = parentID
		if err := intent.Validate(); err != nil {
			return err
		}

		if dcur.Status == domain.StatusReady && dcur.PendingRetry != nil {
			if openPRIntentAlreadyQueued(dcur.PendingRetry, intent) {
				updated = &dcur
				return nil
			}
			return fmt.Errorf("%w: task already queued with different retry intent", domain.ErrConflict)
		}
		if dcur.Status != domain.StatusReview {
			return fmt.Errorf("%w: task status is %q, want review", domain.ErrInvalidInput, dcur.Status)
		}

		nextSeq, err := eventsaudit.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		approvalPayload, err := json.Marshal(openPRApprovalPayload{
			From:    string(domain.StatusReview),
			CycleID: intent.ParentCycleID,
		})
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, nextSeq, taskeventsdomain.EventApprovalGranted, by, approvalPayload); err != nil {
			return err
		}
		nextSeq++
		payload, err := json.Marshal(map[string]any{
			"kind":            string(intent.Kind),
			"mode":            string(intent.Mode),
			"parent_cycle_id": intent.ParentCycleID,
			"skip_verify":     intent.SkipVerify,
		})
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, nextSeq, taskeventsdomain.EventOpenPRRequested, by, payload); err != nil {
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
func resolveOpenPRParentCycleInTx(tx *gorm.DB, taskID, explicit string) (string, error) {
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
			return "", fmt.Errorf("%w: open_pr parent cycle must be succeeded", domain.ErrInvalidInput)
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
	return "", fmt.Errorf("%w: no succeeded cycle to open PR from", domain.ErrInvalidInput)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func openPRIntentAlreadyQueued(queued *domain.PendingRetry, intent domain.PendingRetry) bool {
	if queued == nil || queued.NormalizeKind() != domain.PendingKindOpenPR {
		return false
	}
	return queued.Equal(&intent)
}
