package composition

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

func (a *API) requireWorktreeRootForOpenPR(ctx context.Context, t *taskcoredomain.Task) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "composition.API.requireWorktreeRootForOpenPR")
	if t == nil {
		return fmt.Errorf("%w: task required", taskcoredomain.ErrInvalidInput)
	}
	if t.WorktreeID == nil || strings.TrimSpace(*t.WorktreeID) == "" {
		return fmt.Errorf("%w: worktree_id required to open PR", taskcoredomain.ErrInvalidInput)
	}
	root := a.resolveWorktreeRootTaskID(ctx, strings.TrimSpace(*t.WorktreeID))
	if root == nil {
		return fmt.Errorf("%w: cannot resolve worktree root for open-pr", taskcoredomain.ErrInvalidInput)
	}
	if *root != t.ID {
		return fmt.Errorf("%w: only the worktree root task may open the stack PR (root=%s)", taskcoredomain.ErrInvalidInput, *root)
	}
	return nil
}

func (a *API) requireNonRootStackLayer(ctx context.Context, t *taskcoredomain.Task) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "composition.API.requireNonRootStackLayer")
	if t == nil {
		return fmt.Errorf("%w: task required", taskcoredomain.ErrInvalidInput)
	}
	if t.WorktreeID == nil || strings.TrimSpace(*t.WorktreeID) == "" {
		return fmt.Errorf("%w: approve from review requires a stacked worktree layer", taskcoredomain.ErrInvalidInput)
	}
	root := a.resolveWorktreeRootTaskID(ctx, strings.TrimSpace(*t.WorktreeID))
	if root == nil {
		return fmt.Errorf("%w: cannot resolve worktree root", taskcoredomain.ErrInvalidInput)
	}
	if *root == t.ID {
		return fmt.Errorf("%w: worktree root must use Approve & Open PR, not Approve from review", taskcoredomain.ErrInvalidInput)
	}
	return nil
}

// ApplyStackPullRequestURLs stamps pull_request_url on family tasks by layer branch name.
func (a *API) ApplyStackPullRequestURLs(ctx context.Context, worktreeID string, urlByBranch map[string]string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "composition.API.ApplyStackPullRequestURLs",
		"worktree_id", worktreeID)
	worktreeID = strings.TrimSpace(worktreeID)
	if a == nil || a.taskcore == nil || worktreeID == "" || len(urlByBranch) == 0 {
		return nil
	}
	family, err := a.taskcore.ListFlat(ctx, 200, 0, &taskcorestore.ListFilter{WorktreeID: &worktreeID})
	if err != nil {
		return err
	}
	for i := range family {
		branch := gitinventorystore.TaskBranchName(family[i].ID)
		url := strings.TrimSpace(urlByBranch[branch])
		if url == "" {
			continue
		}
		u := url
		if _, _, err := a.taskcore.Update(ctx, family[i].ID, taskcorestore.UpdateTaskInput{
			PullRequestURL: &u,
		}, taskcoredomain.ActorAgent); err != nil {
			return fmt.Errorf("stamp pull_request_url for %s: %w", family[i].ID, err)
		}
	}
	return nil
}
