package store

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// CreateGitWorktreeInput adds a worktree on disk and persists the row.
type CreateGitWorktreeInput = contract.CreateGitWorktreeInput

// ListGitWorktrees returns worktrees for a repository.
func (s *Store) ListGitWorktrees(ctx context.Context, projectID, repoID string) ([]gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ListGitWorktrees")
	if _, err := s.GetGitRepository(ctx, projectID, repoID); err != nil {
		return nil, err
	}
	var rows []model.GitWorktree
	err := s.db.WithContext(ctx).
		Where("repository_id = ?", repoID).
		Order("is_main DESC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list git worktrees: %w", err)
	}
	return model.ToDomainGitWorktrees(rows), nil
}

// GetGitWorktree returns one worktree by ID. The projectID parameter is
// accepted for API compatibility but ignored — repositories are global.
func (s *Store) GetGitWorktree(ctx context.Context, projectID, worktreeID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.GetGitWorktree")
	var row model.GitWorktree
	err := s.db.WithContext(ctx).
		Where("id = ?", worktreeID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gitdomain.GitWorktree{}, gitdomain.NewGitErr(gitdomain.GitCodeWorktreeNotFound, "worktree not found")
		}
		return gitdomain.GitWorktree{}, fmt.Errorf("get git worktree: %w", err)
	}
	return model.ToDomainGitWorktree(row), nil
}

// CreateGitWorktree adds a linked worktree via git and inserts a row.
func (s *Store) CreateGitWorktree(ctx context.Context, projectID, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.CreateGitWorktree")
	repo, err := s.GetGitRepository(ctx, projectID, repoID)
	if err != nil {
		return gitdomain.GitWorktree{}, err
	}
	return s.createGitWorktreeOnRepo(ctx, repo, input, gitSvc)
}

// UnregisterGitWorktree removes Hamix registration for a worktree without
// running git worktree remove — the checkout directory stays on disk.
func (s *Store) UnregisterGitWorktree(ctx context.Context, projectID, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.UnregisterGitWorktree")
	if _, err := s.GetGitWorktree(ctx, projectID, worktreeID); err != nil {
		return err
	}
	if err := guardNoRunningTask(ctx, s.db, worktreeID); err != nil {
		return err
	}
	res := s.db.WithContext(ctx).Delete(&model.GitWorktree{}, "id = ?", worktreeID)
	if res.Error != nil {
		return fmt.Errorf("unregister git worktree row: %w", res.Error)
	}
	return nil
}

// RemoveGitWorktreeFromDisk runs git worktree remove then deletes the Hamix row.
func (s *Store) RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.RemoveGitWorktreeFromDisk")
	wt, err := s.GetGitWorktree(ctx, projectID, worktreeID)
	if err != nil {
		return err
	}
	return s.removeGitWorktreeFromDisk(ctx, wt, force, gitSvc)
}

// RemoveGitWorktreeFromDiskByID is the global-route variant of RemoveGitWorktreeFromDisk.
func (s *Store) RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.RemoveGitWorktreeFromDiskByID")
	wt, err := s.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return err
	}
	return s.removeGitWorktreeFromDisk(ctx, wt, force, gitSvc)
}

func (s *Store) removeGitWorktreeFromDisk(ctx context.Context, wt gitdomain.GitWorktree, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.removeGitWorktreeFromDisk")
	if wt.IsMain {
		return fmt.Errorf("%w: cannot remove main worktree from disk", taskcoredomain.ErrInvalidInput)
	}
	if err := guardNoRunningTask(ctx, s.db, wt.ID); err != nil {
		return err
	}
	repo, err := s.GetGitRepositoryByID(ctx, wt.RepositoryID)
	if err != nil {
		return err
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	opened, err := gitSvc.OpenRepository(ctx, repo.Path)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if err := gitSvc.RemoveWorktree(ctx, opened, wt.Path, force); err != nil {
		return mapGitworkRemoveErr(err)
	}
	res := s.db.WithContext(ctx).Delete(&model.GitWorktree{}, "id = ?", wt.ID)
	if res.Error != nil {
		return fmt.Errorf("delete git worktree row: %w", res.Error)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func mapGitworkCreateErr(err error) error {
	switch {
	case errors.Is(err, gitwork.ErrWorktreeExists):
		return gitdomain.NewGitErr(gitdomain.GitCodePathExists, "worktree path already exists")
	case errors.Is(err, gitwork.ErrBranchCheckedOut):
		return gitdomain.NewGitErr(gitdomain.GitCodeBranchCheckedOut, "branch is checked out in another worktree")
	default:
		return err
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func mapGitworkRemoveErr(err error) error {
	if errors.Is(err, gitwork.ErrDirty) {
		return gitdomain.NewGitErr(gitdomain.GitCodePathExists, "worktree has uncommitted changes; use force")
	}
	return err
}
