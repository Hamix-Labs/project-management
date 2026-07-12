package handler

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func invalidInputDetail(err error) string {
	s := err.Error()
	const mark = "tasks: invalid input: "
	if i := strings.Index(s, mark); i >= 0 {
		return strings.TrimSpace(s[i+len(mark):])
	}
	return ""
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func repoErrUserMessage(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.handler.repoErrUserMessage")
	if d := invalidInputDetail(err); d != "" {
		return d
	}
	return err.Error()
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func methodNotAllowed() error {
	return errors.New("method not allowed")
}
