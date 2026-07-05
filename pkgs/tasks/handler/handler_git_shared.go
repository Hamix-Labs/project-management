package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// gitProjectScope is empty for global /git/* routes; non-empty for /projects/{id}/git/*.
type gitProjectScope string

func (h *Handler) serveListGitRepositories(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	out := make([]gitRepositoryJSON, 0)
	if scope == "" {
		rows, err := h.store.ListAllGitRepositoriesWithSummary(r.Context())
		if err != nil {
			writeGitStoreError(w, r, op, err)
			return
		}
		out = make([]gitRepositoryJSON, 0, len(rows))
		for _, row := range rows {
			out = append(out, h.gitRepositorySummaryJSON(row.Repository, row.MainBranchName, row.LinkedWorktreeCount))
		}
	} else {
		rows, err := h.store.ListGitRepositories(r.Context(), string(scope))
		if err != nil {
			writeGitStoreError(w, r, op, err)
			return
		}
		out = make([]gitRepositoryJSON, 0, len(rows))
		for _, row := range rows {
			out = append(out, h.gitRepositoryJSON(row))
		}
	}
	writeJSON(w, r, op, http.StatusOK, gitRepositoriesListResponse{Repositories: out})
}

func (h *Handler) serveCreateGitRepository(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
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
	gitSvc := h.gitService()
	var repo domain.GitRepository
	var err error
	if scope == "" {
		repo, err = h.store.CreateGlobalGitRepository(r.Context(), input, gitSvc)
	} else {
		repo, err = h.store.CreateGitRepository(r.Context(), string(scope), input, gitSvc)
	}
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, h.gitRepositoryJSON(repo))
}

func (h *Handler) serveGetGitRepository(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	var repo domain.GitRepository
	var err error
	if scope == "" {
		repo, err = h.store.GetGitRepositoryByID(r.Context(), repoID)
	} else {
		repo, err = h.store.GetGitRepository(r.Context(), string(scope), repoID)
	}
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, h.gitRepositoryJSON(repo))
}

func (h *Handler) serveDeleteGitRepository(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	var err error
	if scope == "" {
		err = h.store.DeleteGlobalGitRepository(r.Context(), repoID)
	} else {
		err = h.store.DeleteGitRepository(r.Context(), string(scope), repoID)
	}
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) serveListGitWorktrees(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	var rows []domain.GitWorktree
	var err error
	if scope == "" {
		rows, err = h.store.ListGitWorktreesByRepo(r.Context(), repoID)
	} else {
		rows, err = h.store.ListGitWorktrees(r.Context(), string(scope), repoID)
	}
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

func (h *Handler) serveCreateGitWorktree(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
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
	gitSvc := h.gitService()
	var wt domain.GitWorktree
	var err error
	if scope == "" {
		wt, err = h.store.CreateGitWorktreeForRepo(r.Context(), repoID, input, gitSvc)
	} else {
		wt, err = h.store.CreateGitWorktree(r.Context(), string(scope), repoID, input, gitSvc)
	}
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, h.gitWorktreeJSON(wt))
}

func (h *Handler) serveDeleteGitWorktree(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	worktreeID := r.PathValue("worktreeId")
	removeFromDisk, force := gitWorktreeDeleteQuery(r, op)
	gitSvc := h.gitService()
	var err error
	if removeFromDisk {
		if scope == "" {
			err = h.store.RemoveGitWorktreeFromDiskByID(r.Context(), worktreeID, force, gitSvc)
		} else {
			err = h.store.RemoveGitWorktreeFromDisk(r.Context(), string(scope), worktreeID, force, gitSvc)
		}
	} else if scope == "" {
		err = h.store.UnregisterGitWorktreeByID(r.Context(), worktreeID)
	} else {
		err = h.store.UnregisterGitWorktree(r.Context(), string(scope), worktreeID)
	}
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) serveListGitBranches(w http.ResponseWriter, r *http.Request, op string, scope gitProjectScope) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	var rows []domain.GitBranch
	var err error
	if scope == "" {
		rows, err = h.store.ListGitBranchesByRepo(r.Context(), repoID)
	} else {
		rows, err = h.store.ListGitBranches(r.Context(), string(scope), repoID)
	}
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
