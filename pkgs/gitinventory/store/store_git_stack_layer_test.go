package store

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
	"github.com/google/uuid"
)

func TestEnsureTaskStackLayer_addsChildLayer(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	remote := t.TempDir()
	runGitStore(t, remote, "init", "--bare", "-b", "main")
	parent := t.TempDir()
	main := filepath.Join(parent, "main")
	runGitStore(t, parent, "clone", remote, main)
	runGitStore(t, main, "config", "user.email", "t@example.com")
	runGitStore(t, main, "config", "user.name", "Test")
	runGitStore(t, main, "commit", "--allow-empty", "-m", "init")
	runGitStore(t, main, "push", "-u", "origin", "main")

	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	rootID := uuid.NewString()
	wt, err := s.AllocateTaskWorktree(ctx, repo.ID, rootID)
	if err != nil {
		t.Fatalf("AllocateTaskWorktree: %v", err)
	}
	rootBranch := TaskBranchName(rootID)
	if wt.Name != rootBranch {
		t.Fatalf("worktree name=%q want %q", wt.Name, rootBranch)
	}

	childID := uuid.NewString()
	if err := s.EnsureTaskStackLayer(ctx, wt.ID, childID); err != nil {
		t.Fatalf("EnsureTaskStackLayer: %v", err)
	}
	childBranch := TaskBranchName(childID)
	head, err := gitSvc.WorktreeCurrentBranch(ctx, wt.Path)
	if err != nil {
		t.Fatalf("WorktreeCurrentBranch: %v", err)
	}
	if head != childBranch {
		t.Fatalf("HEAD=%q want %q", head, childBranch)
	}
	got, err := s.GetGitWorktreeByID(ctx, wt.ID)
	if err != nil {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	br, err := s.GetGitBranchByID(ctx, got.BranchID)
	if err != nil {
		t.Fatalf("GetGitBranchByID: %v", err)
	}
	if br.Name != childBranch {
		t.Fatalf("active branch=%q want %q", br.Name, childBranch)
	}
	if got.Name != rootBranch {
		t.Fatalf("worktree name changed to %q; want stable root %q", got.Name, rootBranch)
	}

	// Idempotent for same task.
	if err := s.EnsureTaskStackLayer(ctx, wt.ID, childID); err != nil {
		t.Fatalf("EnsureTaskStackLayer again: %v", err)
	}
}
