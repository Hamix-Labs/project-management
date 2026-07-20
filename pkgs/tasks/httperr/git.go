package httperr

import (
	"errors"
	"net/http"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

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
		return http.StatusConflict, "", apijson.ConflictDetail(err, apijson.TasksConflictMark)
	default:
		return status, code, msg
	}
}

// WriteGitStoreError writes a git-aware JSON error response.
//
//funclogmeasure:skip category=hot-path reason="JSON helper; HTTP handlers emit operation traces."
func WriteGitStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	status, code, msg := GitErrHTTP(err)
	if code != "" {
		apijson.WriteJSONCodedError(w, r, op, status, code, msg)
		return
	}
	apijson.WriteJSONError(w, r, op, status, msg, calltrace.Path)
}
