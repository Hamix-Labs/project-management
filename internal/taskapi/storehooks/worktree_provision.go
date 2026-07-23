package storehooks

import (
	"context"
	"sync"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

// WorktreeProvisioner aliases the canonical taskcore/store hook interface.
type WorktreeProvisioner = taskcorestore.WorktreeProvisioner

// WorktreeProvisionRegistry holds an optional WorktreeProvisioner hook.
type WorktreeProvisionRegistry struct {
	mu sync.RWMutex
	p  WorktreeProvisioner
}

// Set registers p for post-create worktree allocate (nil clears).
//
//funclogmeasure:skip category=hot-path reason="In-process hook registry; provision traces at composition chokepoints."
func (r *WorktreeProvisionRegistry) Set(p WorktreeProvisioner) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.p = p
	r.mu.Unlock()
}

// Enqueue forwards to the registered provisioner (nil-safe).
//
//funclogmeasure:skip category=hot-path reason="In-process hook forwarder; allocate traces at provisioner chokepoints."
func (r *WorktreeProvisionRegistry) Enqueue(ctx context.Context, taskID, repositoryID string) {
	r.mu.RLock()
	p := r.p
	r.mu.RUnlock()
	if p == nil {
		return
	}
	p.Enqueue(ctx, taskID, repositoryID)
}
