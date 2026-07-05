package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) listGlobalGitRepositories(w http.ResponseWriter, r *http.Request) {
	h.serveListGitRepositories(w, r, "git.repositories.list_global", "")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) createGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	h.serveCreateGitRepository(w, r, "git.repositories.create_global", "")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) getGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	h.serveGetGitRepository(w, r, "git.repositories.get_global", "")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) deleteGlobalGitRepository(w http.ResponseWriter, r *http.Request) {
	h.serveDeleteGitRepository(w, r, "git.repositories.delete_global", "")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) listGlobalGitWorktrees(w http.ResponseWriter, r *http.Request) {
	h.serveListGitWorktrees(w, r, "git.worktrees.list_global", "")
}

func (h *Handler) listGlobalGitWorktreesLive(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.list_live"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.listGlobalGitWorktreesLive")
	r = calltrace.WithRequestRoot(r, op)
	repo, err := h.store.GetGitRepositoryByID(r.Context(), r.PathValue("repoId"))
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	inventory, err := h.store.RepoWorktreeInventory(r.Context(), repo, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitLiveWorktreeJSON, 0, len(inventory))
	for _, wt := range inventory {
		out = append(out, gitLiveWorktreeJSON{
			Path:       wt.Path,
			Branch:     wt.Branch,
			IsMain:     wt.IsMain,
			Detached:   wt.Detached,
			Registered: wt.Registered,
			Locked:     wt.Locked,
			Prunable:   wt.Prunable,
		})
	}
	writeJSON(w, r, op, http.StatusOK, gitLiveWorktreesListResponse{Worktrees: out})
}

func (h *Handler) listGlobalGitWorktreesCheckoutStatus(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.checkout_status"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.listGlobalGitWorktreesCheckoutStatus")
	r = calltrace.WithRequestRoot(r, op)
	repo, err := h.store.GetGitRepositoryByID(r.Context(), r.PathValue("repoId"))
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	rows, err := h.store.RepoWorktreeCheckoutStatus(r.Context(), repo, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitWorktreeCheckoutStatusJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, gitWorktreeCheckoutStatusFromRow(row))
	}
	writeJSON(w, r, op, http.StatusOK, gitWorktreeCheckoutStatusListResponse{Worktrees: out})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by listGlobalGitWorktreesCheckoutStatus."
func gitWorktreeCheckoutStatusFromRow(row store.WorktreeCheckoutStatusRow) gitWorktreeCheckoutStatusJSON {
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

func (h *Handler) probeGlobalGitWorktree(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.probe"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.probeGlobalGitWorktree")
	r = calltrace.WithRequestRoot(r, op)
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeJSONError(w, r, op, http.StatusBadRequest, "path required")
		return
	}
	result, err := h.store.ProbeGitWorktree(r.Context(), r.PathValue("repoId"), path, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, gitWorktreeProbeResponse{
		Path:       result.Path,
		Linked:     result.Linked,
		IsMain:     result.IsMain,
		Branch:     result.Branch,
		Registered: result.Registered,
	})
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) createGlobalGitWorktree(w http.ResponseWriter, r *http.Request) {
	h.serveCreateGitWorktree(w, r, "git.worktrees.create_global", "")
}

func (h *Handler) registerGlobalGitWorktree(w http.ResponseWriter, r *http.Request) {
	const op = "git.worktrees.register_global"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.registerGlobalGitWorktree")
	r = calltrace.WithRequestRoot(r, op)
	var body gitWorktreeRegisterJSON
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	var bind store.BindBranchInput
	if body.Branch != nil {
		bind = store.BindBranchInput{
			Name:         body.Branch.Name,
			CreateBranch: body.Branch.CreateBranch,
			StartPoint:   body.Branch.StartPoint,
		}
	}
	wt, err := h.store.RegisterExistingGitWorktree(r.Context(), r.PathValue("repoId"), body.Path, body.Name, bind, h.gitService())
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, h.gitWorktreeJSON(wt))
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) deleteGlobalGitWorktree(w http.ResponseWriter, r *http.Request) {
	h.serveDeleteGitWorktree(w, r, "git.worktrees.delete_global", "")
}

//funclogmeasure:skip category=delegate-already-logs reason="Global route wrapper; serve* emits operation trace."
func (h *Handler) listGlobalGitBranches(w http.ResponseWriter, r *http.Request) {
	h.serveListGitBranches(w, r, "git.branches.list_global", "")
}

func (h *Handler) listGlobalGitBranchesLive(w http.ResponseWriter, r *http.Request) {
	const op = "git.branches.list_live"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.listGlobalGitBranchesLive")
	r = calltrace.WithRequestRoot(r, op)
	repo, err := h.store.GetGitRepositoryByID(r.Context(), r.PathValue("repoId"))
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	gitSvc := h.gitService()
	opened, err := gitSvc.OpenRepository(r.Context(), repo.Path)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	live, err := gitSvc.ListBranches(r.Context(), opened)
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	out := make([]gitLiveBranchJSON, 0, len(live))
	for _, b := range live {
		out = append(out, gitLiveBranchJSON{Name: b.Name, HeadSHA: b.HeadSHA})
	}
	writeJSON(w, r, op, http.StatusOK, gitLiveBranchesListResponse{Branches: out})
}

func (h *Handler) listRepoProjects(w http.ResponseWriter, r *http.Request) {
	const op = "git.repositories.projects.list"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.listRepoProjects")
	r = calltrace.WithRequestRoot(r, op)
	rows, err := h.store.ListProjectsByRepository(r.Context(), r.PathValue("repoId"))
	if err != nil {
		writeGitStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, projectsListResponse{Projects: rows, Limit: len(rows)})
}
