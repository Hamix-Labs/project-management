package store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// EnsureTaskStackLayer ensures the worktree is checked out on the task's
// hamix/task-* layer branch, creating it via gh stack add when missing, and
// rebinds git_worktrees.branch_id to that layer (ADR-0097).
func (s *Store) EnsureTaskStackLayer(ctx context.Context, worktreeID, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.EnsureTaskStackLayer",
		"worktree_id", worktreeID, "task_id", taskID)
	worktreeID = strings.TrimSpace(worktreeID)
	taskID = strings.TrimSpace(taskID)
	if worktreeID == "" || taskID == "" {
		return fmt.Errorf("%w: worktree_id and task_id required", taskcoredomain.ErrInvalidInput)
	}
	layer := TaskBranchName(taskID)
	wt, err := s.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return err
	}
	if wt.IsMain {
		return fmt.Errorf("%w: cannot use stack layers on main worktree", taskcoredomain.ErrInvalidInput)
	}
	repo, err := s.GetGitRepositoryByID(ctx, wt.RepositoryID)
	if err != nil {
		return err
	}
	curBr, err := s.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		return err
	}
	head, headErr := s.gitSvc().WorktreeCurrentBranch(ctx, wt.Path)
	if headErr != nil {
		return fmt.Errorf("current branch: %w", headErr)
	}
	if curBr.Name == layer && head == layer {
		return nil
	}
	exists, err := s.localBranchExists(ctx, wt.Path, layer)
	if err != nil {
		return err
	}
	if !exists {
		addErr := s.stackCLI().Add(ctx, wt.Path, layer)
		exists, err = s.localBranchExists(ctx, wt.Path, layer)
		if err != nil {
			return err
		}
		if !exists {
			// Fallback when gh stack is missing/noop: branch at worktree HEAD.
			opened, openErr := s.gitSvc().OpenRepository(ctx, wt.Path)
			if openErr != nil {
				return fmt.Errorf("stack add %q: %v; open worktree: %w", layer, addErr, openErr)
			}
			if _, createErr := s.gitSvc().CreateBranch(ctx, opened, layer, "HEAD"); createErr != nil {
				return fmt.Errorf("stack add %q: %v; create branch: %w", layer, addErr, createErr)
			}
			if coErr := s.gitSvc().Checkout(ctx, wt.Path, layer); coErr != nil {
				return fmt.Errorf("checkout new layer %q: %w", layer, coErr)
			}
		}
	} else if head != layer {
		if err := s.gitSvc().Checkout(ctx, wt.Path, layer); err != nil {
			return fmt.Errorf("checkout layer %q: %w", layer, err)
		}
	}
	br, err := s.ResolveOrCreateBranchForRepo(ctx, repo, BindBranchInput{
		Name:         layer,
		CreateBranch: false,
	})
	if err != nil {
		return err
	}
	return s.SetWorktreeActiveBranch(ctx, wt.ID, br.ID)
}

// SetWorktreeActiveBranch rebinds git_worktrees.branch_id to an existing branch
// in the same repository. Worktree name (root layer identity) is unchanged.
func (s *Store) SetWorktreeActiveBranch(ctx context.Context, worktreeID, branchID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.SetWorktreeActiveBranch",
		"worktree_id", worktreeID, "branch_id", branchID)
	worktreeID = strings.TrimSpace(worktreeID)
	branchID = strings.TrimSpace(branchID)
	if worktreeID == "" || branchID == "" {
		return fmt.Errorf("%w: worktree_id and branch_id required", taskcoredomain.ErrInvalidInput)
	}
	wt, err := s.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return err
	}
	br, err := s.GetGitBranchByID(ctx, branchID)
	if err != nil {
		return err
	}
	if br.RepositoryID != wt.RepositoryID {
		return fmt.Errorf("%w: branch repository mismatch", taskcoredomain.ErrInvalidInput)
	}
	if err := s.GuardBranchNotBoundToOtherWorktree(ctx, branchID, worktreeID); err != nil {
		return err
	}
	if wt.BranchID == branchID {
		return nil
	}
	res := s.db.WithContext(ctx).Model(&model.GitWorktree{}).
		Where("id = ?", worktreeID).
		Update("branch_id", branchID)
	if res.Error != nil {
		return fmt.Errorf("update worktree active branch: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gitdomain.NewGitErr(gitdomain.GitCodeWorktreeNotFound, "worktree not found")
	}
	return nil
}

func (s *Store) localBranchExists(ctx context.Context, worktreePath, branch string) (bool, error) {
	opened, err := s.gitSvc().OpenRepository(ctx, worktreePath)
	if err != nil {
		return false, err
	}
	if _, err = s.gitSvc().BranchHead(ctx, opened, branch); err != nil {
		return false, nil
	}
	return true, nil
}
