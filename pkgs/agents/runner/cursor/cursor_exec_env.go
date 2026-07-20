package cursor

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/adapterkit"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

func buildEnv(reqEnv map[string]string, extraKeys []string) []string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.buildEnv",
		"req_env_count", len(reqEnv), "extra_keys", len(extraKeys))
	return adapterkit.BuildEnv(reqEnv, envPolicy(extraKeys))
}

func isDeniedEnvKey(k string) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.isDeniedEnvKey", "key", k)
	return adapterkit.IsDeniedEnvKey(k, envPolicy(nil))
}

func liveHomePaths() []string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.liveHomePaths")
	return adapterkit.LiveHomePaths()
}

// Redact replaces secret-shaped substrings in s with [REDACTED] markers
// and rewrites absolute home paths to "~". It is exported so callers
// (worker logs, future adapters) can apply the same floor.
func Redact(s string) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.Redact", "bytes", len(s))
	return redact(s, liveHomePaths())
}

func redact(s string, homePaths []string) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.redact",
		"bytes", len(s), "home_paths", len(homePaths))
	return adapterkit.Redact(s, redactionPolicy(homePaths))
}

func withOptionalTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.withOptionalTimeout",
		"timeout_ns", int64(d))
	if d <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}

func isCtxErr(ctx context.Context) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.isCtxErr")
	err := ctx.Err()
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

func defaultExecFn(ctx context.Context, dir string, env []string, stdin []byte, name string, args ...string) ([]byte, []byte, int, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.defaultExecFn",
		"dir", dir, "env_count", len(env), "stdin_bytes", len(stdin), "name", name, "argc", len(args))
	return adapterkit.DefaultExec(ctx, dir, env, stdin, name, args...)
}

func defaultStreamExecFn(ctx context.Context, dir string, env []string, stdin []byte, name string, onStdoutLine func([]byte), args ...string) ([]byte, []byte, int, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.defaultStreamExecFn",
		"dir", dir, "env_count", len(env), "stdin_bytes", len(stdin), "name", name, "argc", len(args))
	return adapterkit.DefaultStreamExec(ctx, dir, env, stdin, name, onStdoutLine, args...)
}

func normalizePipeReadError(err error) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.normalizePipeReadError",
		"err", err)
	if isClosedPipeReadError(err) {
		return nil
	}
	return adapterkit.NormalizePipeReadError(err)
}

func isClosedPipeReadError(err error) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.isClosedPipeReadError",
		"err", err)
	return adapterkit.IsClosedPipeReadError(err)
}
