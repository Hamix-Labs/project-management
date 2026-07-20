// Package gittest provides shared git repository bootstrap helpers for
// integration and handler tests. Centralizes init, skip-when-missing-git,
// and store git-binding seeding so harness, handler, and worker tests do
// not duplicate exec.Command sequences.
package gittest

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
)

// SkipIfNoGit skips t when the git binary is not on PATH.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func SkipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed; skipping git test")
	}
}

// Init initializes dir as a git repository with an empty initial commit.
// Uses default branch naming and inline user identity for the commit.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func Init(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"-c", "user.email=t@e.local", "-c", "user.name=t", "commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// InitOrSkip calls SkipIfNoGit then Init.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func InitOrSkip(t *testing.T, dir string) {
	t.Helper()
	SkipIfNoGit(t)
	Init(t, dir)
}

// InitMain initializes dir on branch main with user config and an empty commit.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func InitMain(t *testing.T, dir string) {
	t.Helper()
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "test@test.local"},
		{"config", "user.name", "Test"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	if out, err := exec.Command("git", "-C", dir, "commit", "-m", "init", "--allow-empty").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
}

// EnsureMain ensures dir is a git repository on main without re-init when
// already present.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func EnsureMain(t *testing.T, dir string) {
	t.Helper()
	if err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run(); err == nil {
		return
	}
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "test@test.local"},
		{"config", "user.name", "Test"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	_ = exec.Command("git", "-C", dir, "commit", "-m", "init", "--allow-empty").Run()
}

// AttachOrigin creates a bare remote, adds it as origin, and pushes the current
// branch. Call after EnsureMain/InitMain (and any file commits). Required for
// AllocateTaskWorktree.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func AttachOrigin(t *testing.T, dir string) {
	t.Helper()
	SkipIfNoGit(t)
	bare := t.TempDir()
	if out, err := exec.Command("git", "init", "--bare", "-b", "main", bare).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v %s", err, out)
	}
	_ = exec.Command("git", "-C", dir, "remote", "remove", "origin").Run()
	if out, err := exec.Command("git", "-C", dir, "remote", "add", "origin", bare).CombinedOutput(); err != nil {
		t.Fatalf("git remote add: %v %s", err, out)
	}
	_ = exec.Command("git", "-C", dir, "add", "-A").Run()
	_ = exec.Command("git", "-C", dir, "-c", "user.email=test@test.local", "-c", "user.name=Test", "commit", "-m", "seed", "--allow-empty").Run()
	branchOut, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v %s", err, branchOut)
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" || branch == "HEAD" {
		branch = "main"
	}
	if out, err := exec.Command("git", "-C", dir, "push", "-u", "origin", branch).CombinedOutput(); err != nil {
		t.Fatalf("git push origin %s: %v %s", branch, err, out)
	}
}

// SeedWorktree registers repoDir in the store and returns the main worktree id.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func SeedWorktree(t *testing.T, st *composition.API, repoDir string) (worktreeID, branchID string) {
	t.Helper()
	EnsureMain(t, repoDir)
	AttachOrigin(t, repoDir)
	ctx := context.Background()
	repoRow, err := st.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{
		Path: repoDir,
	})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wts, err := st.ListGitWorktreesByRepo(ctx, repoRow.ID)
	if err != nil || len(wts) == 0 {
		t.Fatalf("ListGitWorktrees: %v len=%d", err, len(wts))
	}
	if wts[0].BranchID == "" {
		t.Fatalf("main worktree missing branch_id after CreateGitRepository")
	}
	return wts[0].ID, wts[0].BranchID
}

// SeedWorktreeTemp creates a temp git repo, registers it, and returns a
// non-main linked worktree suitable for task binding (agents cannot run on main).
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git bootstrap; not part of production trace paths."
func SeedWorktreeTemp(t *testing.T, st *composition.API) (worktreeID, workDir string) {
	t.Helper()
	dir := t.TempDir()
	InitMain(t, dir)
	ctx := context.Background()
	repoRow, err := st.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{
		Path: dir,
	})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wtPath := filepath.Join(filepath.Dir(dir), "wt-task")
	wt, err := st.CreateGitWorktreeForRepo(ctx, repoRow.ID, gitinventorystore.CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "hamix/task-test",
		CreateBranch: true,
		StartPoint:   "main",
	})
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}
	if wt.IsMain {
		t.Fatal("SeedWorktreeTemp must not return is_main worktree")
	}
	return wt.ID, wtPath
}
