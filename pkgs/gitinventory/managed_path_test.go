package gitinventory_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
)

func TestManagedWorktreeRoot_envOverride(t *testing.T) {
	root := t.TempDir()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, root)
	got := gitinventory.ManagedWorktreeRoot()
	if got != filepath.Clean(root) {
		t.Fatalf("ManagedWorktreeRoot=%q want %q", got, filepath.Clean(root))
	}
}

func TestManagedWorktreePath_layout(t *testing.T) {
	root := t.TempDir()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, root)
	repoPath := filepath.Join(string(filepath.Separator), "repos", "acme")
	if runtime.GOOS == "windows" {
		repoPath = `C:\repos\acme`
	}
	got := gitinventory.ManagedWorktreePath(repoPath, "repo-uuid", "hamix/task-abcd1234")
	want := filepath.ToSlash(filepath.Join(root, "worktrees", "repo-uuid", "hamix-task-abcd1234"))
	if got != want {
		t.Fatalf("ManagedWorktreePath=%q want %q", got, want)
	}
	// Path must not live beside the repo parent (.hamix sibling layout).
	if filepath.Dir(filepath.FromSlash(got)) == filepath.Join(filepath.Dir(repoPath), ".hamix", "repo-uuid", "worktrees") {
		t.Fatalf("path still under repo-parent .hamix: %q", got)
	}
}

func TestManagedWorktreePath_ignoresRepoPath(t *testing.T) {
	root := t.TempDir()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, root)
	a := gitinventory.ManagedWorktreePath("/data/proj", "id1", "feature/x")
	b := gitinventory.ManagedWorktreePath("/elsewhere/other", "id1", "feature/x")
	if a != b {
		t.Fatalf("repoPath must not affect layout: %q vs %q", a, b)
	}
	want := filepath.ToSlash(filepath.Join(root, "worktrees", "id1", "feature-x"))
	if a != want {
		t.Fatalf("ManagedWorktreePath=%q want %q", a, want)
	}
}

func TestBranchPathSlug(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"hamix/task-abcd1234", "hamix-task-abcd1234"},
		{"feature/foo", "feature-foo"},
		{"", "branch"},
		{"!!!", "branch"},
		{"a//b", "a-b"},
		{"main", "main"},
	}
	for _, tc := range cases {
		if got := gitinventory.BranchPathSlug(tc.in); got != tc.want {
			t.Fatalf("BranchPathSlug(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
