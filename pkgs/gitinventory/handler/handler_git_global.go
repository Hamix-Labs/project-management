package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) listGlobalGitRepositories(w http.ResponseWriter, r *http.Request) {
	h.serveListGitRepositories(w, r, "git.repositories.list_global")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) createGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	h.serveCreateGitRepository(w, r, "git.repositories.create_global")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) getGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	h.serveGetGitRepository(w, r, "git.repositories.get_global")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) deleteGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	h.serveDeleteGitRepository(w, r, "git.repositories.delete_global")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) listGlobalGitWorktrees(w http.ResponseWriter, r *http.Request) {
	h.serveListGitWorktrees(w, r, "git.worktrees.list_global")
}

func (h *Handler) listGlobalGitWorktreesCheckoutStatus(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.checkout_status"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.listGlobalGitWorktreesCheckoutStatus")
	r = calltrace.WithRequestRoot(r, op)
	repo, err := h.inventory.GetGitRepositoryByID(r.Context(), r.PathValue("repoId"))
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	rows, err := h.write.RepoWorktreeCheckoutStatus(r.Context(), repo, h.gitService())
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitWorktreeCheckoutStatusJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, gitWorktreeCheckoutStatusFromRow(row))
	}
	writeJSON(w, r, op, http.StatusOK, gitWorktreeCheckoutStatusListResponse{Worktrees: out})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by listGlobalGitWorktreesCheckoutStatus."
func gitWorktreeCheckoutStatusFromRow(row contract.WorktreeCheckoutStatusRow) gitWorktreeCheckoutStatusJSON {
	j := gitWorktreeCheckoutStatusJSON{
		WorktreeID: row.WorktreeID,
		Available:  row.Available,
	}
	if !row.Available {
		j.Reason = row.Reason
		return j
	}
	st := row.Status
	j.Dirty = st.Dirty
	j.Detached = st.Detached
	if !st.HeadCommitAt.IsZero() {
		j.HeadCommitAt = st.HeadCommitAt.UTC().Format(time.RFC3339)
	}
	j.HasUpstream = st.HasUpstream
	if st.HasUpstream {
		j.Upstream = st.Upstream
		ahead := st.Ahead
		behind := st.Behind
		j.Ahead = &ahead
		j.Behind = &behind
	}
	return j
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) deleteGlobalGitWorktree(w http.ResponseWriter, r *http.Request) {
	h.serveDeleteGitWorktree(w, r, "git.worktrees.delete_global")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) listGlobalGitBranches(w http.ResponseWriter, r *http.Request) {
	h.serveListGitBranches(w, r, "git.branches.list_global")
}

func (h *Handler) listRepoProjects(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.projects.list"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.listRepoProjects")
	r = calltrace.WithRequestRoot(r, op)
	rows, err := h.projects.ListProjectsByRepository(r.Context(), r.PathValue("repoId"))
	if err != nil {
		WriteGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, projectsListResponse{Projects: rows, Limit: len(rows)})
}
