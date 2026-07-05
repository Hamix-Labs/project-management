package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

var (
	handlerSharedGitDir   string
	handlerSharedGitReady bool
	handlerSharedGitOnce  sync.Once
)

// handlerSharedGitRepoDir returns one git repo on disk shared by all handler
// tests. Per-test SQLite DBs still get their own registered worktree rows, but
// we avoid hundreds of git init sequences that made the package exceed CI time.
func handlerSharedGitRepoDir(t *testing.T) string {
	t.Helper()
	handlerSharedGitOnce.Do(func() {
		if _, err := exec.LookPath("git"); err != nil {
			return
		}
		dir, err := os.MkdirTemp("", "hamix-handler-shared-git-*")
		if err != nil {
			panic("handler shared git dir: " + err.Error())
		}
		gittest.EnsureMain(t, dir)
		handlerSharedGitDir = dir
		handlerSharedGitReady = true
	})
	if !handlerSharedGitReady {
		t.Skip("git not on PATH")
	}
	return handlerSharedGitDir
}

type handlerGitBinding struct {
	repositoryID string
	projectID    string
	worktreeID   string
}

var handlerGitBindings sync.Map // baseURL -> handlerGitBinding

func registerHandlerGitBinding(t *testing.T, baseURL string, binding handlerGitBinding) {
	t.Helper()
	handlerGitBindings.Store(baseURL, binding)
	t.Cleanup(func() {
		handlerGitBindings.Delete(baseURL)
	})
}

func handlerGitBindingForURL(baseURL string) (handlerGitBinding, bool) {
	v, ok := handlerGitBindings.Load(baseURL)
	if !ok {
		return handlerGitBinding{}, false
	}
	return v.(handlerGitBinding), true
}

func setHandlerTestGitBinding(t *testing.T, st *store.Store, worktreeID string) handlerGitBinding {
	t.Helper()
	ctx := context.Background()
	wt, err := st.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	defaultProj, err := st.GetDefaultProjectForRepository(ctx, wt.RepositoryID)
	if err != nil {
		t.Fatalf("GetDefaultProjectForRepository: %v", err)
	}
	return handlerGitBinding{
		repositoryID: wt.RepositoryID,
		projectID:    defaultProj.ID,
		worktreeID:   worktreeID,
	}
}

func seedHandlerTestGitRepo(t *testing.T, st *store.Store) handlerGitBinding {
	t.Helper()
	dir := handlerSharedGitRepoDir(t)
	wtID, _ := gittest.SeedWorktree(t, st, dir)
	return setHandlerTestGitBinding(t, st, wtID)
}

func withCreateGitBinding(baseURL, jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	binding, ok := handlerGitBindingForURL(baseURL)
	if !ok {
		return jsonBody
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return jsonBody
	}
	var extra []string
	if !strings.Contains(jsonBody, "project_id") {
		extra = append(extra, `"project_id":"`+binding.projectID+`"`)
	}
	if !strings.Contains(jsonBody, "worktree_id") {
		extra = append(extra, `"worktree_id":"`+binding.worktreeID+`"`)
	}
	if len(extra) == 0 {
		return jsonBody
	}
	return jsonBody[:len(jsonBody)-1] + `,` + strings.Join(extra, ",") + `}`
}

func withCreateChecklistForURL(baseURL, jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	if strings.Contains(jsonBody, "checklist_items") {
		return withCreateGitBinding(baseURL, jsonBody)
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return withCreateGitBinding(baseURL, jsonBody)
	}
	out := jsonBody[:len(jsonBody)-1] + `,"checklist_items":[{"text":"` + testCriterionText + `"}]}`
	return withCreateGitBinding(baseURL, out)
}

func withCreateTaskDefaults(baseURL, jsonBody string) string {
	return withCreateGitBinding(baseURL, withCreateChecklistForURL(baseURL, jsonBody))
}

const directHandlerTestURL = "http://handler-direct.test"

func newBoundTaskServer(t *testing.T, build func(st *store.Store) http.Handler) *httptest.Server {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	binding := seedHandlerTestGitRepo(t, st)
	srv := httptest.NewServer(build(st))
	registerHandlerGitBinding(t, srv.URL, binding)
	t.Cleanup(srv.Close)
	return srv
}

func newDirectBoundHandler(t *testing.T, build func(st *store.Store) http.Handler) http.Handler {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	binding := seedHandlerTestGitRepo(t, st)
	registerHandlerGitBinding(t, directHandlerTestURL, binding)
	return build(st)
}

func mustHandlerGitBinding(t *testing.T, baseURL string) handlerGitBinding {
	t.Helper()
	binding, ok := handlerGitBindingForURL(baseURL)
	if !ok {
		t.Fatal("handler git binding not registered for " + baseURL)
	}
	return binding
}

func boundTaskHandler(st *store.Store, opts ...HandlerOption) http.Handler {
	base := []HandlerOption{WithRepoProvider(NewSettingsRepoProvider(st))}
	return NewHandler(st, NewSSEHub(), nil, append(base, opts...)...)
}
