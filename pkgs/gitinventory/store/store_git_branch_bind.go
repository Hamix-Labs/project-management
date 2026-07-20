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
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BindBranchInput registers or creates a repo-level branch row for worktree assignment.
type BindBranchInput = contract.BindBranchInput

// ResolveOrCreateBranchForRepo returns a git_branches row for name, creating via git when requested.
func (s *Store) ResolveOrCreateBranchForRepo(
	ctx context.Context,
	repo gitdomain.GitRepository,
	input BindBranchInput,
) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ResolveOrCreateBranchForRepo")
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return gitdomain.GitBranch{}, fmt.Errorf("%w: branch name required", taskcoredomain.ErrInvalidInput)
	}
	if input.CreateBranch {
		return s.CreateGitBranchForRepo(ctx, repo.ID, CreateGitBranchInput{
			Name:       name,
			StartPoint: input.StartPoint,
		})
	}
	var existing model.GitBranch
	err := s.db.WithContext(ctx).
		Where("repository_id = ? AND name = ?", repo.ID, name).
		First(&existing).Error
	if err == nil {
		return model.ToDomainGitBranch(existing), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return gitdomain.GitBranch{}, fmt.Errorf("lookup git branch: %w", err)
	}
	opened, err := s.gitSvc().OpenRepository(ctx, repo.Path)
	if err != nil {
		return gitdomain.GitBranch{}, fmt.Errorf("open repository: %w", err)
	}
	live, err := s.gitSvc().ListBranches(ctx, opened)
	if err != nil {
		return gitdomain.GitBranch{}, fmt.Errorf("list branches: %w", err)
	}
	var headSHA string
	for _, b := range live {
		if b.Name == name {
			headSHA = b.HeadSHA
			break
		}
	}
	if headSHA == "" {
		return gitdomain.GitBranch{}, fmt.Errorf("%w: branch %q not found in repository", taskcoredomain.ErrInvalidInput, name)
	}
	now := time.Now().UTC()
	row := gitdomain.GitBranch{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Name:         name,
		HeadSHA:      headSHA,
		CreatedAt:    now,
	}
	branchRow := model.FromDomainGitBranch(row)
	if err := s.db.WithContext(ctx).Create(&branchRow).Error; err != nil {
		if storekernel.IsDuplicateKey(err) {
			var dup model.GitBranch
			if findErr := s.db.WithContext(ctx).
				Where("repository_id = ? AND name = ?", repo.ID, name).
				First(&dup).Error; findErr == nil {
				return model.ToDomainGitBranch(dup), nil
			}
		}
		return gitdomain.GitBranch{}, fmt.Errorf("register git branch row: %w", err)
	}
	return row, nil
}

// resolveBranchForWorktree resolves or creates a branch and guards one-worktree-per-branch.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Store) resolveBranchForWorktree(
	ctx context.Context,
	repo gitdomain.GitRepository,
	worktreeID string,
	input BindBranchInput,
) (gitdomain.GitBranch, error) {
	br, err := s.ResolveOrCreateBranchForRepo(ctx, repo, input)
	if err != nil {
		return gitdomain.GitBranch{}, err
	}
	if err := s.GuardBranchNotBoundToOtherWorktree(ctx, br.ID, worktreeID); err != nil {
		return gitdomain.GitBranch{}, err
	}
	return br, nil
}
