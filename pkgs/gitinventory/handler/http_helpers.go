package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

type jsonCodedErrorBody struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Gitinventory cannot import pkgs/tasks/handlerhttp (handlerhttp delegates git
// store errors back here). These thin helpers mirror handlerhttp JSON behavior.
// InvalidInputDetail is shared via pkgs/tasks/apijson (cycle-safe).

//funclogmeasure:skip category=hot-path reason="JSON helper mirroring handlerhttp; HTTP handlers emit operation traces."
func decodeJSON(ctx context.Context, r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("json trailing data: %w", err)
	}
	return fmt.Errorf("%w: json trailing data", taskcoredomain.ErrInvalidInput)
}

//funclogmeasure:skip category=hot-path reason="JSON helper mirroring handlerhttp; HTTP handlers emit operation traces."
func writeJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	apijson.ApplySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		writeJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	w.WriteHeader(code)
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n"))
}

//funclogmeasure:skip category=hot-path reason="JSON helper mirroring handlerhttp; HTTP handlers emit operation traces."
func writeJSONError(w http.ResponseWriter, r *http.Request, op string, code int, msg string) {
	apijson.WriteJSONError(w, r, op, code, msg, calltrace.Path)
}

//funclogmeasure:skip category=hot-path reason="JSON helper mirroring handlerhttp; HTTP handlers emit operation traces."
func writeError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	msg := http.StatusText(code)
	if code == http.StatusBadRequest {
		msg = err.Error()
	}
	writeJSONError(w, r, op, code, msg)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func conflictDetail(err error) string {
	s := err.Error()
	const mark = "tasks: conflict: "
	if i := strings.Index(s, mark); i >= 0 {
		return strings.TrimSpace(s[i+len(mark):])
	}
	return ""
}

// GitErrHTTP maps git and domain store errors to HTTP status, code, and message.
//
//funclogmeasure:skip category=hot-path reason="Pure error mapper; HTTP handlers emit operation traces."
func GitErrHTTP(err error) (status int, code, msg string) {
	status = http.StatusInternalServerError
	msg = "internal server error"
	if err == nil {
		return status, code, msg
	}
	if c := gitdomain.GitErrCode(err); c != "" {
		code = c
		msg = err.Error()
		switch c {
		case gitdomain.GitCodeRepositoryNotFound, gitdomain.GitCodeWorktreeNotFound, gitdomain.GitCodeBranchNotFound:
			status = http.StatusNotFound
		case gitdomain.GitCodeNotARepository, gitdomain.GitCodePathExists, gitdomain.GitCodeBranchExists,
			gitdomain.GitCodeBranchCheckedOut, gitdomain.GitCodeHasRunningTask, gitdomain.GitCodeDuplicate,
			gitdomain.GitCodeBranchBoundToWorktree, gitdomain.GitCodeProjectRepoMismatch,
			gitdomain.GitCodeBootstrapMismatch:
			status = http.StatusConflict
		default:
			status = http.StatusBadRequest
		}
		return status, code, msg
	}
	switch {
	case errors.Is(err, taskcoredomain.ErrNotFound):
		return http.StatusNotFound, "", "not found"
	case errors.Is(err, taskcoredomain.ErrInvalidInput):
		return http.StatusBadRequest, "", apijson.InvalidInputDetail(err, apijson.TasksInvalidInputMark)
	case errors.Is(err, taskcoredomain.ErrConflict):
		return http.StatusConflict, "", conflictDetail(err)
	default:
		return status, code, msg
	}
}

// WriteGitStoreError writes a git-aware JSON error response.
//
//funclogmeasure:skip category=hot-path reason="JSON helper mirroring handlerhttp; HTTP handlers emit operation traces."
func WriteGitStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	status, code, msg := GitErrHTTP(err)
	if code != "" {
		writeJSONCodedError(w, r, op, status, code, msg)
		return
	}
	writeJSONError(w, r, op, status, msg)
}

//funclogmeasure:skip category=hot-path reason="JSON helper mirroring handlerhttp; HTTP handlers emit operation traces."
func writeJSONCodedError(w http.ResponseWriter, r *http.Request, op string, status int, code, msg string) {
	apijson.ApplySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
