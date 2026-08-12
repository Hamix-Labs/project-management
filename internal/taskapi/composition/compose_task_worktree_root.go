package composition

import (
	"context"
	"log/slog"
	"strings"

	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

// enrichTaskWorktreeRoot sets WorktreeRootTaskID when the task is bound to a
// Hamix-managed task worktree (branch hamix/task-*).
func (a *API) enrichTaskWorktreeRoot(ctx context.Context, t *taskcoredomain.Task) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.enrichTaskWorktreeRoot")
	if a == nil || t == nil || t.WorktreeID == nil {
		return
	}
	wtID := strings.TrimSpace(*t.WorktreeID)
	if wtID == "" {
		return
	}
	root := a.resolveWorktreeRootTaskID(ctx, wtID)
	t.WorktreeRootTaskID = root
}

func (a *API) enrichTasksWorktreeRoots(ctx context.Context, tasks []taskcoredomain.Task) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.enrichTasksWorktreeRoots")
	if a == nil || len(tasks) == 0 {
		return
	}
	cache := make(map[string]*string)
	for i := range tasks {
		if tasks[i].WorktreeID == nil {
			continue
		}
		wtID := strings.TrimSpace(*tasks[i].WorktreeID)
		if wtID == "" {
			continue
		}
		if root, ok := cache[wtID]; ok {
			tasks[i].WorktreeRootTaskID = root
			continue
		}
		root := a.resolveWorktreeRootTaskID(ctx, wtID)
		cache[wtID] = root
		tasks[i].WorktreeRootTaskID = root
	}
}

func (a *API) resolveWorktreeRootTaskID(ctx context.Context, worktreeID string) *string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.resolveWorktreeRootTaskID")
	if a.git == nil || a.taskcore == nil {
		return nil
	}
	wt, err := a.git.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil || wt.IsMain {
		return nil
	}
	// Prefer worktree name: allocate stamps the root layer branch there and it
	// stays stable when active branch_id moves across stack layers (ADR-0097).
	rootBranch := strings.TrimSpace(wt.Name)
	if !strings.HasPrefix(rootBranch, "hamix/task-") {
		br, brErr := a.git.GetGitBranchByID(ctx, wt.BranchID)
		if brErr != nil {
			return nil
		}
		rootBranch = br.Name
		if !strings.HasPrefix(rootBranch, "hamix/task-") {
			return nil
		}
	}
	family, err := a.taskcore.ListFlat(ctx, 200, 0, &taskcorestore.ListFilter{WorktreeID: &worktreeID})
	if err != nil {
		return nil
	}
	for i := range family {
		if gitinventorystore.TaskBranchName(family[i].ID) == rootBranch {
			id := family[i].ID
			return &id
		}
	}
	return nil
}
