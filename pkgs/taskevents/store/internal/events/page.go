package events

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
	"gorm.io/gorm"
)

// Page is one window of audit events (newest first) plus stable
// paging metadata. RangeStart / RangeEnd are 1-based positions
// counted from the newest row, so the UI can render "showing N-M
// of Total" without re-counting.
type Page = contract.TaskEventsPage

// PageCursor returns events in descending seq (newest first) using
// keyset paging. Neither cursor: first page (newest rows). beforeSeq:
// rows with seq strictly less than beforeSeq (older page). afterSeq:
// rows with seq strictly greater than afterSeq (newer page), still
// returned newest first. Limit is coerced to [1, 200] with a default
// of 50; beforeSeq and afterSeq must not both be set.
func PageCursor(ctx context.Context, db *gorm.DB, taskID string, limit int, beforeSeq, afterSeq *int64) (*Page, error) {
	defer storekernel.DeferLatency(storekernel.OpListTaskEventsPage)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.events.PageCursor")
	var err error
	taskID, limit, err = validatePageInputs(taskID, limit, beforeSeq, afterSeq)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).Where("task_id = ?", taskID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count task events: %w", err)
	}

	q := db.WithContext(ctx).Where("task_id = ?", taskID)
	if beforeSeq != nil {
		q = q.Where("seq < ?", *beforeSeq)
	} else if afterSeq != nil {
		q = q.Where("seq > ?", *afterSeq)
	}
	var rows []eventsmodel.TaskEvent
	err = q.Order("seq DESC").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list task events page: %w", err)
	}

	out := &Page{Events: eventsmodel.ToDomainTaskEvents(rows), Total: total}
	if err := fillPageBounds(ctx, db, taskID, out, out.Events); err != nil {
		return nil, err
	}
	return out, nil
}

// ApprovalPending reports whether an approval is outstanding: among
// approval-related events, the latest by seq decides — granted clears
// pending, requested sets it. Used by the handler to gate the
// /tasks/{id}/approval write paths.
func ApprovalPending(ctx context.Context, db *gorm.DB, taskID string) (bool, error) {
	defer storekernel.DeferLatency(storekernel.OpApprovalPending)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.events.ApprovalPending")
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return false, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	types := []taskeventsdomain.EventType{taskeventsdomain.EventApprovalRequested, taskeventsdomain.EventApprovalGranted}
	var row eventsmodel.TaskEvent
	err := db.WithContext(ctx).
		Where("task_id = ? AND type IN ?", taskID, types).
		Order("seq DESC").
		Limit(1).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("approval pending lookup: %w", err)
	}
	switch row.Type {
	case taskeventsdomain.EventApprovalGranted:
		return false, nil
	case taskeventsdomain.EventApprovalRequested:
		return true, nil
	default:
		return false, nil
	}
}

func fillPageBounds(ctx context.Context, db *gorm.DB, taskID string, out *Page, rows []taskeventsdomain.TaskEvent) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.events.fillPageBounds")
	if len(rows) == 0 {
		return nil
	}
	maxSeq := rows[0].Seq
	minSeq := rows[len(rows)-1].Seq

	var newerThanMax int64
	if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
		Where("task_id = ? AND seq > ?", taskID, maxSeq).
		Count(&newerThanMax).Error; err != nil {
		return fmt.Errorf("count newer task events: %w", err)
	}
	var olderThanMin int64
	if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
		Where("task_id = ? AND seq < ?", taskID, minSeq).
		Count(&olderThanMin).Error; err != nil {
		return fmt.Errorf("count older task events: %w", err)
	}

	out.RangeStart = newerThanMax + 1
	out.RangeEnd = newerThanMax + int64(len(rows))
	out.HasMoreNewer = newerThanMax > 0
	out.HasMoreOlder = olderThanMin > 0
	return nil
}

func validatePageInputs(taskID string, limit int, beforeSeq, afterSeq *int64) (string, int, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.events.validatePageInputs")
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return "", 0, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if beforeSeq != nil && afterSeq != nil {
		return "", 0, fmt.Errorf("%w: before_seq and after_seq are mutually exclusive", taskcoredomain.ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if beforeSeq != nil && *beforeSeq < 1 {
		return "", 0, fmt.Errorf("%w: before_seq must be a positive integer", taskcoredomain.ErrInvalidInput)
	}
	if afterSeq != nil && *afterSeq < 1 {
		return "", 0, fmt.Errorf("%w: after_seq must be a positive integer", taskcoredomain.ErrInvalidInput)
	}
	return taskID, limit, nil
}
