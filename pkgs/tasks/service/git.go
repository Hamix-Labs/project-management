package service

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// GitStore is the git persistence surface used by Git orchestration.
type GitStore interface {
	ReconcileGitRepository(ctx context.Context, projectID, repoID string, input store.ReconcileGitInput, gitSvc gitwork.Service) (store.ReconcileGitOutput, error)
	RelocateGitRepository(ctx context.Context, projectID, repoID, path string, gitSvc gitwork.Service) (store.ReconcileGitOutput, error)
	RelocateGitWorktree(ctx context.Context, worktreeID, path string, gitSvc gitwork.Service) (domain.GitWorktree, error)
}

// Git composes gitwork.Service with the store git facade so handlers do not
// pass gitSvc into store methods directly.
type Git struct {
	Store GitStore
	Svc   gitwork.Service
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (g *Git) gitSvc() gitwork.Service {
	if g.Svc != nil {
		return g.Svc
	}
	return gitwork.New()
}

// ReconcileRepository syncs Hamix git rows with on-disk worktrees.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store git facade wrapper; traces emit at store boundary."
func (g *Git) ReconcileRepository(
	ctx context.Context,
	projectID, repoID string,
	input store.ReconcileGitInput,
) (store.ReconcileGitOutput, error) {
	return g.Store.ReconcileGitRepository(ctx, projectID, repoID, input, g.gitSvc())
}

// RelocateRepository moves a registered repository root on disk.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store git facade wrapper; traces emit at store boundary."
func (g *Git) RelocateRepository(
	ctx context.Context,
	projectID, repoID, path string,
) (store.ReconcileGitOutput, error) {
	return g.Store.RelocateGitRepository(ctx, projectID, repoID, path, g.gitSvc())
}

// RelocateWorktree moves a registered worktree path on disk.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store git facade wrapper; traces emit at store boundary."
func (g *Git) RelocateWorktree(
	ctx context.Context,
	worktreeID, path string,
) (domain.GitWorktree, error) {
	return g.Store.RelocateGitWorktree(ctx, worktreeID, path, g.gitSvc())
}
