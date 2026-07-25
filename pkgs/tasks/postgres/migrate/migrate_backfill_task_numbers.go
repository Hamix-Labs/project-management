package migrate

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"gorm.io/gorm"
)

// migrateBackfillTaskNumbers (schema rev 17) assigns dense per-project
// tasks.number values for existing rows and sets projects.next_task_number
// to n+1. Tasks with null project_id stay unnumbered. Idempotent: skips
// projects whose tasks already have numbers (next_task_number > 1 or any
// task.number set).
func migrateBackfillTaskNumbers(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateBackfillTaskNumbers")

	var projects []projectmodel.Project
	if err := db.WithContext(ctx).Find(&projects).Error; err != nil {
		return fmt.Errorf("list projects for task number backfill: %w", err)
	}

	for _, p := range projects {
		if err := backfillProjectTaskNumbers(ctx, db, p); err != nil {
			return err
		}
	}
	return nil
}

type taskCreatedRow struct {
	ID        string
	Number    *int
	CreatedAt *time.Time `gorm:"column:created_at"`
}

func backfillProjectTaskNumbers(ctx context.Context, db *gorm.DB, p projectmodel.Project) error {
	var numbered int64
	if err := db.WithContext(ctx).Model(&taskmodel.Task{}).
		Where("project_id = ? AND number IS NOT NULL", p.ID).
		Count(&numbered).Error; err != nil {
		return fmt.Errorf("count numbered tasks for project %s: %w", p.ID, err)
	}
	if numbered > 0 || p.NextTaskNumber > 1 {
		// Already migrated or live creates advanced the counter.
		return nil
	}

	var rows []taskCreatedRow
	err := db.WithContext(ctx).
		Table("tasks").
		Select("tasks.id, tasks.number, te.at AS created_at").
		Joins("LEFT JOIN task_events te ON te.task_id = tasks.id AND te.seq = 1 AND te.type = ?", "task_created").
		Where("tasks.project_id = ?", p.ID).
		Order("te.at ASC NULLS LAST, tasks.id ASC").
		Scan(&rows).Error
	if err != nil {
		// SQLite lacks NULLS LAST — fall back to simpler order.
		err = db.WithContext(ctx).
			Table("tasks").
			Select("tasks.id, tasks.number, te.at AS created_at").
			Joins("LEFT JOIN task_events te ON te.task_id = tasks.id AND te.seq = 1 AND te.type = ?", "task_created").
			Where("tasks.project_id = ?", p.ID).
			Order("te.at ASC, tasks.id ASC").
			Scan(&rows).Error
		if err != nil {
			return fmt.Errorf("list tasks for project %s backfill: %w", p.ID, err)
		}
	}

	next := 1
	for _, row := range rows {
		n := next
		if err := db.WithContext(ctx).Model(&taskmodel.Task{}).
			Where("id = ?", row.ID).
			Update("number", n).Error; err != nil {
			return fmt.Errorf("assign task number %d to %s: %w", n, row.ID, err)
		}
		next++
	}
	if err := db.WithContext(ctx).Model(&projectmodel.Project{}).
		Where("id = ?", p.ID).
		Update("next_task_number", next).Error; err != nil {
		return fmt.Errorf("set next_task_number for project %s: %w", p.ID, err)
	}
	if len(rows) > 0 {
		slog.Info("migrate backfill task numbers",
			"project_id", p.ID,
			"assigned", len(rows),
			"next_task_number", next)
	}
	return nil
}
