package claudecode

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry"
)

const (
	RunnerID          = "claude-code"
	RunnerLabel       = "Claude Code CLI"
	DefaultBinaryHint = "claude"
)

// Register adds the scaffold Claude Code adapter to the global registry.
// It is not called from registry/all — import
// pkgs/agents/runner/registry/scaffold (or call Register explicitly) to opt in.
func Register() {
	slog.Debug("trace", "cmd", "claudecode", "operation", "agents.runner.claudecode.register")
	registry.Register(
		registry.Descriptor{
			ID:                RunnerID,
			Label:             RunnerLabel,
			DefaultBinaryHint: DefaultBinaryHint,
		},
		func(opts registry.BuildOptions) (runner.Runner, error) {
			return New(Options{
				BinaryPath:   opts.BinaryPath,
				Version:      opts.Version,
				DefaultModel: opts.CursorModel,
			}), nil
		},
	)
}
