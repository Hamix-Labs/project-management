package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

type jsonCodedErrorBody struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitService() gitwork.Service {
	if h.git != nil {
		return h.git
	}
	return gitwork.New()
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func gitErrHTTP(err error) (status int, code, msg string) {
	status = http.StatusInternalServerError
	msg = "internal server error"
	if err == nil {
		return status, code, msg
	}
	if c := domain.GitErrCode(err); c != "" {
		code = c
		msg = err.Error()
		switch c {
		case domain.GitCodeRepositoryNotFound, domain.GitCodeWorktreeNotFound, domain.GitCodeBranchNotFound:
			status = http.StatusNotFound
		case domain.GitCodeNotARepository, domain.GitCodePathExists, domain.GitCodeBranchExists,
			domain.GitCodeBranchCheckedOut, domain.GitCodeHasRunningTask, domain.GitCodeDuplicate,
			domain.GitCodeBranchBoundToWorktree, domain.GitCodeProjectRepoMismatch,
			domain.GitCodeBootstrapMismatch:
			status = http.StatusConflict
		default:
			status = http.StatusBadRequest
		}
		return status, code, msg
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "", "not found"
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, "", invalidInputDetail(err)
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "", conflictDetail(err)
	default:
		return status, code, msg
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeGitStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	status, code, msg := gitErrHTTP(err)
	if code != "" {
		writeJSONCodedError(w, r, op, status, code, msg)
		return
	}
	if status >= 500 {
		writeJSONError(w, r, op, status, msg)
		return
	}
	writeJSONError(w, r, op, status, msg)
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeJSONCodedError(w http.ResponseWriter, r *http.Request, op string, status int, code, msg string) {
	setJSONHeaders(w)
	w.WriteHeader(status)
	body := jsonCodedErrorBody{Error: msg, Code: code}
	if r != nil {
		body.RequestID = logctx.RequestIDFromContext(r.Context())
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}

func gitWorktreeDeleteQuery(r *http.Request, op string) (removeFromDisk, force bool) {
	removeFromDisk = r.URL.Query().Get("remove_from_disk") == "true"
	force = r.URL.Query().Get("force") == "true"
	if force && !removeFromDisk {
		slog.Warn("deprecated query param ignored", "cmd", calltrace.LogCmd, "operation", op, "param", "force")
	}
	return removeFromDisk, force
}
