package draftsidecar

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	draftassistmetrics "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/metrics"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// Host is a started sidecar supervisor plus the SDK runner that dials it.
// Close tears the child down. Ready implements draftassist ReadyProbe.
type Host struct {
	Runner *Runner
	sup    *Supervisor
}

// Ready reports supervisor probe state. runner is always "sdk".
//
//funclogmeasure:skip category=hot-path reason="Ready-probe accessor; /draft-assist/ready already traces."
func (h *Host) Ready() (bool, string, string) {
	if h == nil || h.sup == nil {
		return false, RunnerName, draftassistmetrics.ReasonSidecarDown
	}
	return h.sup.Ready()
}

// Close stops the supervisor. Safe on a nil Host.
//
//funclogmeasure:skip category=tool-required-noop reason="Shutdown hook; child exit is logged in the supervise loop."
func (h *Host) Close() error {
	if h == nil || h.sup == nil {
		return nil
	}
	return h.sup.Close()
}

// MustHost locates hamix-draft-agent, requires CURSOR_API_KEY, spawns the
// sidecar, and blocks until GET /readyz reports ready. It never falls
// back to a fake runner; callers must fail boot on error.
//
//funclogmeasure:skip category=tool-required-noop reason="Boot-time sidecar host; Start/Close emit lifecycle logs."
func MustHost(ctx context.Context) (*Host, error) {
	return mustHost(ctx, Options{Stderr: os.Stderr})
}

func mustHost(ctx context.Context, opts Options) (*Host, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftsidecar.MustHost")
	if strings.TrimSpace(opts.APIKey) == "" && strings.TrimSpace(os.Getenv(APIKeyEnv)) == "" {
		return nil, fmt.Errorf("draftsidecar: %s is not set", APIKeyEnv)
	}
	if strings.TrimSpace(opts.BinaryPath) == "" {
		bin, err := ResolveBinary()
		if err != nil {
			return nil, err
		}
		opts.BinaryPath = bin
	}
	sup := NewSupervisor(opts)
	if err := sup.Start(ctx); err != nil {
		return nil, err
	}
	if err := waitReady(ctx, sup); err != nil {
		_ = sup.Close()
		return nil, err
	}
	slog.Info("draft-assist runner=sdk",
		"cmd", calltrace.LogCmd,
		"operation", "draftsidecar.MustHost",
		"binary", opts.BinaryPath,
		"port", sup.Port(),
	)
	return &Host{
		Runner: NewRunner(RunnerOptions{PortSource: sup}),
		sup:    sup,
	}, nil
}

func waitReady(ctx context.Context, s *Supervisor) error {
	timeout := s.opts.StartupTimeout
	if timeout <= 0 {
		timeout = defaultStartupTimeout
	}
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		ready, _, reason := s.Ready()
		if ready {
			return nil
		}
		if time.Now().After(deadline) {
			if reason == "" {
				reason = draftassistmetrics.ReasonSidecarDown
			}
			return fmt.Errorf("draftsidecar: sidecar not ready (%s)", reason)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("draftsidecar: wait for ready: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
