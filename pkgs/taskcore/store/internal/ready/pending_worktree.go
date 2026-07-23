package ready

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"gorm.io/gorm"
)

// PendingWorktreeRow is a ready task still waiting for managed-worktree allocate.
type PendingWorktreeRow struct {
	TaskID       string
	RepositoryID string
}

// ListPendingWorktree returns ready tasks with no worktree_id, joined to the
// project repository for eager provision (ADR-0083).
func ListPendingWorktree(ctx context.Context, db *gorm.DB, limit int) ([]PendingWorktreeRow, error) {
	defer storekernel.DeferLatency(storekernel.OpListReadyQueue)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ready.ListPendingWorktree")
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	type scanRow struct {
		TaskID       string `gorm:"column:task_id"`
		RepositoryID string `gorm:"column:repository_id"`
	}
	var rows []scanRow
	err := db.WithContext(ctx).Model(&model.Task{}).
		Select("tasks.id AS task_id, projects.repository_id AS repository_id").
		Joins("INNER JOIN projects ON projects.id = tasks.project_id").
		Where("tasks.status = ?", domain.StatusReady).
		Where("tasks.worktree_id IS NULL OR tasks.worktree_id = ''").
		Where("projects.repository_id IS NOT NULL AND projects.repository_id <> ''").
		Order("tasks.id ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list tasks pending worktree: %w", err)
	}
	out := make([]PendingWorktreeRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, PendingWorktreeRow{TaskID: r.TaskID, RepositoryID: r.RepositoryID})
	}
	return out, nil
}
