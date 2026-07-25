package store

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateGitBranchInput creates a local branch via git.
type CreateGitBranchInput = contract.CreateGitBranchInput

// ListGitBranches returns branches for a repository.
func (s *Store) ListGitBranches(ctx context.Context, projectID, repoID string) ([]gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ListGitBranches")
	if _, err := s.GetGitRepository(ctx, projectID, repoID); err != nil {
		return nil, err
	}
	var rows []model.GitBranch
	err := s.db.WithContext(ctx).
		Where("repository_id = ?", repoID).
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list git branches: %w", err)
	}
	return model.ToDomainGitBranches(rows), nil
}

// GetGitBranch returns one branch by ID. The projectID parameter is
// accepted for API compatibility but ignored — repositories are global.
func (s *Store) GetGitBranch(ctx context.Context, projectID, branchID string) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.GetGitBranch")
	var row model.GitBranch
	err := s.db.WithContext(ctx).
		Where("id = ?", branchID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gitdomain.GitBranch{}, gitdomain.NewGitErr(gitdomain.GitCodeBranchNotFound, "branch not found")
		}
		return gitdomain.GitBranch{}, fmt.Errorf("get git branch: %w", err)
	}
	return model.ToDomainGitBranch(row), nil
}

// CreateGitBranch creates a branch via git and inserts a row.
func (s *Store) CreateGitBranch(ctx context.Context, projectID, repoID string, input CreateGitBranchInput) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.CreateGitBranch")
	repo, err := s.GetGitRepository(ctx, projectID, repoID)
	if err != nil {
		return gitdomain.GitBranch{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return gitdomain.GitBranch{}, fmt.Errorf("%w: name required", taskcoredomain.ErrInvalidInput)
	}
	opened, err := s.gitSvc().OpenRepository(ctx, repo.Path)
	if err != nil {
		return gitdomain.GitBranch{}, fmt.Errorf("open repository: %w", err)
	}
	created, err := s.gitSvc().CreateBranch(ctx, opened, name, strings.TrimSpace(input.StartPoint))
	if err != nil {
		if errors.Is(err, gitwork.ErrBranchExists) {
			return gitdomain.GitBranch{}, gitdomain.NewGitErr(gitdomain.GitCodeBranchExists, "branch already exists")
		}
		return gitdomain.GitBranch{}, err
	}
	row := gitdomain.GitBranch{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Name:         created.Name,
		HeadSHA:      created.HeadSHA,
		CreatedAt:    time.Now().UTC(),
	}
	branchRow := model.FromDomainGitBranch(row)
	if err := s.db.WithContext(ctx).Create(&branchRow).Error; err != nil {
		if storekernel.IsDuplicateKey(err) {
			return gitdomain.GitBranch{}, gitdomain.NewGitErr(gitdomain.GitCodeBranchExists, "branch already exists")
		}
		return gitdomain.GitBranch{}, fmt.Errorf("create git branch row: %w", err)
	}
	return row, nil
}

// DeleteGitBranch removes a branch via git and the database.
func (s *Store) DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.DeleteGitBranch")
	_ = projectID
	return s.DeleteGitBranchByID(ctx, branchID, force)
}

// DeleteGitBranchByID is the global-route variant of DeleteGitBranch.
func (s *Store) DeleteGitBranchByID(ctx context.Context, branchID string, force bool) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.DeleteGitBranchByID")
	return s.deleteGitBranch(ctx, "", branchID, force)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Store) deleteGitBranch(ctx context.Context, projectID, branchID string, force bool) error {
	branch, err := s.GetGitBranch(ctx, projectID, branchID)
	if err != nil {
		return err
	}
	if err := guardNoRunningTask(ctx, s.db, branchID); err != nil {
		return err
	}
	repo, err := s.GetGitRepository(ctx, projectID, branch.RepositoryID)
	if err != nil {
		return err
	}
	opened, err := s.gitSvc().OpenRepository(ctx, repo.Path)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	worktrees, err := s.gitSvc().ListWorktrees(ctx, opened)
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}
	for _, wt := range worktrees {
		if wt.Branch == branch.Name {
			return gitdomain.NewGitErr(gitdomain.GitCodeBranchCheckedOut, "branch is checked out in a worktree")
		}
	}
	if err := s.gitSvc().DeleteBranch(ctx, opened, branch.Name, force); err != nil {
		if errors.Is(err, gitwork.ErrBranchCheckedOut) {
			return gitdomain.NewGitErr(gitdomain.GitCodeBranchCheckedOut, "branch is checked out in a worktree")
		}
		return err
	}
	res := s.db.WithContext(ctx).Delete(&model.GitBranch{}, "id = ?", branchID)
	if res.Error != nil {
		return fmt.Errorf("delete git branch row: %w", res.Error)
	}
	return nil
}
