package service

import (
	"context"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

// GitStore is the git persistence surface used by Git orchestration.
type GitStore interface {
	ReconcileGitRepository(ctx context.Context, projectID, repoID string, input gitinventorystore.ReconcileGitInput) (gitinventorystore.ReconcileGitOutput, error)
	RelocateGitRepository(ctx context.Context, projectID, repoID, path string) (gitinventorystore.ReconcileGitOutput, error)
	RelocateGitWorktree(ctx context.Context, worktreeID, path string) (domain.GitWorktree, error)
}

// Git is a thin orchestration facade over GitStore (gitwork is injected into the store).
type Git struct {
	Store GitStore
}

// ReconcileRepository syncs Hamix git rows with on-disk worktrees.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store git facade wrapper; traces emit at store boundary."
func (g *Git) ReconcileRepository(
	ctx context.Context,
	projectID, repoID string,
	input gitinventorystore.ReconcileGitInput,
) (gitinventorystore.ReconcileGitOutput, error) {
	return g.Store.ReconcileGitRepository(ctx, projectID, repoID, input)
}

// RelocateRepository moves a registered repository root on disk.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store git facade wrapper; traces emit at store boundary."
func (g *Git) RelocateRepository(
	ctx context.Context,
	projectID, repoID, path string,
) (gitinventorystore.ReconcileGitOutput, error) {
	return g.Store.RelocateGitRepository(ctx, projectID, repoID, path)
}

// RelocateWorktree moves a registered worktree path on disk.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store git facade wrapper; traces emit at store boundary."
func (g *Git) RelocateWorktree(
	ctx context.Context,
	worktreeID, path string,
) (domain.GitWorktree, error) {
	return g.Store.RelocateGitWorktree(ctx, worktreeID, path)
}
