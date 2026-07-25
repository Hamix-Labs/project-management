package activity

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// activityTypes is the fixed IN-list of event types surfaced on /tasks/activity.
var activityTypes = []string{
	string(taskeventsdomain.EventStatusChanged),
	string(taskeventsdomain.EventPhaseFailed),
	string(taskeventsdomain.EventApprovalGranted),
}

// activityRow is the raw projection from a LEFT JOIN of task_events ⋊ tasks.
type activityRow struct {
	TaskID     string
	Seq        int64
	At         time.Time
	Type       taskeventsdomain.EventType
	By         taskeventsdomain.Actor
	Data       datatypes.JSON `gorm:"column:data_json"`
	TaskTitle  *string        `gorm:"column:task_title"`
	TaskNumber *int           `gorm:"column:task_number"`
}

// List returns paginated activity events across all tasks for the fixed type
// set (status_changed, phase_failed, approval_granted), newest-first.
func List(ctx context.Context, db *gorm.DB, in contract.ListActivityInput) (contract.ListActivityResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.activity.List",
		"limit", in.Limit, "offset", in.Offset)

	limit := in.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	q := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
		Where("task_events.type IN ?", activityTypes)
	if in.Since != nil {
		q = q.Where("task_events.at >= ?", in.Since.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return contract.ListActivityResult{}, fmt.Errorf("count task activity: %w", err)
	}

	var rows []activityRow
	listQ := db.WithContext(ctx).
		Table("task_events").
		Select("task_events.task_id, task_events.seq, task_events.at, task_events.type, task_events.by, task_events.data_json, tasks.title AS task_title, tasks.number AS task_number").
		Joins("LEFT JOIN tasks ON tasks.id = task_events.task_id").
		Where("task_events.type IN ?", activityTypes)
	if in.Since != nil {
		listQ = listQ.Where("task_events.at >= ?", in.Since.UTC())
	}
	if err := listQ.Order("task_events.at DESC, task_events.seq DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error; err != nil {
		return contract.ListActivityResult{}, fmt.Errorf("list task activity: %w", err)
	}

	events := make([]contract.ActivityEvent, 0, len(rows))
	for _, r := range rows {
		ev := contract.ActivityEvent{
			TaskID:     r.TaskID,
			Seq:        r.Seq,
			At:         r.At,
			Type:       r.Type,
			By:         taskcoredomain.Actor(r.By),
			Data:       []byte(r.Data),
			TaskTitle:  r.TaskTitle,
			TaskNumber: r.TaskNumber,
		}
		events = append(events, ev)
	}
	return contract.ListActivityResult{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Events: events,
	}, nil
}
