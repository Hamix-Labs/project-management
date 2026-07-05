package handler

import (
	"net/http"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitWorktreeJSON(w domain.GitWorktree) gitWorktreeJSON {
	return gitWorktreeJSON{
		ID:           w.ID,
		RepositoryID: w.RepositoryID,
		Path:         w.Path,
		HostPath:     h.pathMap.DisplayHostPath(w.Path),
		Name:         w.Name,
		IsMain:       w.IsMain,
		BranchID:     w.BranchID,
		CreatedAt:    w.CreatedAt.UTC().Format(time.RFC3339),
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) listGitWorktrees(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.list"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveListGitWorktrees(w, r, op, gitProjectScope(projectID))
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) createGitWorktree(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.create"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveCreateGitWorktree(w, r, op, gitProjectScope(projectID))
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) deleteGitWorktree(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.delete"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveDeleteGitWorktree(w, r, op, gitProjectScope(projectID))
}
