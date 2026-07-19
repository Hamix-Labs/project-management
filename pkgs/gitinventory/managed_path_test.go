package gitinventory_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
)

func TestManagedWorktreePath_layout(t *testing.T) {
	repoPath := filepath.Join(string(filepath.Separator), "repos", "acme")
	if runtime.GOOS == "windows" {
		repoPath = `C:\repos\acme`
	}
	got := gitinventory.ManagedWorktreePath(repoPath, "repo-uuid", "hamix/task-abcd1234")
	want := filepath.ToSlash(filepath.Join(filepath.Dir(repoPath), ".hamix", "repo-uuid", "worktrees", "hamix-task-abcd1234"))
	if got != want {
		t.Fatalf("ManagedWorktreePath=%q want %q", got, want)
	}
}

func TestManagedWorktreePath_stableUnderRepoParent(t *testing.T) {
	a := gitinventory.ManagedWorktreePath("/data/proj", "id1", "feature/x")
	b := gitinventory.ManagedWorktreePath("/data/proj/", "id1", "feature/x")
	if a != b {
		t.Fatalf("unstable paths: %q vs %q", a, b)
	}
	if !filepath.IsAbs(filepath.FromSlash(a)) && runtime.GOOS != "windows" {
		// On Unix /data/proj is absolute; path uses ToSlash.
		if a[0] != '/' {
			t.Fatalf("expected absolute-ish path, got %q", a)
		}
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
