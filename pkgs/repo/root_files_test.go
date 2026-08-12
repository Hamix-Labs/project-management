package repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	abs := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestRoot_Files_honors_gitignore(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, ".gitignore", "secrets/\n*.log\n")
	writeFile(t, dir, "web/src/main.tsx", "x")
	writeFile(t, dir, "secrets/.env", "TOKEN=1")
	writeFile(t, dir, "build.log", "noise")

	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	listing, err := r.Files(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if listing.Source != FileListSourceGit {
		t.Fatalf("source = %q, want git", listing.Source)
	}
	if !slices.Contains(listing.Paths, "web/src/main.tsx") {
		t.Fatalf("expected the tracked-or-new source file, got %v", listing.Paths)
	}
	for _, ignored := range []string{"secrets/.env", "build.log"} {
		if slices.Contains(listing.Paths, ignored) {
			t.Fatalf("ignored path %q leaked into the listing: %v", ignored, listing.Paths)
		}
	}
	if slices.Contains(listing.Paths, ".git/config") {
		t.Fatalf("git internals leaked into the listing: %v", listing.Paths)
	}
}

func TestRoot_Files_lists_tracked_but_ignored_files(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "config.local", "kept")
	add := exec.Command("git", "-C", dir, "add", "config.local")
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	writeFile(t, dir, ".gitignore", "config.local\n")

	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	listing, err := r.Files(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// A file already under version control stays referenceable even once a
	// rule would ignore it, which is what --cached buys.
	if !slices.Contains(listing.Paths, "config.local") {
		t.Fatalf("tracked-but-ignored file missing from %v", listing.Paths)
	}
}

func TestRoot_Files_falls_back_to_a_walk_outside_git(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "notes/todo.md", "x")
	writeFile(t, dir, "node_modules/pkg/index.js", "x")

	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	listing, err := r.Files(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if listing.Source != FileListSourceWalk {
		t.Fatalf("source = %q, want walk", listing.Source)
	}
	if !slices.Contains(listing.Paths, "notes/todo.md") {
		t.Fatalf("expected the plain file, got %v", listing.Paths)
	}
	if slices.Contains(listing.Paths, "node_modules/pkg/index.js") {
		t.Fatalf("walk fallback should still skip dependency directories: %v", listing.Paths)
	}
}

func TestRoot_Files_returns_sorted_paths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "z.md", "x")
	writeFile(t, dir, "a.md", "x")
	writeFile(t, dir, "m/b.md", "x")

	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	listing, err := r.Files(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.IsSorted(listing.Paths) {
		t.Fatalf("paths are not sorted: %v", listing.Paths)
	}
}

func TestRoot_FilesPage_cursor_and_filter(t *testing.T) {
	resetFileListCacheForTest()
	dir := t.TempDir()
	for _, rel := range []string{"a.go", "b.go", "c_test.go", "readme.md"} {
		writeFile(t, dir, rel, "x")
	}
	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}

	page1, err := r.FilesPage(context.Background(), "", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.Paths) != 2 || !page1.HasMore || page1.NextAfter == "" {
		t.Fatalf("page1 = %+v", page1)
	}

	page2, err := r.FilesPage(context.Background(), "", page1.NextAfter, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Paths) != 2 {
		t.Fatalf("page2 paths = %v", page2.Paths)
	}
	if page2.HasMore {
		t.Fatalf("expected no more after second page: %+v", page2)
	}

	filtered, err := r.FilesPage(context.Background(), "test", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(filtered.Paths, "c_test.go") {
		t.Fatalf("filter missed c_test.go: %v", filtered.Paths)
	}
	for _, p := range filtered.Paths {
		if !strings.Contains(strings.ToLower(p), "test") {
			t.Fatalf("unfiltered path %q in %v", p, filtered.Paths)
		}
	}
}
