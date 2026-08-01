package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

func (h *Handler) serveListGitRepositories(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	rows, err := h.inventory.ListAllGitRepositoriesWithSummary(r.Context())
	if err != nil {
		WriteGitStoreError(w, r, op, err)
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
	input := contract.CreateGitRepositoryInput{
		Path:          body.Path,
		HostPath:      body.HostPath,
		DefaultBranch: body.DefaultBranch,
	}
	repo, err := h.write.CreateGlobalGitRepository(r.Context(), input)
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, h.gitRepositoryJSON(repo))
}

func (h *Handler) serveGetGitRepository(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	repo, err := h.inventory.GetGitRepositoryByID(r.Context(), repoID)
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, h.gitRepositoryJSON(repo))
}

func (h *Handler) serveDeleteGitRepository(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	if err := h.inventory.DeleteGlobalGitRepository(r.Context(), repoID); err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) serveListGitWorktrees(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	rows, err := h.inventory.ListGitWorktreesByRepo(r.Context(), repoID)
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	staleMap, err := h.inventory.WorktreeStaleMap(r.Context(), rows, time.Now().UTC())
	if err != nil {
		// Stale hints are enrichment only; do not fail the worktree list (UI
		// bindings depend on this endpoint returning 200).
		slog.Warn("worktree stale map failed; continuing without stale flags",
			"cmd", calltrace.LogCmd, "operation", op+".stale_map_err",
			"repository_id", repoID, "err", err)
		staleMap = map[string]bool{}
	}
	out := make([]gitWorktreeJSON, 0, len(rows))
	for _, row := range rows {
		j := h.gitWorktreeJSON(row)
		j.Stale = staleMap[row.ID]
		out = append(out, j)
	}
	writeJSON(w, r, op, http.StatusOK, gitWorktreesListResponse{Worktrees: out})
}

func (h *Handler) serveGetGitWorktree(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	worktreeID := r.PathValue("worktreeId")
	wt, err := h.inventory.GetGitWorktreeByID(r.Context(), worktreeID)
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	repo, err := h.inventory.GetGitRepositoryByID(r.Context(), wt.RepositoryID)
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	branchName := ""
	if wt.BranchID != "" {
		br, err := h.inventory.GetGitBranchByID(r.Context(), wt.BranchID)
		if err != nil {
			WriteGitStoreError(w, r, op, err)
			return
		}
		branchName = br.Name
	}
	base := h.gitWorktreeJSON(wt)
	writeJSON(w, r, op, http.StatusOK, gitWorktreeDetailJSON{
		ID:                 base.ID,
		RepositoryID:       base.RepositoryID,
		Path:               base.Path,
		HostPath:           base.HostPath,
		Name:               base.Name,
		IsMain:             base.IsMain,
		BranchID:           base.BranchID,
		CreatedAt:          base.CreatedAt,
		RepositoryPath:     repo.Path,
		RepositoryHostPath: h.paths.DisplayHostPath(repo.Path),
		BranchName:         branchName,
	})
}

func (h *Handler) serveDeleteGitWorktree(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	worktreeID := r.PathValue("worktreeId")
	removeFromDisk, force := gitWorktreeDeleteQuery(r, op)
	var err error
	if removeFromDisk {
		err = h.write.RemoveGitWorktreeFromDiskByID(r.Context(), worktreeID, force)
	} else {
		err = h.inventory.UnregisterGitWorktreeByID(r.Context(), worktreeID)
	}
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) serveListGitBranches(w http.ResponseWriter, r *http.Request, op string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	rows, err := h.inventory.ListGitBranchesByRepo(r.Context(), repoID)
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitBranchJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, toGitBranchJSON(row))
	}
	writeJSON(w, r, op, http.StatusOK, gitBranchesListResponse{Branches: out})
}
