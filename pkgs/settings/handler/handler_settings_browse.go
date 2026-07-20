package handler

import (
	"errors"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func (h *Handler) workspaceRoots(w http.ResponseWriter, r *http.Request) {
	const op = "settings.workspace_roots"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.workspaceRoots")
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op)
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "working directory unavailable")
		return
	}
	gitRepos, err := h.gitInventory.ListAllGitRepositories(r.Context())
	if err != nil {
		slog.Log(r.Context(), slog.LevelError, "list git repositories failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "workspace roots unavailable")
		return
	}
	roots, env, err := repo.ResolveWorkspacePickerRoots(wd, gitRepos, workspaceRootsScope(r))
	if err != nil {
		slog.Log(r.Context(), slog.LevelError, "workspace roots failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "browse roots unavailable")
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, workspaceRootsResponse{Roots: roots, Environment: env})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func workspaceRootsScope(r *http.Request) string {
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	if scope == repo.WorkspacePickerScopeExpanded {
		return repo.WorkspacePickerScopeExpanded
	}
	return repo.WorkspacePickerScopeDefault
}

func (h *Handler) browseDirs(w http.ResponseWriter, r *http.Request) {
	const op = "settings.browse_dirs"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.browseDirs")
	r = calltrace.WithRequestRoot(r, op)
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	handlerhttp.DebugHTTPRequest(r, op, "browse_path", handlerhttp.TruncateRunes(path, maxHTTPLogTitleRunes))
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}
	if len(path) > maxRepoRelPathQueryBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path too long")
		return
	}

	var listing repo.BrowseDirListing
	var listErr error
	if repo.CustomBrowseRootsConfigured() {
		wd, err := os.Getwd()
		if err != nil {
			handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "working directory unavailable")
			return
		}
		roots, _, err := repo.ResolveBrowseRoots(wd)
		if err != nil {
			handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "browse roots unavailable")
			return
		}
		listing, listErr = repo.ListBrowseDirs(roots, path)
	} else {
		listing, listErr = repo.ListBrowseDirsUnrestricted(path)
	}
	if listErr != nil {
		if errors.Is(listErr, taskcoredomain.ErrInvalidInput) {
			handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, repoErrUserMessage(listErr))
			return
		}
		slog.Log(r.Context(), slog.LevelError, "browse dirs failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", listErr)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "browse failed")
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, browseDirsResponse{
		Path:       listing.Path,
		ParentPath: listing.ParentPath,
		IsGitRepo:  listing.IsGitRepo,
		Entries:    listing.Entries,
	})
}
