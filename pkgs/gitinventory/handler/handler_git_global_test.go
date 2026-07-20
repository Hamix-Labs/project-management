package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
)

func createGlobalGitRepo(t *testing.T, h http.Handler, main string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"path": main})
	req := httptest.NewRequest(http.MethodPost, "/git/repositories", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create repo status=%d body=%s", rec.Code, rec.Body.String())
	}
	var repo gitRepositoryJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &repo); err != nil {
		t.Fatal(err)
	}
	if repo.ID == "" {
		t.Fatal("empty repo id")
	}
	return repo.ID
}

func addHandlerGitWorktree(t *testing.T, main, branch string) string {
	t.Helper()
	wtPath := filepath.Join(filepath.Dir(main), "wt-"+branch)
	runHandlerGit(t, main, "worktree", "add", wtPath, "-b", branch)
	t.Cleanup(func() { _ = os.RemoveAll(wtPath) })
	return wtPath
}

func seedLinkedWorktreeViaStore(t *testing.T, st *composition.API, repoID, main, branch string) gitdomain.GitWorktree {
	t.Helper()
	wtPath := filepath.Join(filepath.Dir(main), "wt-"+branch)
	wt, err := st.CreateGitWorktreeForRepo(context.Background(), repoID, gitinventorystore.CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       branch,
		CreateBranch: true,
		StartPoint:   "main",
	}, gitwork.New())
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(wtPath) })
	return wt
}

func TestHandler_getAndDeleteGlobalGitRepository(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)

	getReq := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID, nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", getRec.Code, getRec.Body.String())
	}
	var repo gitRepositoryJSON
	if err := json.Unmarshal(getRec.Body.Bytes(), &repo); err != nil {
		t.Fatal(err)
	}
	if repo.Path == "" {
		t.Fatal("empty path")
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/git/repositories/"+repoID, nil)
	delRec := httptest.NewRecorder()
	h.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}

	projReq := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID+"/projects", nil)
	projRec := httptest.NewRecorder()
	h.ServeHTTP(projRec, projReq)
	if projRec.Code != http.StatusNotFound && projRec.Code != http.StatusOK {
		t.Fatalf("projects after delete status=%d body=%s", projRec.Code, projRec.Body.String())
	}
	if projRec.Code == http.StatusOK {
		var body struct {
			Projects []any `json:"projects"`
		}
		if err := json.Unmarshal(projRec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Projects) != 0 {
			t.Fatalf("expected no projects after repo delete, got %d", len(body.Projects))
		}
	}
}

func TestHandler_createGlobalGitWorktree_routeRemoved(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	req := httptest.NewRequest(http.MethodPost, "/git/repositories/"+repoID+"/worktrees", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusNotFound {
		t.Fatalf("operator create should be gone: status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandler_listGlobalGitBranches(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)

	listReq := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID+"/branches", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("branches status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var branches gitBranchesListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &branches); err != nil {
		t.Fatal(err)
	}
	if len(branches.Branches) < 1 {
		t.Fatal("expected at least main branch")
	}
}

func TestHandler_listRepoProjects(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	req := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID+"/projects", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("projects status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandler_relocateGlobalGitWorktree(t *testing.T) {
	h, st, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	wt := seedLinkedWorktreeViaStore(t, st, repoID, main, "relocate-wt")
	movedPath := filepath.Join(filepath.Dir(main), "relocate-wt-moved")
	runHandlerGit(t, main, "worktree", "move", wt.Path, movedPath)
	t.Cleanup(func() { _ = os.RemoveAll(movedPath) })

	body, _ := json.Marshal(map[string]string{"path": movedPath})
	req := httptest.NewRequest(http.MethodPost, "/git/worktrees/"+wt.ID+"/relocate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("relocate status=%d body=%s", rec.Code, rec.Body.String())
	}
	var updated gitWorktreeJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if worktreePathKeyHandler(updated.Path) != worktreePathKeyHandler(movedPath) {
		t.Fatalf("path=%q want %q", updated.Path, movedPath)
	}
}

func TestHandler_listGlobalGitRepositories_afterCreate(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	createGlobalGitRepo(t, h, main)
	req := httptest.NewRequest(http.MethodGet, "/git/repositories", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp gitRepositoriesListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Repositories) != 1 {
		t.Fatalf("len=%d want 1", len(resp.Repositories))
	}
	if resp.Repositories[0].MainBranchName == "" {
		t.Fatal("expected main_branch_name on list response")
	}
	if resp.Repositories[0].LinkedWorktreeCount != 0 {
		t.Fatalf("linked_worktree_count=%d want 0", resp.Repositories[0].LinkedWorktreeCount)
	}
}

func TestHandler_listGlobalGitWorktrees_serializesBranchID(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)

	req := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID+"/worktrees", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp gitWorktreesListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Worktrees) != 1 {
		t.Fatalf("len=%d want 1 main worktree", len(resp.Worktrees))
	}
	if resp.Worktrees[0].BranchID == "" {
		t.Fatal("main worktree must have branch_id after global register")
	}
}

func TestHandler_gitErrHTTP_domainSentinels(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"not found", taskcoredomain.ErrNotFound, http.StatusNotFound},
		{"invalid input", taskcoredomain.ErrInvalidInput, http.StatusBadRequest},
		{"conflict", taskcoredomain.ErrConflict, http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, _ := GitErrHTTP(tt.err)
			if status != tt.status || code != "" {
				t.Fatalf("status=%d code=%q want status=%d code=\"\"", status, code, tt.status)
			}
		})
	}
}

