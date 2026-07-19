package migrate

import (
	"context"
	"fmt"
	"log/slog"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"gorm.io/gorm"
)

// migrateOrphanRepoProjects (schema rev 11) deletes projects whose
// repository_id no longer exists in git_repositories. These orphans appear
// when a repository was deleted before cascade cleanup existed, leaving
// duplicate "Default" rows in the projects list.
func migrateOrphanRepoProjects(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateOrphanRepoProjects")
	var orphans []projectmodel.Project
	err := db.WithContext(ctx).
		Where("repository_id IS NOT NULL").
		Where("repository_id NOT IN (?)", db.Model(&gitmodel.GitRepository{}).Select("id")).
		Find(&orphans).Error
	if err != nil {
		return fmt.Errorf("list orphan projects: %w", err)
	}
	if len(orphans) == 0 {
		return nil
	}
	ids := make([]string, 0, len(orphans))
	for _, p := range orphans {
		ids = append(ids, p.ID)
	}
	if err := db.WithContext(ctx).
		Where("project_id IN ?", ids).
		Delete(&projectmodel.ProjectContextEdge{}).Error; err != nil {
		return fmt.Errorf("delete orphan project context edges: %w", err)
	}
	if err := db.WithContext(ctx).
		Where("project_id IN ?", ids).
		Delete(&projectmodel.ProjectContextItem{}).Error; err != nil {
		return fmt.Errorf("delete orphan project context items: %w", err)
	}
	if err := db.WithContext(ctx).
		Exec("DELETE FROM task_context_snapshots WHERE project_id IN ?", ids).Error; err != nil {
		return fmt.Errorf("delete orphan task context snapshots: %w", err)
	}
	if err := db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&projectmodel.Project{}).Error; err != nil {
		return fmt.Errorf("delete orphan projects: %w", err)
	}
	slog.Info("migrate orphan repo projects", "deleted", len(ids))
	return nil
}
