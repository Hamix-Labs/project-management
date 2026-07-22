package migrate

import (
	"context"
	"fmt"
	"log/slog"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"gorm.io/gorm"
)

// migrateUnlinkedProjects (schema rev 13) attaches non-default projects with a
// null repository_id to the sole registered repository. Pre-ADR-0042 rows can
// appear in GET /projects but are invisible to the create-task picker, which
// lists projects by repository. With 0 or 2+ repos the target is ambiguous —
// leave rows alone and log when any remain unlinked.
func migrateUnlinkedProjects(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateUnlinkedProjects")

	var unlinked int64
	if err := db.WithContext(ctx).
		Model(&projectmodel.Project{}).
		Where("repository_id IS NULL AND is_default = ?", false).
		Count(&unlinked).Error; err != nil {
		return fmt.Errorf("count unlinked projects: %w", err)
	}
	if unlinked == 0 {
		return nil
	}

	var repos []gitmodel.GitRepository
	if err := db.WithContext(ctx).Find(&repos).Error; err != nil {
		return fmt.Errorf("list repositories: %w", err)
	}
	if len(repos) != 1 {
		slog.Info("migrate unlinked projects skipped",
			"unlinked", unlinked,
			"repositories", len(repos),
			"reason", "need exactly one repository to attach")
		return nil
	}

	repoID := repos[0].ID
	res := db.WithContext(ctx).
		Model(&projectmodel.Project{}).
		Where("repository_id IS NULL AND is_default = ?", false).
		Update("repository_id", repoID)
	if res.Error != nil {
		return fmt.Errorf("attach unlinked projects: %w", res.Error)
	}
	slog.Info("migrate unlinked projects",
		"attached", res.RowsAffected,
		"repository_id", repoID)
	return nil
}
