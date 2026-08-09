package taskcore_test

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

type repoFilesPayload struct {
	Paths     []string `json:"paths"`
	Truncated bool     `json:"truncated"`
	Source    string   `json:"source"`
}

func getRepoFiles(t *testing.T, baseURL, worktreeID string) repoFilesPayload {
	t.Helper()
	res, err := http.Get(baseURL + handlertest.RepoFilesWithWorktree(worktreeID))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("files status %d", res.StatusCode)
	}
	var payload repoFilesPayload
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestHTTP_repo_files_excludes_ignored_paths(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	for _, args := range [][]string{{"init"}, {"config", "user.email", "t@e.com"}, {"config", "user.name", "t"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	for rel, body := range map[string]string{
		".gitignore":   "ignored/\n",
		"note.txt":     "a\nb\n",
		"ignored/.env": "TOKEN=1",
	} {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	srv, wtID, _ := handlertest.NewBoundRepoServer(t, dir)
	defer srv.Close()

	payload := getRepoFiles(t, srv.URL, wtID)
	if payload.Source != "git" {
		t.Fatalf("source = %q, want git", payload.Source)
	}
	if !slices.Contains(payload.Paths, "note.txt") {
		t.Fatalf("expected note.txt in %v", payload.Paths)
	}
	if slices.Contains(payload.Paths, "ignored/.env") {
		t.Fatalf("ignored file was offered for @-mention: %v", payload.Paths)
	}
	if payload.Truncated {
		t.Fatal("a three-file repository must not report truncation")
	}
}

func TestHTTP_repo_files_requires_a_worktree(t *testing.T) {
	srv, _, _ := handlertest.NewBoundRepoServer(t, t.TempDir())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/repo/files")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing worktree_id: status %d want %d", res.StatusCode, http.StatusBadRequest)
	}
}
