package handler

import (
	"net/http"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitRepositoryJSON(r domain.GitRepository) gitRepositoryJSON {
	return h.gitRepositorySummaryJSON(r, "", 0)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitRepositorySummaryJSON(r domain.GitRepository, mainBranch string, linkedWorktrees int) gitRepositoryJSON {
	return gitRepositoryJSON{
		ID:                  r.ID,
		Path:                r.Path,
		GitCommonDir:        r.GitCommonDir,
		HostPath:            h.pathMap.DisplayHostPath(r.Path),
		DefaultBranch:       r.DefaultBranch,
		MainBranchName:      mainBranch,
		LinkedWorktreeCount: linkedWorktrees,
		CreatedAt:           r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) listGitRepositories(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.list"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveListGitRepositories(w, r, op, gitProjectScope(projectID))
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) createGitRepository(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.create"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveCreateGitRepository(w, r, op, gitProjectScope(projectID))
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) getGitRepository(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.get"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveGetGitRepository(w, r, op, gitProjectScope(projectID))
}

//funclogmeasure:skip category=delegate-already-logs reason="Project-scoped route wrapper; serve* emits operation trace."
func (h *Handler) deleteGitRepository(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.delete"
	projectID, err := parseGitProjectID(r)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	h.serveDeleteGitRepository(w, r, op, gitProjectScope(projectID))
}
