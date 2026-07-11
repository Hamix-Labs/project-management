package handler

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/service"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// gitStoreAdapter forwards HandlerAPI git methods to service.GitStore.
type gitStoreAdapter struct {
	store.HandlerAPI
}

var _ service.GitStore = gitStoreAdapter{}

//funclogmeasure:skip category=delegate-already-logs reason="Thin HandlerAPI adapter; traces emit at store boundary."
func (a gitStoreAdapter) ReconcileGitRepository(
	ctx context.Context,
	projectID, repoID string,
	input store.ReconcileGitInput,
	gitSvc gitwork.Service,
) (store.ReconcileGitOutput, error) {
	return a.HandlerAPI.ReconcileGitRepository(ctx, projectID, repoID, input, gitSvc)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin HandlerAPI adapter; traces emit at store boundary."
func (a gitStoreAdapter) RelocateGitRepository(
	ctx context.Context,
	projectID, repoID, path string,
	gitSvc gitwork.Service,
) (store.ReconcileGitOutput, error) {
	return a.HandlerAPI.RelocateGitRepository(ctx, projectID, repoID, path, gitSvc)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin HandlerAPI adapter; traces emit at store boundary."
func (a gitStoreAdapter) RelocateGitWorktree(
	ctx context.Context,
	worktreeID, path string,
	gitSvc gitwork.Service,
) (domain.GitWorktree, error) {
	return a.HandlerAPI.RelocateGitWorktree(ctx, worktreeID, path, gitSvc)
}
