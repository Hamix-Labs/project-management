package store

import "context"

// WorktreeProvisioner eagerly allocates a managed worktree for a task after
// create (ADR-0083). Implementations live outside this package; composition
// registers the hook via internal/taskapi/storehooks.
type WorktreeProvisioner interface {
	// Enqueue schedules allocate for taskID on repositoryID (nil-safe no-op).
	Enqueue(ctx context.Context, taskID, repositoryID string)
	// Stop releases the provisioner loop during process shutdown.
	Stop()
}
