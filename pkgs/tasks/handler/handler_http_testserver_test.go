package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestserver"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
)

func taskTestHandlerBuilder(st *composition.API, workspace *repo.Root) http.Handler {
	return NewHandler(st, NewSSEHub(), workspace, WithRepoProvider(NewSettingsRepoProvider(st)))
}

func newTaskTestHandler(st *composition.API, opts ...HandlerOption) http.Handler {
	return newTaskTestHandlerWithHub(st, NewSSEHub(), opts...)
}

func newTaskTestHandlerWithHub(st *composition.API, hub *SSEHub, opts ...HandlerOption) http.Handler {
	base := []HandlerOption{WithRepoProvider(NewSettingsRepoProvider(st))}
	return NewHandler(st, hub, nil, append(base, opts...)...)
}

func newTaskTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return tasktestserver.New(t, taskTestHandlerBuilder)
}

func newTaskCreateTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	st, srv := tasktestserver.NewWithStore(t, taskTestHandlerBuilder)
	binding := seedHandlerTestGitRepo(t, st)
	registerHandlerGitBinding(t, srv.URL, binding)
	return srv
}

func newTaskTestServerWithStore(t *testing.T) (*httptest.Server, *composition.API) {
	t.Helper()
	st, srv := tasktestserver.NewWithStore(t, taskTestHandlerBuilder)
	return srv, st
}

func newTaskCreateTestServerWithStore(t *testing.T) (*httptest.Server, *composition.API) {
	t.Helper()
	st, srv := tasktestserver.NewWithStore(t, taskTestHandlerBuilder)
	binding := seedHandlerTestGitRepo(t, st)
	registerHandlerGitBinding(t, srv.URL, binding)
	return srv, st
}

func newTaskCreateTestServerWithHub(t *testing.T, hub *SSEHub, opts ...HandlerOption) (*httptest.Server, *composition.API) {
	t.Helper()
	st, srv := tasktestserver.NewWithStore(t, func(s *composition.API, workspace *repo.Root) http.Handler {
		return newTaskTestHandlerWithHub(s, hub, opts...)
	})
	binding := seedHandlerTestGitRepo(t, st)
	registerHandlerGitBinding(t, srv.URL, binding)
	return srv, st
}

func newTaskCreateTestServerFromStore(t *testing.T, st *composition.API, opts ...HandlerOption) *httptest.Server {
	t.Helper()
	binding := seedHandlerTestGitRepo(t, st)
	srv := httptest.NewServer(newTaskTestHandler(st, opts...))
	registerHandlerGitBinding(t, srv.URL, binding)
	return srv
}

func seedTestGitWorktree(t *testing.T, st *composition.API, repoDir string) (worktreeID, branchID string) {
	t.Helper()
	return tasktestserver.SeedWorktree(t, st, repoDir)
}

func newTaskTestServerWithRepo(t *testing.T, repoDir string) (*httptest.Server, string, string) {
	srv, _, wt, br := newTaskTestServerWithRepoStore(t, repoDir)
	return srv, wt, br
}

func newTaskTestServerWithRepoStore(t *testing.T, repoDir string) (*httptest.Server, *composition.API, string, string) {
	t.Helper()
	srv, st, wt, br, _ := tasktestserver.NewWithRepoStore(t, repoDir, taskTestHandlerBuilder)
	binding := setHandlerTestGitBinding(t, st, wt)
	registerHandlerGitBinding(t, srv.URL, binding)
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
