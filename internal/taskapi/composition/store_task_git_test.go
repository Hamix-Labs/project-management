package composition_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
)

func seedLinkedWorktree(t *testing.T, s *composition.API) string {
	t.Helper()
	ctx := context.Background()
	gitSvc := gitwork.New()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "Test"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	repoRow, err := s.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: dir}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wt, err := s.CreateGitWorktreeForRepo(ctx, repoRow.ID, gitinventorystore.CreateGitWorktreeInput{
		Path:         filepath.Join(filepath.Dir(dir), "wt-feature"),
		Branch:       "feature",
		CreateBranch: true,
		StartPoint:   "main",
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}
	return wt.ID
}

func TestStore_ValidateTaskWorktreeBinding(t *testing.T) {
	ctx := context.Background()
	s := composition.NewAPI(tasktestdb.OpenSQLite(t))
	wtID := seedLinkedWorktree(t, s)
	if err := s.ValidateTaskWorktreeBinding(ctx, nil, wtID); err != nil {
		t.Fatalf("valid binding: %v", err)
	}
	if err := s.ValidateTaskWorktreeBinding(ctx, nil, "00000000-0000-0000-0000-000000000099"); err == nil {
		t.Fatal("expected not found for bogus worktree_id")
	}
}

func TestStore_ValidateTaskWorktreeBinding_refusesMain(t *testing.T) {
	ctx := context.Background()
	s := composition.NewAPI(tasktestdb.OpenSQLite(t))
	gitSvc := gitwork.New()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	repoRow, err := s.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: dir}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wts, err := s.ListGitWorktreesByRepo(ctx, repoRow.ID)
	if err != nil || len(wts) == 0 {
		t.Fatalf("ListGitWorktrees: %v", err)
	}
	var mainID string
	for _, wt := range wts {
		if wt.IsMain {
			mainID = wt.ID
			break
		}
	}
	if mainID == "" {
		t.Fatal("expected main worktree")
	}
	err = s.ValidateTaskWorktreeBinding(ctx, nil, mainID)
	if err == nil {
		t.Fatal("expected refuse main worktree")
	}
}

func TestStore_AgentWorkerGitIdle(t *testing.T) {
	ctx := context.Background()
	s := composition.NewAPI(tasktestdb.OpenSQLite(t))
	idle, reason, err := s.AgentWorkerGitIdle(ctx)
	if err != nil {
		t.Fatalf("AgentWorkerGitIdle: %v", err)
	}
	if !idle || reason != "no_repository_registered" {
		t.Fatalf("got idle=%v reason=%q", idle, reason)
	}
}

func TestStore_ResolveTaskGitContextByWorktree(t *testing.T) {
	ctx := context.Background()
	s := composition.NewAPI(tasktestdb.OpenSQLite(t))
	wtID := seedLinkedWorktree(t, s)
	gitCtx, err := s.ResolveTaskGitContext(ctx, wtID)
	if err != nil {
		t.Fatalf("ResolveTaskGitContext: %v", err)
	}
	if gitCtx.WorktreePath == "" || gitCtx.BranchName == "" {
		t.Fatalf("got %#v", gitCtx)
	}
}
