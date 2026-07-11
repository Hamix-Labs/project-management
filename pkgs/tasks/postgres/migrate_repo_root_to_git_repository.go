package postgres

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	taskdomain "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func migrateRepoRootToGitRepository(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateRepoRootToGitRepository")
	var path string
	err := db.WithContext(ctx).
		Raw(`SELECT COALESCE(repo_root, '') FROM app_settings WHERE id = ?`, taskdomain.AppSettingsRowID).
		Scan(&path).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "repo_root") {
			return nil
		}
		return err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	gitSvc := gitwork.New()
	opened, err := gitSvc.OpenRepository(ctx, path)
	if err != nil {
		slog.Warn("repo_root migration skipped: not a git repository", "path", path, "err", err)
		return nil
	}
	repoRoot := opened.Root
	var existing int64
	if err := db.WithContext(ctx).Model(&gitmodel.GitRepository{}).
		Where("path = ?", repoRoot).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	branches, err := gitSvc.ListBranches(ctx, opened)
	if err != nil {
		slog.Warn("repo_root migration skipped: list branches failed", "path", path, "err", err)
		return nil
	}
	now := time.Now().UTC()
	repo := gitmodel.FromDomainGitRepository(gitdomain.GitRepository{
		ID: uuid.NewString(), Path: opened.Root, DefaultBranch: "main", CreatedAt: now, UpdatedAt: now,
	})
	mainWT := gitmodel.FromDomainGitWorktree(gitdomain.GitWorktree{
		ID: uuid.NewString(), RepositoryID: repo.ID, Path: opened.Root, Name: "main", IsMain: true, CreatedAt: now,
	})
	var branchRows []gitmodel.GitBranch
	for _, b := range branches {
		branchRows = append(branchRows, gitmodel.FromDomainGitBranch(gitdomain.GitBranch{
			ID: uuid.NewString(), RepositoryID: repo.ID, Name: b.Name, HeadSHA: b.HeadSHA, CreatedAt: now,
		}))
	}
	if len(branchRows) == 0 {
		branchRows = append(branchRows, gitmodel.FromDomainGitBranch(gitdomain.GitBranch{
			ID: uuid.NewString(), RepositoryID: repo.ID, Name: "main", CreatedAt: now,
		}))
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&repo).Error; err != nil {
			return err
		}
		if err := tx.Create(&mainWT).Error; err != nil {
			return err
		}
		return tx.Create(&branchRows).Error
	})
}
