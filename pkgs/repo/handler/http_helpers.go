package handler

import (
	"errors"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func repoErrUserMessage(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.handler.repoErrUserMessage")
	if d := apijson.InvalidInputDetail(err, apijson.TasksInvalidInputMark); d != "" {
		return d
	}
	return err.Error()
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func methodNotAllowed() error {
	return errors.New("method not allowed")
}
