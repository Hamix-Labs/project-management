package handler

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
)

func repoErrUserMessage(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.repoErrUserMessage")
	return apijson.UserFacingMessage(err, apijson.SettingsInvalidInputMark, apijson.TasksInvalidInputMark)
}
