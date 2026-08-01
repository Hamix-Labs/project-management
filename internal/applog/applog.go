// Package applog installs process-wide slog for Hamix hosts (taskapi, hamix-desktop):
// JSONL file logging plus an optional stderr floor.
package applog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

// OpenJSONL creates the log directory if needed and opens a new JSON-lines log file
// named {cmdPrefix}-YYYY-MM-DD-HHMMSS-<nanos>.jsonl.
// dirFlag takes precedence over HAMIX_LOG_DIR; when both are empty, "logs" (relative
// to the process working directory) is used.
// fileLevel is the minimum level for the bootstrap Debug line written into the new file.
func OpenJSONL(cmdPrefix, dirFlag string, fileLevel slog.Level) (f *os.File, path string, err error) {
	prefix := strings.TrimSpace(cmdPrefix)
	if prefix == "" {
		prefix = "hamix"
	}
	abs, err := ResolveLogDir(dirFlag)
	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, "", fmt.Errorf("create log directory: %w", err)
	}
	now := time.Now()
	name := fmt.Sprintf("%s-%s-%09d.jsonl", prefix, now.Format("2006-01-02-150405"), now.Nanosecond())
	path = filepath.Join(abs, name)
	f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return nil, "", fmt.Errorf("open log file: %w", err)
	}
	early := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: fileLevel}))
	early.Debug("trace", "cmd", prefix, "operation", "applog.OpenJSONL", "path", path)
	return f, path, nil
}

// ResolveLogDir returns the absolute log directory path.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ResolveLogDir(dirFlag string) (string, error) {
	dir := strings.TrimSpace(dirFlag)
	if dir == "" {
		dir = strings.TrimSpace(os.Getenv("HAMIX_LOG_DIR"))
	}
	if dir == "" {
		dir = "logs"
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve log directory: %w", err)
	}
	return abs, nil
}

// Close syncs and closes a log file opened by OpenJSONL. Safe with nil.
func Close(cmd string, logFile *os.File) {
	slog.Debug("trace", "cmd", cmd, "operation", "applog.Close")
	if logFile == nil {
		return
	}
	if err := logFile.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: log file sync: %v\n", cmd, err)
	}
	if err := logFile.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: log file close: %v\n", cmd, err)
	}
}

// InstallConfig configures slog.SetDefault for a host process.
type InstallConfig struct {
	// File is the JSONL destination; nil when Minimized or when file logging is off.
	File *os.File
	// FileLevel is the minimum level written to File (ignored when File is nil).
	FileLevel slog.Level
	// StderrLevel, when non-nil, also writes text logs at that minimum to os.Stderr.
	// taskapi leaves this nil (file-only). hamix-desktop sets Warn.
	StderrLevel *slog.Level
	// Minimized means no JSONL file; only slog.Error to stderr (text).
	Minimized bool
	// ProcessLogSeq is the process-wide log_seq fallback counter (required).
	ProcessLogSeq *atomic.Uint64
}

// Install sets slog.Default to a request/sequence-wrapped handler fan-out.
func Install(cfg InstallConfig) {
	slog.Debug("trace", "operation", "applog.Install", "minimized", cfg.Minimized)
	var handlers []slog.Handler
	if cfg.Minimized {
		handlers = append(handlers, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	} else {
		if cfg.File != nil {
			handlers = append(handlers, slog.NewJSONHandler(cfg.File, &slog.HandlerOptions{Level: cfg.FileLevel}))
		}
		if cfg.StderrLevel != nil {
			handlers = append(handlers, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: *cfg.StderrLevel}))
		}
		if len(handlers) == 0 {
			// Defensive: never leave the process on Go's default Info→stderr flood.
			handlers = append(handlers, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
		}
	}
	base := slog.Handler(handlers[0])
	if len(handlers) > 1 {
		base = &multiHandler{handlers: handlers}
	}
	seq := cfg.ProcessLogSeq
	if seq == nil {
		seq = &atomic.Uint64{}
	}
	slog.SetDefault(slog.New(logctx.WrapSlogHandlerWithLogSequence(
		logctx.WrapSlogHandlerWithRequestContext(base),
		seq,
	)))
}

// EmitFileLoggingConfig writes the standard "logging config" line at fileLevel.
func EmitFileLoggingConfig(cmd string, fileLevel slog.Level) {
	slog.Log(context.Background(), fileLevel, "logging config",
		"cmd", cmd, "operation", cmd+".logging",
		"min_level", fileLevel.String(), "json_file", true)
}

// multiHandler fans a record out to every inner handler that Enabled allows.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var first error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: next}
}
