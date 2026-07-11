package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"gorm.io/gorm"
)

// GetGitWorktreeByID loads a worktree row by primary key.
func (s *Store) GetGitWorktreeByID(ctx context.Context, worktreeID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.GetGitWorktreeByID")
	worktreeID = strings.TrimSpace(worktreeID)
	if worktreeID == "" {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: worktree_id required", domain.ErrInvalidInput)
	}
	var row model.GitWorktree
	err := s.db.WithContext(ctx).Where("id = ?", worktreeID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gitdomain.GitWorktree{}, gitdomain.NewGitErr(gitdomain.GitCodeWorktreeNotFound, "worktree not found")
		}
		return gitdomain.GitWorktree{}, fmt.Errorf("get git worktree: %w", err)
	}
	return model.ToDomainGitWorktree(row), nil
}

// GetGitBranchByID loads a branch row by primary key.
func (s *Store) GetGitBranchByID(ctx context.Context, branchID string) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.GetGitBranchByID")
	branchID = strings.TrimSpace(branchID)
	if branchID == "" {
		return gitdomain.GitBranch{}, fmt.Errorf("%w: branch_id required", domain.ErrInvalidInput)
	}
	var row model.GitBranch
	err := s.db.WithContext(ctx).Where("id = ?", branchID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gitdomain.GitBranch{}, gitdomain.NewGitErr(gitdomain.GitCodeBranchNotFound, "branch not found")
		}
		return gitdomain.GitBranch{}, fmt.Errorf("get git branch: %w", err)
	}
	return model.ToDomainGitBranch(row), nil
}

// GetGitRepositoryByID loads a repository row by primary key.
func (s *Store) GetGitRepositoryByID(ctx context.Context, repoID string) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.GetGitRepositoryByID")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return gitdomain.GitRepository{}, fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	var row model.GitRepository
	err := s.db.WithContext(ctx).Where("id = ?", repoID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gitdomain.GitRepository{}, gitdomain.NewGitErr(gitdomain.GitCodeRepositoryNotFound, "repository not found")
		}
		return gitdomain.GitRepository{}, fmt.Errorf("get git repository: %w", err)
	}
	return model.ToDomainGitRepository(row), nil
}
