package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/AlexsanderHamir/Hamix/internal/applog"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
)

// run_logging.go owns slog/log-file setup for taskapi via internal/applog.
// Split off run_helpers.go per backend/go/layout.mdc
// (do not grow a single file into a junk drawer of unrelated concerns).

func emitTaskAPIFileLoggingConfig(minLevel slog.Level) {
	slog.Debug("trace", "cmd", cmdName, "operation", "taskapi.emitTaskAPIFileLoggingConfig")
	applog.EmitFileLoggingConfig(cmdName, minLevel)
}

func openTaskAPILogging(logDir, logLevelFlag string, disableLogging bool) (minLevel slog.Level, logFile *os.File, logPath string, minimized bool, err error) {
	slog.Debug("trace", "cmd", cmdName, "operation", "taskapi.openTaskAPILogging")
	minLevel, err = taskapiconfig.ResolveLogLevel(logLevelFlag)
	if err != nil {
		return minLevel, nil, "", false, err
	}
	minimized = taskapiconfig.LoggingMinimized(disableLogging)
	if minimized {
		return minLevel, nil, "", minimized, nil
	}
	var openErr error
	logFile, logPath, openErr = applog.OpenJSONL(cmdName, logDir, minLevel)
	if openErr != nil {
		return minLevel, nil, "", false, openErr
	}
	return minLevel, logFile, logPath, minimized, nil
}

func deferCloseTaskAPILogFile(logFile *os.File) {
	slog.Debug("trace", "cmd", cmdName, "operation", "taskapi.deferCloseTaskAPILogFile")
	applog.Close(cmdName, logFile)
}

func installTaskAPIDefaultLogger(logFile *os.File, minimized bool, minLevel slog.Level, processLogSeq *atomic.Uint64, logPath string) {
	if minimized {
		fmt.Fprintf(os.Stderr, "%s: logging minimized (no log file; errors only to stderr); set by -disable-logging or %s\n", cmdName, taskapiconfig.EnvDisableLogging)
	} else {
		fmt.Fprintf(os.Stderr, "%s: writing structured logs to %s (min level %s)\n", cmdName, logPath, minLevel.String())
	}
	applog.Install(applog.InstallConfig{
		File:          logFile,
		FileLevel:     minLevel,
		Minimized:     minimized,
		ProcessLogSeq: processLogSeq,
	})
	slog.Debug("trace", "cmd", cmdName, "operation", "taskapi.run")
	if !minimized {
		emitTaskAPIFileLoggingConfig(minLevel)
	}
}
