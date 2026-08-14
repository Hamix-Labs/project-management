package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
)

// RunHandle is the coordination surface the runner uses to publish events
// while a run executes. Emit is idempotent: the handler tolerates duplicates.
// The store is the concrete implementation in production.
type RunHandle interface {
	// Emit publishes one wire event on the session bus. Kind should be
	// EventStatus / EventToken / EventTool / EventPatch / EventError / EventDone.
	Emit(ctx context.Context, sessionID, runID string, kind domain.EventKind, data any) error
}

// Runner is the Cursor-SDK seam. Plan 2 ships a fake implementation;
// Plan 3 replaces it with an @cursor/sdk-backed one.
//
// Run must return promptly (the handler already responded 202). All
// long-running work happens inside a goroutine the runner owns. Emit a
// terminal Done event before returning; the store releases the run slot
// when it observes a terminal event.
type Runner interface {
	// Name returns a stable identifier for observability and /ready.
	Name() string
	// Run executes one turn on the session. The runner is responsible for
	// its own goroutines; ctx is cancelled by the store on CancelRun or
	// session delete.
	Run(ctx context.Context, sessionID, runID string, in RunInput, h RunHandle) error
}
