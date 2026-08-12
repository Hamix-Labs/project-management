package composition

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/storehooks"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// TaskGitContext is the resolved filesystem path and branch name for a task binding.
type TaskGitContext = taskcorecontract.TaskGitContext

// ValidateTaskWorktreeBinding checks worktree_id exists and project/repo alignment.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; git binding traces at storehooks chokepoints."
func (a *API) ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error {
	return storehooks.ValidateTaskWorktreeBinding(ctx, a.gitDeps(), projectID, worktreeID)
}

// ResolveTaskGitContext loads worktree path and branch name via worktree_id.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; git context traces at storehooks chokepoints."
func (a *API) ResolveTaskGitContext(ctx context.Context, worktreeID string) (TaskGitContext, error) {
	return storehooks.ResolveTaskGitContext(ctx, a.gitDeps(), worktreeID)
}

// EnsureTaskStackLayer checks out the task's stack layer branch on its worktree.
func (a *API) EnsureTaskStackLayer(ctx context.Context, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "composition.API.EnsureTaskStackLayer",
		"task_id", taskID)
	if a == nil || a.git == nil || a.taskcore == nil {
		return fmt.Errorf("%w: store not configured", taskcoredomain.ErrInvalidInput)
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("%w: task_id required", taskcoredomain.ErrInvalidInput)
	}
	task, err := a.taskcore.Get(ctx, taskID)
	if err != nil {
		return err
	}
	if task.WorktreeID == nil || strings.TrimSpace(*task.WorktreeID) == "" {
		return fmt.Errorf("%w: worktree_id required", taskcoredomain.ErrInvalidInput)
	}
	return a.git.EnsureTaskStackLayer(ctx, *task.WorktreeID, taskID)
}

// AgentWorkerGitIdle reports whether the worker should stay idle for git registration reasons.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; git idle traces at storehooks chokepoints."
func (a *API) AgentWorkerGitIdle(ctx context.Context) (idle bool, reason string, err error) {
	return storehooks.AgentWorkerGitIdle(ctx, a.git)
}
