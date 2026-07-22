package tasks

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type onTaskDoneCommit struct {
	SHA     string `json:"sha"`
	Message string `json:"message,omitempty"`
}

type onTaskDonePayload struct {
	WorktreeID string             `json:"worktree_id,omitempty"`
	Commits    []onTaskDoneCommit `json:"commits"`
}

type approvalGrantedPayload struct {
	From    string `json:"from"`
	CycleID string `json:"cycle_id,omitempty"`
}

// RequestTaskApprove transitions a task from review → done after human approval.
// Returns (task, prevStatus, err). Emits status_changed, approval_granted, and
// on_task_done (commits from the latest succeeded cycle).
func RequestTaskApprove(ctx context.Context, db *gorm.DB, taskID string, by domain.Actor) (*domain.Task, domain.Status, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.RequestTaskApprove", "task_id", taskID)
	if err := domain.ValidateActor(by); err != nil {
		return nil, "", err
	}
	if by != domain.ActorUser {
		return nil, "", fmt.Errorf("%w: approve requires user actor", domain.ErrInvalidInput)
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, "", fmt.Errorf("%w: id", domain.ErrInvalidInput)
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
		if dcur.Status != domain.StatusReview {
			return fmt.Errorf("%w: task status is %q, want review", domain.ErrInvalidInput, dcur.Status)
		}
		if err := checkliststore.ValidateCanMarkDoneInTx(tx, taskID); err != nil {
			return err
		}
		cycleID, commits, err := latestSucceededCycleCommitsInTx(tx, taskID)
		if err != nil {
			return err
		}
		nextSeq, err := eventsaudit.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		statusPayload, err := storekernel.EventPairJSON(string(domain.StatusReview), string(domain.StatusDone))
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, nextSeq, taskeventsdomain.EventStatusChanged, by, statusPayload); err != nil {
			return err
		}
		nextSeq++
		approvalPayload, err := json.Marshal(approvalGrantedPayload{
			From:    string(domain.StatusReview),
			CycleID: cycleID,
		})
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, nextSeq, taskeventsdomain.EventApprovalGranted, by, approvalPayload); err != nil {
			return err
		}
		nextSeq++
		donePayload := onTaskDonePayload{Commits: commits}
		if dcur.WorktreeID != nil {
			donePayload.WorktreeID = *dcur.WorktreeID
		}
		if donePayload.Commits == nil {
			donePayload.Commits = []onTaskDoneCommit{}
		}
		rawDone, err := json.Marshal(donePayload)
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, nextSeq, taskeventsdomain.EventOnTaskDone, by, rawDone); err != nil {
			return err
		}
		dcur.Status = domain.StatusDone
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

func latestSucceededCycleCommitsInTx(tx *gorm.DB, taskID string) (cycleID string, commits []onTaskDoneCommit, err error) {
	var cycles []cyclesmodel.TaskCycle
	if err := tx.Where("task_id = ?", taskID).Order("attempt_seq DESC").Limit(50).Find(&cycles).Error; err != nil {
		return "", nil, fmt.Errorf("list cycles: %w", err)
	}
	var latest *cyclesdomain.TaskCycle
	for i := range cycles {
		dc := cyclesmodel.ToDomainTaskCycle(cycles[i])
		if dc.Status == cyclesdomain.CycleStatusSucceeded {
			latest = &dc
			break
		}
	}
	if latest == nil {
		return "", []onTaskDoneCommit{}, nil
	}
	var rows []cyclesmodel.TaskCycleCommit
	if err := tx.Where("cycle_id = ?", latest.ID).Order("seq ASC").Find(&rows).Error; err != nil {
		return "", nil, fmt.Errorf("list cycle commits: %w", err)
	}
	out := make([]onTaskDoneCommit, 0, len(rows))
	for _, r := range rows {
		out = append(out, onTaskDoneCommit{SHA: r.SHA, Message: r.Message})
	}
	return latest.ID, out, nil
}
