package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

//funclogmeasure:skip category=hot-path reason="Pure request parser without I/O; operation trace is emitted by reconcile handlers."
func parseGitReconcileRequest(r *http.Request) (contract.ReconcileGitInput, error) {
	var body gitReconcileRequest
	if r.Body != nil {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil && err != io.EOF {
			return contract.ReconcileGitInput{}, err
		}
	}
	return contract.ReconcileGitInput{
		BootstrapPath:         strings.TrimSpace(body.BootstrapPath),
		RepairGit:             body.Repair,
		DryRun:                body.DryRun,
		AllowRemove:           true,
		AllowCheckoutDiscover: true,
		AllowDiscover:         false,
	}, nil
}

//funclogmeasure:skip category=delegate-already-logs reason="Shared reconcile handler; public routes emit trace before calling this helper."
func (h *Handler) runGitReconcile(w http.ResponseWriter, r *http.Request, op, projectID, repoID string) {
	input, err := parseGitReconcileRequest(r)
	if err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	out, err := h.write.ReconcileGitRepository(r.Context(), projectID, repoID, input, h.gitService())
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusAccepted, toGitReconcileResponse(out))
}

func (h *Handler) reconcileGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.reconcile_global"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.reconcileGlobalGitRepository")
	r = calltrace.WithRequestRoot(r, op)
	h.runGitReconcile(w, r, op, "", r.PathValue("repoId"))
}

func (h *Handler) syncGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.sync_global"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.syncGlobalGitRepository")
	r = calltrace.WithRequestRoot(r, op)
	out, err := h.write.SyncGitRepository(r.Context(), r.PathValue("repoId"), h.gitService())
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusAccepted, toGitReconcileResponse(out))
}

func (h *Handler) relocateGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.relocate_global"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.relocateGlobalGitRepository")
	r = calltrace.WithRequestRoot(r, op)
	repoID := r.PathValue("repoId")
	var body gitRelocateRepositoryRequest
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	out, err := h.write.RelocateGitRepository(r.Context(), "", repoID, body.Path, h.gitService())
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusAccepted, toGitReconcileResponse(out))
}

func (h *Handler) relocateGlobalGitWorktree(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.relocate_global"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.relocateGlobalGitWorktree")
	r = calltrace.WithRequestRoot(r, op)
	var body gitRelocateWorktreeRequest
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	wt, err := h.write.RelocateGitWorktree(r.Context(), r.PathValue("worktreeId"), body.Path, h.gitService())
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, h.gitWorktreeJSON(wt))
}
