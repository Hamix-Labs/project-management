package handlertest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestserver"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
)

//funclogmeasure:skip category=tool-required-noop reason="Test-only handler wiring; not part of production trace paths."
func buildHandler(st *composition.API, workspace *repo.Root) http.Handler {
	opts := []handler.HandlerOption{}
	if workspace != nil {
		opts = append(opts, handler.WithRepoProvider(handler.NewSettingsRepoProvider(st)))
	}
	return handler.NewHandler(st, handler.NewSSEHub(), workspace, opts...)
}

// NewServer returns an httptest.Server wrapping handler.NewHandler with SQLite,
// SSE hub, and no workspace repo.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewServer(t *testing.T) *httptest.Server {
	t.Helper()
	return tasktestserver.New(t, buildHandler)
}

// NewServerWithStore is like [NewServer] but also returns the store for direct DB setup.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewServerWithStore(t *testing.T) (*httptest.Server, *composition.API) {
	t.Helper()
	st, srv := tasktestserver.NewWithStore(t, buildHandler)
	return srv, st
}

// NewServerWithRepo is like [NewServer] but mounts a workspace repo rooted at repoDir.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewServerWithRepo(t *testing.T, repoDir string) *httptest.Server {
	t.Helper()
	return tasktestserver.NewWithRepo(t, repoDir, buildHandler)
}

// NewServerWithRepoStore mounts a workspace repo, seeds git worktree rows, and returns IDs for repo routes.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewServerWithRepoStore(t *testing.T, repoDir string) (*httptest.Server, *composition.API, string, string, string) {
	t.Helper()
	return tasktestserver.NewWithRepoStore(t, repoDir, buildHandler)
}

// SeedGitWorktree seeds git rows for tests that need worktree_id without a full
// bound create server.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only store seed; not part of production trace paths."
func SeedGitWorktree(t *testing.T, st *composition.API, repoDir string) (worktreeID, branchID string) {
	t.Helper()
	return tasktestserver.SeedWorktree(t, st, repoDir)
}

//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewBoundRepoServer(t *testing.T, repoDir string) (*httptest.Server, string, string) {
	t.Helper()
	srv, _, wt, br := NewBoundRepoServerWithStore(t, repoDir)
	return srv, wt, br
}

// NewBoundRepoServerWithStore is like handler's newTaskTestServerWithRepoStore.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewBoundRepoServerWithStore(t *testing.T, repoDir string) (*httptest.Server, *composition.API, string, string) {
	t.Helper()
	srv, st, wt, br, _ := NewServerWithRepoStore(t, repoDir)
	RegisterGitBinding(t, srv.URL, htGitBindingFromWorktree(t, st, wt))
	return srv, st, wt, br
}
