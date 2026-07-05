package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func toGitBranchJSON(b domain.GitBranch) gitBranchJSON {
	return gitBranchJSON{
		ID:           b.ID,
		RepositoryID: b.RepositoryID,
		Name:         b.Name,
		HeadSHA:      b.HeadSHA,
		CreatedAt:    b.CreatedAt.UTC().Format(time.RFC3339),
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) listGitBranches(w http.ResponseWriter, r *http.Request) {
	const op = "git.branches.list"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveListGitBranches(w, r, op, gitProjectScope(projectID))
}

func (h *Handler) createGitBranch(w http.ResponseWriter, r *http.Request) {
	const op = "git.branches.create"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.createGitBranch")
	r = calltrace.WithRequestRoot(r, op)
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	var body gitBranchCreateJSON
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	branch, err := h.store.CreateGitBranch(r.Context(), projectID, r.PathValue("repoId"), store.CreateGitBranchInput{
		Name:       body.Name,
		StartPoint: body.StartPoint,
	}, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, toGitBranchJSON(branch))
}

func (h *Handler) deleteGitBranch(w http.ResponseWriter, r *http.Request) {
	const op = "git.branches.delete"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.deleteGitBranch")
	r = calltrace.WithRequestRoot(r, op)
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	force := r.URL.Query().Get("force") == "true" || r.URL.Query().Get("force") == "1"
	if err := h.store.DeleteGitBranch(r.Context(), projectID, r.PathValue("branchId"), force, h.gitService()); err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
