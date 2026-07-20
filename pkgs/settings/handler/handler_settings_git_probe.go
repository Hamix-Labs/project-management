package handler

import (
	"errors"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func (h *Handler) gitRepositoryProbe(w http.ResponseWriter, r *http.Request) {
	const op = "settings.git_repository_probe"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.gitRepositoryProbe")
	r = calltrace.WithRequestRoot(r, op)
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	handlerhttp.DebugHTTPRequest(r, op, "probe_path", handlerhttp.TruncateRunes(path, maxHTTPLogTitleRunes))
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}
	if path == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path required")
		return
	}
	if len(path) > maxRepoRelPathQueryBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path too long")
		return
	}

	gitSvc := h.gitService()
	opened, err := gitSvc.OpenRepository(r.Context(), path)
	if err != nil {
		if errors.Is(err, gitwork.ErrNotARepository) {
			handlerhttp.WriteJSON(w, r, op, http.StatusOK, gitRepositoryProbeResponse{
				Path:            path,
				IsGitRepository: false,
				Branches:        []gitLiveBranchJSON{},
			})
			return
		}
		if errors.Is(err, taskcoredomain.ErrInvalidInput) {
			handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, repoErrUserMessage(err))
			return
		}
		slog.Log(r.Context(), slog.LevelError, "git repository probe failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "git probe failed")
		return
	}

	live, err := gitSvc.ListBranches(r.Context(), opened)
	if err != nil {
		slog.Log(r.Context(), slog.LevelError, "list branches for probe failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "git probe failed")
		return
	}

	out := make([]gitLiveBranchJSON, 0, len(live))
	current := ""
	for _, b := range live {
		out = append(out, gitLiveBranchJSON{Name: b.Name, HeadSHA: b.HeadSHA})
		if b.IsCurrent {
			current = b.Name
		}
	}

	mainRoot, _, err := gitSvc.ResolveRegistration(r.Context(), opened.Root)
	if err != nil {
		slog.Log(r.Context(), slog.LevelError, "resolve registration for probe failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "git probe failed")
		return
	}

	handlerhttp.WriteJSON(w, r, op, http.StatusOK, gitRepositoryProbeResponse{
		Path:            opened.Root,
		MainPath:        mainRoot,
		IsMain:          gitwork.PathKeyEqual(opened.Root, mainRoot),
		IsGitRepository: true,
		CurrentBranch:   current,
		Branches:        out,
	})
}
