package gitexec_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitexec"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func seedFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	abs := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListFiles_lists_tracked_and_untracked_but_not_ignored(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	git(t, dir, "init")
	git(t, dir, "config", "user.email", "t@e.com")
	git(t, dir, "config", "user.name", "t")
	seedFile(t, dir, ".gitignore", "build/\n")
	seedFile(t, dir, "tracked.go", "package main")
	git(t, dir, "add", "tracked.go")
	seedFile(t, dir, "untracked.go", "package main")
	seedFile(t, dir, "build/out.bin", "x")

	paths, truncated, err := gitexec.ListFiles(context.Background(), dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if truncated {
		t.Fatal("no limit was set, so nothing should be truncated")
	}
	for _, want := range []string{"tracked.go", "untracked.go", ".gitignore"} {
		if !slices.Contains(paths, want) {
			t.Fatalf("expected %q in %v", want, paths)
		}
	}
	if slices.Contains(paths, "build/out.bin") {
		t.Fatalf("ignored path leaked: %v", paths)
	}
}

func TestListFiles_reports_truncation_at_the_limit(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	git(t, dir, "init")
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		seedFile(t, dir, name, "x")
	}

	paths, truncated, err := gitexec.ListFiles(context.Background(), dir, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !truncated {
		t.Fatal("expected truncation at a limit below the file count")
	}
	if len(paths) != 2 {
		t.Fatalf("got %d paths, want 2: %v", len(paths), paths)
	}
}

func TestListFiles_signals_no_work_tree(t *testing.T) {
	requireGit(t)

	t.Run("plain directory", func(t *testing.T) {
		dir := t.TempDir()
		seedFile(t, dir, "note.txt", "x")
		_, _, err := gitexec.ListFiles(context.Background(), dir, 0)
		if !errors.Is(err, gitexec.ErrNoWorkTree) {
			t.Fatalf("err = %v, want ErrNoWorkTree", err)
		}
	})

	t.Run("bare repository", func(t *testing.T) {
		dir := t.TempDir()
		git(t, dir, "init", "--bare")
		_, _, err := gitexec.ListFiles(context.Background(), dir, 0)
		if !errors.Is(err, gitexec.ErrNoWorkTree) {
			t.Fatalf("err = %v, want ErrNoWorkTree", err)
		}
	})
}
