package handler

import (
	"errors"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitexec"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
)

type repoSearchResponse struct {
	Paths []string `json:"paths"`
}

type repoValidateRangeResponse struct {
	OK        bool   `json:"ok"`
	LineCount int    `json:"line_count,omitempty"`
	Warning   string `json:"warning,omitempty"`
}

type repoFileResponse struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Binary    bool   `json:"binary"`
	Truncated bool   `json:"truncated"`
	SizeBytes int64  `json:"size_bytes"`
	LineCount int    `json:"line_count"`
	Warning   string `json:"warning,omitempty"`
}

// MaxSearchQueryBytes caps substring search cost; exported for contract tests in pkgs/tasks/handler.
const MaxSearchQueryBytes = 512

// MaxRelPathQueryBytes caps repo-relative path query length.
const MaxRelPathQueryBytes = 4096

// MaxLineQueryParamBytes caps start/end line query params.
const MaxLineQueryParamBytes = 32

// MaxShaQueryBytes caps commit SHA query length.
const MaxShaQueryBytes = 64

var repoShaQueryPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

type repoDiffResponse struct {
	SHA          string `json:"sha"`
	Patch        string `json:"patch"`
	Truncated    bool   `json:"truncated"`
	SizeBytes    int    `json:"size_bytes"`
	Author       string `json:"author,omitempty"`
	AuthorEmail  string `json:"author_email,omitempty"`
	ParentSHA    string `json:"parent_sha,omitempty"`
	FilesChanged int    `json:"files_changed,omitempty"`
	Insertions   int    `json:"insertions,omitempty"`
	Deletions    int    `json:"deletions,omitempty"`
}

type repoUnavailableErrorBody struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

func (h *Handler) requireWorktreeRepo(w http.ResponseWriter, r *http.Request, op string) (*repo.Root, bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.handler.Handler.requireWorktreeRepo", "http_op", op)
	worktreeID := strings.TrimSpace(r.URL.Query().Get("worktree_id"))
	if worktreeID == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "worktree_id query parameter is required")
		return nil, false
	}
	if h.provider == nil {
		handlerhttp.WriteJSON(w, r, op, http.StatusNotFound, repoUnavailableErrorBody{
			Error:  "worktree not found",
			Reason: repo.RepoReasonWorktreeNotFound,
		})
		return nil, false
	}
	root, reason, err := h.provider.OpenWorktreeRoot(r.Context(), worktreeID)
	if err != nil {
		slog.Log(r.Context(), slog.LevelError, "repo provider failed",
			"cmd", calltrace.LogCmd, "operation", op, "reason", reason, "err", err)
		handlerhttp.WriteJSON(w, r, op, http.StatusInternalServerError, repoUnavailableErrorBody{
			Error: err.Error(), Reason: reason,
		})
		return nil, false
	}
	if root == nil {
		if reason == repo.RepoReasonWorktreeNotFound {
			handlerhttp.WriteJSON(w, r, op, http.StatusNotFound, repoUnavailableErrorBody{
				Error:  "worktree not found",
				Reason: reason,
			})
			return nil, false
		}
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "worktree_id query parameter is required")
		return nil, false
	}
	return root, true
}

func (h *Handler) repoSearch(w http.ResponseWriter, r *http.Request) {
	const op = "repo.search"
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op, "search_q", handlerhttp.TruncateRunes(r.URL.Query().Get("q"), handlerhttp.MaxHTTPLogTitleRunes))
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, methodNotAllowed(), http.StatusMethodNotAllowed)
		return
	}
	root, ok := h.requireWorktreeRepo(w, r, op)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	if len(q) > MaxSearchQueryBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "search query too long")
		return
	}
	t0 := time.Now()
	paths, err := root.Search(q)
	dur := time.Since(t0)
	if err != nil {
		slog.Log(r.Context(), slog.LevelError, "repo operation failed", "cmd", calltrace.LogCmd, "operation", op, "duration_ms", dur.Milliseconds(), "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "search failed")
		return
	}
	slog.Info("repo search completed", "cmd", calltrace.LogCmd, "operation", op, "path_count", len(paths), "duration_ms", dur.Milliseconds(), "q_empty", strings.TrimSpace(q) == "")
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, repoSearchResponse{Paths: paths})
}

func (h *Handler) repoValidateRange(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.handler.Handler.repoValidateRange")
	const op = "repo.validate_range"
	r = calltrace.WithRequestRoot(r, op)
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	startStr := strings.TrimSpace(r.URL.Query().Get("start"))
	endStr := strings.TrimSpace(r.URL.Query().Get("end"))
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, methodNotAllowed(), http.StatusMethodNotAllowed)
		return
	}
	root, ok := h.requireWorktreeRepo(w, r, op)
	if !ok {
		return
	}
	if path == "" || startStr == "" || endStr == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path, start, and end query parameters are required")
		return
	}
	if len(path) > MaxRelPathQueryBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path too long")
		return
	}
	if len(startStr) > MaxLineQueryParamBytes || len(endStr) > MaxLineQueryParamBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "start or end too long")
		return
	}
	handlerhttp.DebugHTTPRequest(r, op, "validate_path", handlerhttp.TruncateRunes(path, handlerhttp.MaxHTTPLogTitleRunes), "validate_start", handlerhttp.TruncateRunes(startStr, handlerhttp.MaxHTTPLogTitleRunes), "validate_end", handlerhttp.TruncateRunes(endStr, handlerhttp.MaxHTTPLogTitleRunes))
	start, err1 := strconv.Atoi(startStr)
	end, err2 := strconv.Atoi(endStr)
	if err1 != nil || err2 != nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "start and end must be integers")
		return
	}
	h.writeRepoValidateRangeOutcome(w, r, op, root, path, start, end)
}

