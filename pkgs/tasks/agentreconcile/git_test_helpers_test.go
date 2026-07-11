package agentreconcile

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// seedSecondWorktreeOnRepo adds a linked worktree on a new branch in the same repo.
func seedSecondWorktreeOnRepo(t *testing.T, st *store.Store, firstWorktreeID string) (secondWorktreeID string) {
	t.Helper()
	ctx := context.Background()
	wt, err := st.GetGitWorktreeByID(ctx, firstWorktreeID)
	if err != nil {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	repo, err := st.GetGitRepositoryByID(ctx, wt.RepositoryID)
	if err != nil {
		t.Fatalf("GetGitRepositoryByID: %v", err)
	}
	gitSvc := gitwork.New()
	if out, err := exec.Command("git", "-C", repo.Path, "branch", "feature-b").CombinedOutput(); err != nil {
		t.Fatalf("git branch feature-b: %v %s", err, out)
	}
	wt2Path := filepath.Join(filepath.Dir(repo.Path), "wt-feature-b")
	wt2, err := st.CreateGitWorktreeForRepo(ctx, repo.ID, store.CreateGitWorktreeInput{
		Path:         wt2Path,
		Branch:       "feature-b",
		CreateBranch: false,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo feature-b: %v", err)
	}
	return wt2.ID
}
