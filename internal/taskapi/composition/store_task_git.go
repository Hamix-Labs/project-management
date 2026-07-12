package composition

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/storehooks"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
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

// AgentWorkerGitIdle reports whether the worker should stay idle for git registration reasons.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to storehooks; git idle traces at storehooks chokepoints."
func (a *API) AgentWorkerGitIdle(ctx context.Context) (idle bool, reason string, err error) {
	return storehooks.AgentWorkerGitIdle(ctx, a.git)
}
