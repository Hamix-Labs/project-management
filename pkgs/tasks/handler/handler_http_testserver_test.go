package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestserver"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func taskTestHandlerBuilder(st *store.Store, workspace *repo.Root) http.Handler {
	opts := []HandlerOption{}
	if workspace != nil {
		opts = append(opts, WithRepoProvider(NewSettingsRepoProvider(st)))
	}
	return NewHandler(st, NewSSEHub(), workspace, opts...)
}

func newTaskTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return tasktestserver.New(t, taskTestHandlerBuilder)
}

func newTaskTestServerWithStore(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	st, srv := tasktestserver.NewWithStore(t, taskTestHandlerBuilder)
	return srv, st
}

func seedTestGitWorktree(t *testing.T, st *store.Store, repoDir string) (worktreeID, branchID string) {
	t.Helper()
	return tasktestserver.SeedWorktree(t, st, repoDir)
}

func newTaskTestServerWithRepo(t *testing.T, repoDir string) (*httptest.Server, string, string) {
	srv, _, wt, br, _ := tasktestserver.NewWithRepoStore(t, repoDir, taskTestHandlerBuilder)
	return srv, wt, br
}

func newTaskTestServerWithRepoStore(t *testing.T, repoDir string) (*httptest.Server, *store.Store, string, string) {
	t.Helper()
	srv, st, wt, br, _ := tasktestserver.NewWithRepoStore(t, repoDir, taskTestHandlerBuilder)
	return srv, st, wt, br
}

func repoPathWithWorktree(worktreeID, path string) string {
	q := url.Values{}
	q.Set("worktree_id", worktreeID)
	if path != "" {
		q.Set("path", path)
	}
	return "/repo/file?" + q.Encode()
}

func repoSearchWithWorktree(worktreeID, q string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	v.Set("q", q)
	return "/repo/search?" + v.Encode()
}

func repoValidateRangeWithWorktree(worktreeID, path, start, end string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	v.Set("path", path)
	v.Set("start", start)
	v.Set("end", end)
	return "/repo/validate-range?" + v.Encode()
}

func repoDiffWithWorktree(worktreeID, sha string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	v.Set("sha", sha)
	return "/repo/diff?" + v.Encode()
}
