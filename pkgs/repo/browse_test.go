package repo

import (
	"errors"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestFindInstallRoot_findsGoMod(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := FindInstallRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("got %q want %q", got, dir)
	}
}

func TestListBrowseDirs_listsSubdirectories(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	child := filepath.Join(root, "my-project")
	if err := os.MkdirAll(filepath.Join(child, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "node_modules", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".hamix", "repo", "worktrees"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "other"), 0o755); err != nil {
		t.Fatal(err)
	}
	roots := []BrowseRoot{{ID: "home", Path: root, Label: "Home", Available: true}}
	listing, err := ListBrowseDirs(roots, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(listing.Entries) != 2 {
		t.Fatalf("entries = %d want 2 (skip node_modules and .hamix)", len(listing.Entries))
	}
	for _, e := range listing.Entries {
		if e.Name == "node_modules" || e.Name == ".hamix" {
			t.Fatalf("skipped dir listed: %s", e.Name)
		}
	}
	var sawGit bool
	for _, e := range listing.Entries {
		if e.Name == "my-project" {
			sawGit = e.IsGitRepo
		}
	}
	if !sawGit {
		t.Fatal("expected my-project to be marked git repo")
	}
}

func TestListBrowseDirs_rejectsPathOutsideRoots(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	roots := []BrowseRoot{{ID: "home", Path: root, Label: "Home", Available: true}}
	_, err := ListBrowseDirs(roots, outside)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestListBrowseDirs_emptyPathListsRoots(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	roots := []BrowseRoot{{ID: "home", Path: root, Label: "Home", Available: true}}
	listing, err := ListBrowseDirs(roots, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(listing.Entries) != 1 || listing.Entries[0].Path != root {
		t.Fatalf("got %+v", listing.Entries)
	}
}

func TestResolveBrowseRoots_customEnvOverride(t *testing.T) {
	custom := t.TempDir()
	t.Setenv("HAMIX_BROWSE_ROOTS", custom)
	roots, env, err := ResolveBrowseRoots(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if env != BrowseEnvNative {
		t.Fatalf("env = %q", env)
	}
	if len(roots) != 1 || roots[0].Path != custom {
		t.Fatalf("got %+v", roots)
	}
}
