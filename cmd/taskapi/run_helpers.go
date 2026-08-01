package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/AlexsanderHamir/Hamix/internal/envload"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiruntime"
)

// run_helpers.go owns runTaskAPIService (the cmd.Run body): logging,
// env load, shared runtime start, HTTP serve, and shutdown. Lifecycle
// subsystems (logging, http) live in sibling run_*.go files.

func logHTTPTimeoutsAndShutdown() {
	slog.Info("http server limits", "cmd", cmdName, "operation", "taskapi.http_limits",
		"read_header_timeout_sec", int(readHeaderTimeout.Seconds()),
		"read_timeout_sec", int(readTimeout.Seconds()),
		"idle_timeout_sec", int(idleTimeout.Seconds()),
		"write_timeout_disabled", true,
		"max_header_bytes", maxRequestHeaders,
		"shutdown_timeout_sec", int(shutdownTimeout.Seconds()),
	)
}

func runTaskAPIService(port, host, envPath, logDir, logLevelFlag string, disableLogging bool, migrateFlag bool) int {
	minLevel, logFile, logPath, minimized, err := openTaskAPILogging(logDir, logLevelFlag, disableLogging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmdName, err)
		return 1
	}
	defer deferCloseTaskAPILogFile(logFile)

	var processLogSeq atomic.Uint64
	installTaskAPIDefaultLogger(logFile, minimized, minLevel, &processLogSeq, logPath)

	envLoadedPath, err := envload.Load(envPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: startup failed: %v\n", cmdName, err)
		slog.Error("startup failed", "cmd", cmdName, "operation", "taskapi.startup_env", "err", err)
		return 1
	}
	slog.Info("env loaded", "cmd", cmdName, "operation", "taskapi.startup", "path", envLoadedPath)

	migrateEnabled := taskapiconfig.MigrateEnabled(migrateFlag)
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	rt, err := taskapiruntime.Start(appCtx, taskapiruntime.Options{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Migrate:     migrateEnabled,
		CmdName:     cmdName,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: startup failed: %v\n", cmdName, err)
		slog.Error("startup failed", "cmd", cmdName, "operation", "taskapi.startup_runtime", "err", err)
		return 1
	}
	logHTTPTimeoutsAndShutdown()

	maybeRunGitReconcileOnStartup(appCtx, rt)
	shutdownViaSignal, serveErr := runTaskAPIHTTPServer(appCtx, port, host, rt)

	closeErr := rt.Close()
	if serveErr != nil {
		slog.Error("server error", "cmd", cmdName, "operation", "taskapi.serve", "err", serveErr)
		return 1
	}
	if closeErr != nil {
		return 1
	}
	slog.Info("process exit", "cmd", cmdName, "operation", "taskapi.shutdown", "phase", "exit",
		"db_closed", true, "signal_shutdown", shutdownViaSignal)
	return 0
}