func (h *Handler) writeRepoValidateRangeOutcome(w http.ResponseWriter, r *http.Request, op string, root *repo.Root, path string, start, end int) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.handler.Handler.writeRepoValidateRangeOutcome")
	abs, err := root.Resolve(path)
	if err != nil {
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, repoValidateRangeResponse{
			OK:      false,
			Warning: repoErrUserMessage(err),
		})
		return
	}
	n, err := repo.LineCount(abs)
	if err != nil {
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, repoValidateRangeResponse{
			OK:      false,
			Warning: repoErrUserMessage(err),
		})
		return
	}
	if err := repo.ValidateRange(abs, start, end); err != nil {
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, repoValidateRangeResponse{
			OK:        false,
			LineCount: n,
			Warning:   repoErrUserMessage(err),
		})
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, repoValidateRangeResponse{OK: true, LineCount: n})
}

func (h *Handler) repoFile(w http.ResponseWriter, r *http.Request) {
	const op = "repo.file"
	r = calltrace.WithRequestRoot(r, op)
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	handlerhttp.DebugHTTPRequest(r, op, "file_path", handlerhttp.TruncateRunes(path, handlerhttp.MaxHTTPLogTitleRunes))
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, methodNotAllowed(), http.StatusMethodNotAllowed)
		return
	}
	root, ok := h.requireWorktreeRepo(w, r, op)
	if !ok {
		return
	}
	if path == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path query parameter is required")
		return
	}
	if len(path) > MaxRelPathQueryBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "path too long")
		return
	}
	abs, err := root.Resolve(path)
	if err != nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, repoErrUserMessage(err))
		return
	}
	t0 := time.Now()
	fp, err := repo.ReadFilePreview(abs)
	dur := time.Since(t0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			handlerhttp.WriteJSONError(w, r, op, http.StatusNotFound, "file not found")
			return
		}
		slog.Log(r.Context(), slog.LevelError, "repo operation failed", "cmd", calltrace.LogCmd, "operation", op, "duration_ms", dur.Milliseconds(), "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "read failed")
		return
	}
	slog.Info("repo file preview", "cmd", calltrace.LogCmd, "operation", op, "duration_ms", dur.Milliseconds(), "binary", fp.Binary, "truncated", fp.Truncated, "size_bytes", fp.SizeBytes)
	resp := repoFileResponse{
		Path:      path,
		Content:   fp.Content,
		Binary:    fp.Binary,
		Truncated: fp.Truncated,
		SizeBytes: fp.SizeBytes,
		LineCount: fp.LineCount,
	}
	switch {
	case fp.Binary && fp.Truncated:
		resp.Warning = "This file looks binary or non-text, and the preview was truncated. You can insert the file reference without a line range."
	case fp.Binary:
		resp.Warning = "This file looks binary or non-text. You can insert the file reference without a line range."
	case fp.Truncated:
		resp.Warning = "Preview truncated to the first 32 MiB of the file. Line selection applies only to the visible text."
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}

func (h *Handler) repoDiff(w http.ResponseWriter, r *http.Request) {
	const op = "repo.diff"
	r = calltrace.WithRequestRoot(r, op)
	sha := strings.TrimSpace(r.URL.Query().Get("sha"))
	handlerhttp.DebugHTTPRequest(r, op, "diff_sha", handlerhttp.TruncateRunes(sha, handlerhttp.MaxHTTPLogTitleRunes))
	if r.Method != http.MethodGet {
		handlerhttp.WriteError(w, r, op, methodNotAllowed(), http.StatusMethodNotAllowed)
		return
	}
	root, ok := h.requireWorktreeRepo(w, r, op)
	if !ok {
		return
	}
	if sha == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "sha query parameter is required")
		return
	}
	if len(sha) > MaxShaQueryBytes {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "sha too long")
		return
	}
	if !repoShaQueryPattern.MatchString(sha) {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "invalid sha")
		return
	}
	t0 := time.Now()
	patch, truncated, err := gitexec.ShowCommitPatch(r.Context(), root.Abs(), sha, gitexec.DefaultMaxPatchBytes)
	meta, metaErr := gitexec.LoadCommitMeta(r.Context(), root.Abs(), sha)
	dur := time.Since(t0)
	if err != nil {
		if errors.Is(err, gitexec.ErrNotFound) {
			handlerhttp.WriteJSONError(w, r, op, http.StatusNotFound, "commit not found")
			return
		}
		slog.Log(r.Context(), slog.LevelError, "repo diff failed",
			"cmd", calltrace.LogCmd, "operation", op, "duration_ms", dur.Milliseconds(), "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "diff failed")
		return
	}
	if metaErr != nil && !errors.Is(metaErr, gitexec.ErrNotFound) {
		slog.Log(r.Context(), slog.LevelWarn, "repo diff meta failed",
			"cmd", calltrace.LogCmd, "operation", op, "duration_ms", dur.Milliseconds(), "err", metaErr)
	}
	slog.Info("repo diff completed", "cmd", calltrace.LogCmd, "operation", op,
		"duration_ms", dur.Milliseconds(), "truncated", truncated, "size_bytes", len(patch))
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, repoDiffResponse{
		SHA:          sha,
		Patch:        patch,
		Truncated:    truncated,
		SizeBytes:    len(patch),
		Author:       meta.Author,
		AuthorEmail:  meta.AuthorEmail,
		ParentSHA:    meta.ParentSHA,
		FilesChanged: meta.FilesChanged,
		Insertions:   meta.Insertions,
		Deletions:    meta.Deletions,
	})
}
