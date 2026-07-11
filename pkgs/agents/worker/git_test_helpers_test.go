package worker_test

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func seedWorkerTestGit(t *testing.T, st *store.Store) (worktreeID, workDir string) {
	t.Helper()
	return gittest.SeedWorktreeTemp(t, st)
}

func (h *harness) gitBinding() *string {
	wt := h.worktreeID
	return &wt
}

func (h *harness) repositoryID() string {
	h.t.Helper()
	wt, err := h.store.GetGitWorktreeByID(context.Background(), h.worktreeID)
	if err != nil {
		h.t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	return wt.RepositoryID
}
