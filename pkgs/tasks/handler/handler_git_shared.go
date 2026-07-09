package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func (h *Handler) serveListGitRepositories(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	rows, err := h.store.ListAllGitRepositoriesWithSummary(r.Context())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitRepositoryJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, h.gitRepositorySummaryJSON(row.Repository, row.MainBranchName, row.LinkedWorktreeCount))
	}
	writeJSON(w, r, op, http.StatusOK, gitRepositoriesListResponse{Repositories: out})
}

func (h *Handler) serveCreateGitRepository(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	var body gitRepositoryCreateJSON
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	input := store.CreateGitRepositoryInput{
		Path:          body.Path,
		HostPath:      body.HostPath,
		DefaultBranch: body.DefaultBranch,
	}
	repo, err := h.store.CreateGlobalGitRepository(r.Context(), input, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, h.gitRepositoryJSON(repo))
}

func (h *Handler) serveGetGitRepository(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	repo, err := h.store.GetGitRepositoryByID(r.Context(), repoID)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, h.gitRepositoryJSON(repo))
}

func (h *Handler) serveDeleteGitRepository(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	if err := h.store.DeleteGlobalGitRepository(r.Context(), repoID); err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) serveListGitWorktrees(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	rows, err := h.store.ListGitWorktreesByRepo(r.Context(), repoID)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitWorktreeJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, h.gitWorktreeJSON(row))
	}
	writeJSON(w, r, op, http.StatusOK, gitWorktreesListResponse{Worktrees: out})
}

func (h *Handler) serveCreateGitWorktree(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	var body gitWorktreeCreateJSON
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	input := store.CreateGitWorktreeInput{
		Path:         body.Path,
		Name:         body.Name,
		Branch:       body.Branch,
		CreateBranch: body.CreateBranch,
		StartPoint:   body.StartPoint,
		ForceRemove:  body.ForceRemove,
	}
	repoID := r.PathValue("repoId")
	wt, err := h.store.CreateGitWorktreeForRepo(r.Context(), repoID, input, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, h.gitWorktreeJSON(wt))
}

func (h *Handler) serveDeleteGitWorktree(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	worktreeID := r.PathValue("worktreeId")
	removeFromDisk, force := gitWorktreeDeleteQuery(r, op)
	gitSvc := h.gitService()
	var err error
	if removeFromDisk {
		err = h.store.RemoveGitWorktreeFromDiskByID(r.Context(), worktreeID, force, gitSvc)
	} else {
		err = h.store.UnregisterGitWorktreeByID(r.Context(), worktreeID)
	}
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) serveListGitBranches(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	rows, err := h.store.ListGitBranchesByRepo(r.Context(), repoID)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitBranchJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, toGitBranchJSON(row))
	}
	writeJSON(w, r, op, http.StatusOK, gitBranchesListResponse{Branches: out})
}
