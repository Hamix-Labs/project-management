package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const maxPathIDBytes = 128

func writeStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, taskcoredomain.ErrInvalidInput):
		code = http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict):
		code = http.StatusConflict
	}
	msg := "internal server error"
	if code != http.StatusInternalServerError {
		msg = err.Error()
	}
	handlerhttp.WriteJSONError(w, r, op, code, msg)
	if r != nil {
		ctx := r.Context()
		slog.Log(ctx, slog.LevelWarn, "request failed",
			"cmd", calltrace.LogCmd, "operation", op,
			"request_id", logctx.RequestIDFromContext(ctx),
			"http_status", code, "err", err)
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parsePathID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: id too long", taskcoredomain.ErrInvalidInput)
	}
	return id, nil
}
