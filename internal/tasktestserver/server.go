package tasktestserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
)

// HandlerBuilder wires a task HTTP handler from an opened store (and optional repo root).
type HandlerBuilder func(st *composition.API, workspace *repo.Root) http.Handler

// New returns an httptest.Server using buildHandler with a fresh SQLite store.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func New(t *testing.T, buildHandler HandlerBuilder) *httptest.Server {
	t.Helper()
	_, srv := NewWithStore(t, buildHandler)
	return srv
}

// NewWithStore is like [New] but also returns the store for direct DB setup.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewWithStore(t *testing.T, buildHandler HandlerBuilder) (*composition.API, *httptest.Server) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	return st, httptest.NewServer(buildHandler(st, nil))
}

// NewWithRepo mounts a workspace repo rooted at repoDir.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewWithRepo(t *testing.T, repoDir string, buildHandler HandlerBuilder) *httptest.Server {
	t.Helper()
	srv, _, _, _, _ := NewWithRepoStore(t, repoDir, buildHandler)
	return srv
}

// NewWithRepoStore mounts a workspace repo, seeds git worktree rows, and returns IDs for repo routes.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP wiring; not part of production trace paths."
func NewWithRepoStore(t *testing.T, repoDir string, buildHandler HandlerBuilder) (*httptest.Server, *composition.API, string, string, string) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	worktreeID, branchID := gittest.SeedWorktree(t, st, repoDir)
	r, err := repo.OpenRoot(repoDir)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(buildHandler(st, r)), st, worktreeID, branchID, worktreeID
}

// SeedWorktree seeds git rows for tests that need worktree_id without a full repo server.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only store seed; not part of production trace paths."
func SeedWorktree(t *testing.T, st *composition.API, repoDir string) (worktreeID, branchID string) {
	t.Helper()
	return gittest.SeedWorktree(t, st, repoDir)
}
