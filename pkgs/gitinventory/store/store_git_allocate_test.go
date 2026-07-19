package store

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
	"github.com/google/uuid"
)

func TestTaskBranchName(t *testing.T) {
	id := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	got := TaskBranchName(id)
	if got != "hamix/task-a1b2c3d4" {
		t.Fatalf("TaskBranchName=%q", got)
	}
}

func TestAllocateTaskWorktree(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	remote := t.TempDir()
	runGitStore(t, remote, "init", "--bare", "-b", "main")
	parent := t.TempDir()
	main := filepath.Join(parent, "main")
	runGitStore(t, parent, "clone", remote, main)
	runGitStore(t, main, "config", "user.email", "t@example.com")
	runGitStore(t, main, "config", "user.name", "Test")
	runGitStore(t, main, "commit", "--allow-empty", "-m", "init")
	runGitStore(t, main, "push", "-u", "origin", "main")

	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	taskID := uuid.NewString()
	wt, err := s.AllocateTaskWorktree(ctx, repo.ID, taskID, gitSvc)
	if err != nil {
		t.Fatalf("AllocateTaskWorktree: %v", err)
	}
	if wt.IsMain {
		t.Fatal("allocated worktree must not be is_main")
	}
	wantBranch := TaskBranchName(taskID)
	wantPath := gitinventory.ManagedWorktreePath(repo.Path, repo.ID, wantBranch)
	if filepath.ToSlash(wt.Path) != wantPath {
		t.Fatalf("path=%q want %q", wt.Path, wantPath)
	}
	br, err := s.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		t.Fatalf("GetGitBranchByID: %v", err)
	}
	if br.Name != wantBranch {
		t.Fatalf("branch=%q want %q", br.Name, wantBranch)
	}
	if strings.EqualFold(br.Name, repo.DefaultBranch) {
		t.Fatal("allocated branch must not be default")
	}
}

func TestAllocateTaskWorktree_fetchFailure(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	_, err = s.AllocateTaskWorktree(ctx, repo.ID, uuid.NewString(), gitSvc)
	if err == nil {
		t.Fatal("expected fetch failure without origin")
	}
	if !strings.Contains(err.Error(), "fetch") {
		t.Fatalf("error should mention fetch: %v", err)
	}
}
