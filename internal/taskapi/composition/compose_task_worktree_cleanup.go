package composition

import (
	"context"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// bestEffortRemoveTaskWorktree removes a Hamix-managed task worktree (+ matching
// hamix/task-* branch) after the task row is gone. Failures are logged only —
// task delete already succeeded.
//
// When no tasks remain on the worktree, any last remaining binder (allocator or
// sharer) may trigger cleanup for branches named hamix/task-*. Requiring the
// deleted task id to match the branch name orphans the checkout when the
// allocator is deleted first.
func (a *API) bestEffortRemoveTaskWorktree(ctx context.Context, taskID, worktreeID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.bestEffortRemoveTaskWorktree")
	worktreeID = strings.TrimSpace(worktreeID)
	taskID = strings.TrimSpace(taskID)
	if a == nil || a.git == nil || a.taskcore == nil || worktreeID == "" || taskID == "" {
		return
	}
	wt, err := a.git.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		slog.Warn("task delete worktree cleanup: load worktree", "task_id", taskID, "worktree_id", worktreeID, "err", err)
		return
	}
	if wt.IsMain {
		return
	}
	br, err := a.git.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		slog.Warn("task delete worktree cleanup: load branch", "task_id", taskID, "worktree_id", worktreeID, "branch_id", wt.BranchID, "err", err)
		return
	}
	if !strings.HasPrefix(br.Name, "hamix/task-") {
		return
	}
	n, err := a.taskcore.CountTasksByWorktreeID(ctx, worktreeID)
	if err != nil {
		slog.Warn("task delete worktree cleanup: count remaining tasks", "task_id", taskID, "worktree_id", worktreeID, "err", err)
		return
	}
	if n > 0 {
		return
	}
	if err := a.git.RemoveGitWorktreeFromDiskByID(ctx, worktreeID, true); err != nil {
		slog.Warn("task delete worktree cleanup: remove from disk", "task_id", taskID, "worktree_id", worktreeID, "err", err)
		return
	}
	if err := a.git.DeleteGitBranchByID(ctx, br.ID, true); err != nil {
		slog.Warn("task delete worktree cleanup: delete branch", "task_id", taskID, "worktree_id", worktreeID, "branch_id", br.ID, "err", err)
	}
}
