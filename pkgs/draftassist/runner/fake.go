package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// FakeOptions tunes the fake runner for tests.
type FakeOptions struct {
	// ThinkDelay before first status (default: immediate).
	ThinkDelay time.Duration
	// TokenDelay before first token after thinking (default: 5ms).
	TokenDelay time.Duration
	// SlowAfterToken keeps the run alive without tokens so heartbeats fire.
	SlowAfterToken time.Duration
	// TokenText is the assistant delta (default: a short phrase).
	TokenText string
}

// Fake is an in-process runner that emits status → token → done without a model.
type Fake struct {
	opts FakeOptions
}

// NewFake returns a fake runner with sensible defaults for latency tests.
//
//funclogmeasure:skip category=hot-path reason="Constructor; Run emits the operation trace."
func NewFake(opts FakeOptions) *Fake {
	if opts.TokenDelay == 0 {
		opts.TokenDelay = 5 * time.Millisecond
	}
	if opts.TokenText == "" {
		opts.TokenText = "Here is a tightened prompt draft."
	}
	return &Fake{opts: opts}
}

func (f *Fake) Name() string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.Fake.Name")
	return "fake"
}

func (f *Fake) Run(ctx context.Context, sessionID, runID string, in contract.RunInput, h contract.RunHandle) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.Fake.Run", "session_id", sessionID, "run_id", runID)
	emit := func(kind domain.EventKind, data any) error {
		return h.Emit(ctx, sessionID, runID, kind, data)
	}

	if f.opts.ThinkDelay > 0 {
		select {
		case <-ctx.Done():
			_ = emit(domain.EventDone, domain.DoneEventData{Status: domain.RunStatusCancelled})
			return ctx.Err()
		case <-time.After(f.opts.ThinkDelay):
		}
	}
	if err := emit(domain.EventStatus, domain.StatusEventData{Status: domain.RunStatusThinking}); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		_ = emit(domain.EventDone, domain.DoneEventData{Status: domain.RunStatusCancelled})
		return ctx.Err()
	case <-time.After(f.opts.TokenDelay):
	}
	if err := emit(domain.EventStatus, domain.StatusEventData{Status: domain.RunStatusStreaming}); err != nil {
		return err
	}
	if err := emit(domain.EventToken, domain.TokenEventData{Delta: f.opts.TokenText}); err != nil {
		return err
	}

	if f.opts.SlowAfterToken > 0 {
		select {
		case <-ctx.Done():
			_ = emit(domain.EventDone, domain.DoneEventData{Status: domain.RunStatusCancelled})
			return ctx.Err()
		case <-time.After(f.opts.SlowAfterToken):
		}
	}

	if err := emit(domain.EventDone, domain.DoneEventData{Status: domain.RunStatusDone}); err != nil {
		return err
	}
	_ = in
	return nil
}

var _ contract.Runner = (*Fake)(nil)
