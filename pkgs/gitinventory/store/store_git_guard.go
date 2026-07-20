package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"gorm.io/gorm"
)

// GuardBranchNotBoundToOtherWorktree rejects when branchID is already assigned to another worktree.
func (s *Store) GuardBranchNotBoundToOtherWorktree(ctx context.Context, branchID, exceptWorktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.GuardBranchNotBoundToOtherWorktree")
	branchID = strings.TrimSpace(branchID)
	if branchID == "" {
		return fmt.Errorf("%w: branch_id required", taskcoredomain.ErrInvalidInput)
	}
	var other model.GitWorktree
	q := s.db.WithContext(ctx).Where("branch_id = ?", branchID)
	if exceptWorktreeID = strings.TrimSpace(exceptWorktreeID); exceptWorktreeID != "" {
		q = q.Where("id <> ?", exceptWorktreeID)
	}
	err := q.First(&other).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check branch worktree binding: %w", err)
	}
	return gitdomain.NewGitErr(gitdomain.GitCodeBranchBoundToWorktree, "branch is already assigned to another worktree")
}

// hasRunningTaskOnGitTarget reports whether any task with status running
// references the target id as a worktree or descendant of a repository.
//
//funclogmeasure:skip category=hot-path reason="DB read helper; operation trace is emitted by the calling delete chokepoint."
func hasRunningTaskOnGitTarget(ctx context.Context, db *gorm.DB, targetID string) (bool, error) {
	if targetID == "" {
		return false, nil
	}
	var n int64
	err := db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM tasks
WHERE status = ?
  AND (
        worktree_id = ?
     OR worktree_id IN (
          SELECT id FROM git_worktrees WHERE repository_id = ?
        )
  )`, taskcoredomain.StatusRunning, targetID, targetID).Scan(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

//funclogmeasure:skip category=hot-path reason="DB read helper; operation trace is emitted by the calling delete chokepoint."
func guardNoRunningTask(ctx context.Context, db *gorm.DB, targetID string) error {
	ok, err := hasRunningTaskOnGitTarget(ctx, db, targetID)
	if err != nil {
		return err
	}
	if ok {
		return gitdomain.NewGitErr(gitdomain.GitCodeHasRunningTask, "a task is running on this git target")
	}
	return nil
}