func TestHandler_listGlobalGitWorktreesCheckoutStatus(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)

	req := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID+"/worktrees/checkout-status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("checkout-status status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out gitWorktreeCheckoutStatusListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Worktrees) == 0 {
		t.Fatal("expected at least main worktree")
	}
	found := false
	for _, row := range out.Worktrees {
		if !row.Available {
			t.Fatalf("main should be available: %+v", row)
		}
		if row.Dirty {
			t.Fatalf("expected clean main: %+v", row)
		}
		if row.HeadCommitAt == "" {
			t.Fatalf("expected head_commit_at: %+v", row)
		}
		found = true
	}
	if !found {
		t.Fatal("no available rows")
	}

	if err := os.WriteFile(filepath.Join(main, "dirty-status.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/git/repositories/"+repoID+"/worktrees/checkout-status", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("checkout-status dirty status=%d body=%s", rec2.Code, rec2.Body.String())
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Worktrees) == 0 || !out.Worktrees[0].Dirty {
		t.Fatalf("expected dirty main: %+v", out.Worktrees)
	}
}

func TestHandler_operatorLiveProbeRegisterRoutesRemoved(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	paths := []string{
		"/git/repositories/" + repoID + "/worktrees/live",
		"/git/repositories/" + repoID + "/worktrees/probe?path=" + url.QueryEscape(main),
		"/git/repositories/" + repoID + "/worktrees/register",
		"/git/repositories/" + repoID + "/branches/live",
	}
	for _, p := range paths {
		method := http.MethodGet
		if strings.HasSuffix(p, "/register") {
			method = http.MethodPost
		}
		req := httptest.NewRequest(method, p, strings.NewReader("{}"))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s %s status=%d want 404/405 body=%s", method, p, rec.Code, rec.Body.String())
		}
	}
}

func TestHandler_syncGlobalGitRepository(t *testing.T) {
	h, st, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	_ = seedLinkedWorktreeViaStore(t, st, repoID, main, "sync-wt")
	// Local-only repo has no origin; sync should fail closed with actionable error.
	req := httptest.NewRequest(http.MethodPost, "/git/repositories/"+repoID+"/sync", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusAccepted {
		return
	}
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusConflict {
		t.Fatalf("sync status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandler_reconcileGlobalGitRepository(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	renamed := filepath.Join(filepath.Dir(main), "reconcile-gone")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(renamed); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/git/repositories/"+repoID+"/reconcile", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("reconcile status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out gitReconcileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != "needs_bootstrap_path" {
		t.Fatalf("status=%q want needs_bootstrap_path", out.Status)
	}
}

func TestHandler_reconcileGlobalGitRepository_dryRunOK(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	req := httptest.NewRequest(http.MethodPost, "/git/repositories/"+repoID+"/reconcile", strings.NewReader(`{"dry_run":true}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("reconcile status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out gitReconcileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != "ok" {
		t.Fatalf("status=%q want ok", out.Status)
	}
}

func TestHandler_relocateGlobalGitRepository(t *testing.T) {
	h, _, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	renamed := filepath.Join(filepath.Dir(main), "relocate-http")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, main) })

	body, _ := json.Marshal(map[string]string{"path": renamed})
	req := httptest.NewRequest(http.MethodPost, "/git/repositories/"+repoID+"/relocate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("relocate status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out gitReconcileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != "ok" || !out.Report.RepoPathUpdated {
		t.Fatalf("relocate response=%+v", out)
	}
}

func TestHandler_deleteGlobalGitWorktree(t *testing.T) {
	h, st, main := gitHandlerTest(t)
	repoID := createGlobalGitRepo(t, h, main)
	wt := seedLinkedWorktreeViaStore(t, st, repoID, main, "delete-http")

	delReq := httptest.NewRequest(http.MethodDelete, "/git/worktrees/"+wt.ID+"?force=true", nil)
	delRec := httptest.NewRecorder()
	h.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}
}

func worktreePathKeyHandler(path string) string {
	return strings.ToLower(strings.TrimSuffix(strings.ReplaceAll(filepath.Clean(path), `\`, `/`), `/`))
}

func TestHandler_gitStoreErrorsReturnStableCode(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{
			name:   "repository not found",
			err:    gitdomain.NewGitErr(gitdomain.GitCodeRepositoryNotFound, "repository not found"),
			status: http.StatusNotFound,
			code:   gitdomain.GitCodeRepositoryNotFound,
		},
		{
			name:   "path exists",
			err:    gitdomain.NewGitErr(gitdomain.GitCodePathExists, "worktree path already registered"),
			status: http.StatusConflict,
			code:   gitdomain.GitCodePathExists,
		},
		{
			name:   "bootstrap mismatch",
			err:    gitdomain.NewGitErr(gitdomain.GitCodeBootstrapMismatch, "bootstrap mismatch"),
			status: http.StatusConflict,
			code:   gitdomain.GitCodeBootstrapMismatch,
		},
		{
			name:   "has running task",
			err:    gitdomain.NewGitErr(gitdomain.GitCodeHasRunningTask, "has running task"),
			status: http.StatusConflict,
			code:   gitdomain.GitCodeHasRunningTask,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, msg := GitErrHTTP(tt.err)
			if status != tt.status || code != tt.code || msg == "" {
				t.Fatalf("status=%d code=%q msg=%q", status, code, msg)
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			WriteGitStoreError(rec, req, "test.op", tt.err)
			if rec.Code != tt.status {
				t.Fatalf("write status=%d want %d body=%s", rec.Code, tt.status, rec.Body.String())
			}
			var body apijson.JSONCodedErrorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tt.code {
				t.Fatalf("body code=%q want %q", body.Code, tt.code)
			}
		})
	}
}
