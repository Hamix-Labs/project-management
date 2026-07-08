package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"gorm.io/gorm"
)

func migrateGitCommonDir(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateGitCommonDir")
	if !db.Migrator().HasTable("git_repositories") {
		return nil
	}
	if !tableHasColumnPortable(db, "git_repositories", "git_common_dir") {
		if err := ensureNullableTextColumn(ctx, db, "git_repositories", "git_common_dir"); err != nil {
			return fmt.Errorf("add git_repositories.git_common_dir: %w", err)
		}
	}
	var repos []model.GitRepository
	if err := db.WithContext(ctx).Find(&repos).Error; err != nil {
		return fmt.Errorf("list git repositories for common dir backfill: %w", err)
	}
	if len(repos) == 0 {
		return nil
	}
	gitSvc := gitwork.New()
	seenCommon := make(map[string]string, len(repos))
	for _, repo := range repos {
		if strings.TrimSpace(repo.GitCommonDir) != "" {
			if prev, ok := seenCommon[repo.GitCommonDir]; ok && prev != repo.ID {
				return fmt.Errorf("duplicate git_common_dir %q on repositories %s and %s", repo.GitCommonDir, prev, repo.ID)
			}
			seenCommon[repo.GitCommonDir] = repo.ID
			continue
		}
		path := strings.TrimSpace(repo.Path)
		if path == "" {
			slog.Warn("git_common_dir backfill skipped: empty path", "repository_id", repo.ID)
			continue
		}
		mainRoot, commonDir, err := gitSvc.ResolveRegistration(ctx, path)
		if err != nil {
			slog.Warn("git_common_dir backfill skipped: resolve failed", "repository_id", repo.ID, "path", path, "err", err)
			continue
		}
		if prev, ok := seenCommon[commonDir]; ok && prev != repo.ID {
			return fmt.Errorf("duplicate git object database %q on repositories %s and %s", commonDir, prev, repo.ID)
		}
		seenCommon[commonDir] = repo.ID
		if err := db.WithContext(ctx).Model(&model.GitRepository{}).
			Where("id = ?", repo.ID).
			Updates(map[string]any{"git_common_dir": commonDir, "path": mainRoot}).Error; err != nil {
			return fmt.Errorf("backfill git_common_dir for %s: %w", repo.ID, err)
		}
	}
	return nil
}
