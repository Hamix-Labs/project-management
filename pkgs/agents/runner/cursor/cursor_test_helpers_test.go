package cursor_test

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// captured records every (env, stdin, dir, name, args) tuple a fake ExecFn
// receives, so tests can assert on what the adapter actually invoked.
type captured struct {
	dir   string
	env   []string
	stdin []byte
	name  string
	args  []string
}

// fakeExec returns an ExecFn that records its inputs into *captured and
// returns the configured outputs. cancelOnInvoke=true delays return until
// ctx is cancelled (simulating a long-running child).
func fakeExec(c *captured, stdout, stderr []byte, exitCode int, runErr error, cancelOnInvoke bool) cursor.ExecFn {
	return func(ctx context.Context, dir string, env []string, stdin []byte, name string, args ...string) ([]byte, []byte, int, error) {
		c.dir = dir
		c.env = append([]string(nil), env...)
		c.stdin = append([]byte(nil), stdin...)
		c.name = name
		c.args = append([]string(nil), args...)
		if cancelOnInvoke {
			<-ctx.Done()
			return stdout, stderr, 0, ctx.Err()
		}
		return stdout, stderr, exitCode, runErr
	}
}

func fakeStreamExec(c *captured, stdout, stderr []byte, exitCode int, runErr error) cursor.StreamExecFn {
	return func(ctx context.Context, dir string, env []string, stdin []byte, name string, onStdoutLine func([]byte), args ...string) ([]byte, []byte, int, error) {
		c.dir = dir
		c.env = append([]string(nil), env...)
		c.stdin = append([]byte(nil), stdin...)
		c.name = name
		c.args = append([]string(nil), args...)
		for _, line := range bytes.Split(bytes.TrimSpace(stdout), []byte("\n")) {
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			onStdoutLine(line)
		}
		return stdout, stderr, exitCode, runErr
	}
}

func newAdapter(execFn cursor.ExecFn, extraOpts ...func(*cursor.Options)) *cursor.Adapter {
	opts := cursor.Options{
		BinaryPath:           "fake-cursor-agent",
		ExecFn:               execFn,
		Name:                 "cursor-cli",
		Version:              "test-1.0",
		HomePathReplacements: []string{"/home/runner", `C:\Users\runner`},
	}
	for _, f := range extraOpts {
		f(&opts)
	}
	return cursor.New(opts)
}

func defaultRequest() runner.Request {
	return runner.Request{
		TaskID:     "11111111-1111-4111-8111-111111111111",
		AttemptSeq: 1,
		Phase:      cyclesdomain.PhaseExecute,
		Prompt:     "do the thing",
		WorkingDir: "/repo/work",
		Timeout:    2 * time.Second,
	}
}
func envSliceToMap(env []string) map[string]string {
	out := make(map[string]string, len(env))
	for _, kv := range env {
		i := strings.IndexByte(kv, '=')
		if i < 0 {
			continue
		}
		out[kv[:i]] = kv[i+1:]
	}
	return out
}

func equalStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestAdapter_EffectiveModel pins the per-adapter resolution rule the
// worker depends on for cycle_meta.cursor_model_effective and the new
// Prometheus model label: trim req.CursorModel and use it; otherwise
// fall back to Options.DefaultCursorModel; otherwise return "" so the
// audit row records the truth ("no model configured anywhere") rather
// than a substituted placeholder.
