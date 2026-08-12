package taskcore_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

type repoFilesPayload struct {
	Paths     []string `json:"paths"`
	NextAfter string   `json:"next_after"`
	HasMore   bool     `json:"has_more"`
	Truncated bool     `json:"truncated"`
	Source    string   `json:"source"`
}

func getRepoFiles(t *testing.T, baseURL, worktreeID string, query url.Values) repoFilesPayload {
	t.Helper()
	if query == nil {
		query = url.Values{}
	}
	query.Set("worktree_id", worktreeID)
	res, err := http.Get(baseURL + "/repo/files?" + query.Encode())
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

	payload := getRepoFiles(t, srv.URL, wtID, url.Values{"limit": []string{"50"}})
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
	if payload.HasMore {
		t.Fatal("small repo should fit in one page")
	}
}

func TestHTTP_repo_files_paginates(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.txt", "d.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	srv, wtID, _ := handlertest.NewBoundRepoServer(t, dir)
	defer srv.Close()

	page1 := getRepoFiles(t, srv.URL, wtID, url.Values{"limit": []string{"2"}})
	if len(page1.Paths) != 2 || !page1.HasMore || page1.NextAfter == "" {
		t.Fatalf("page1 = %+v", page1)
	}
	page2 := getRepoFiles(t, srv.URL, wtID, url.Values{
		"limit": []string{"2"},
		"after": []string{page1.NextAfter},
	})
	if len(page2.Paths) != 2 {
		t.Fatalf("page2 = %+v", page2)
	}
	seen := append(append([]string{}, page1.Paths...), page2.Paths...)
	for _, want := range []string{"a.txt", "b.txt", "c.txt", "d.txt"} {
		if !slices.Contains(seen, want) {
			t.Fatalf("missing %s in %v", want, seen)
		}
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
