package composition

import (
	"context"
	"strings"

	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

// enrichTaskWorktreeRoot sets WorktreeRootTaskID when the task is bound to a
// Hamix-managed task worktree (branch hamix/task-*).
func (a *API) enrichTaskWorktreeRoot(ctx context.Context, t *taskcoredomain.Task) {
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
	if a.git == nil || a.taskcore == nil {
		return nil
	}
	wt, err := a.git.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil || wt.IsMain {
		return nil
	}
	br, err := a.git.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		return nil
	}
	if !strings.HasPrefix(br.Name, "hamix/task-") {
		return nil
	}
	family, err := a.taskcore.ListFlat(ctx, 200, 0, &taskcorestore.ListFilter{WorktreeID: &worktreeID})
	if err != nil {
		return nil
	}
	for i := range family {
		if gitinventorystore.TaskBranchName(family[i].ID) == br.Name {
			id := family[i].ID
			return &id
		}
	}
	return nil
}
