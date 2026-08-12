package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
)

// CreateSessionInput binds a new session to a worktree with an initial form
// snapshot. WorktreeID may be empty for tests, but production callers should
// pass the operator's selected worktree.
type CreateSessionInput struct {
	WorktreeID string
	Snapshot   domain.FormSnapshot
}

// RunInput seeds one run inside a session with a fresh user message.
// The runner sees Snapshot as of accept time; later edits are visible only
// on the next run.
type RunInput struct {
	UserMessage string
	Snapshot    domain.FormSnapshot
}

// Subscription is a live event feed on one session. Events is closed when
// the session is deleted; Cancel unregisters without touching the session.
type Subscription struct {
	Events chan domain.Event
	Cancel func()
}

// Store is the in-memory session repository plus event bus.
//
// Sessions are keyed by id; a per-session nonce authenticates MCP tools that
// mutate the form. Snapshot is the operator's latest form draft; the runner
// captures it into the run at accept time.
type Store interface {
	// CreateSession returns a new session with a fresh id and nonce.
	CreateSession(ctx context.Context, in CreateSessionInput) (*domain.Session, error)
	// GetSession returns a snapshot copy of the session record.
	GetSession(ctx context.Context, id string) (*domain.Session, error)
	// UpdateSnapshot overwrites the form snapshot on the session.
	UpdateSnapshot(ctx context.Context, id string, snap domain.FormSnapshot) (*domain.Session, error)
	// UpdatePrompt writes only the prompt field. Fails closed on nonce mismatch.
	UpdatePrompt(ctx context.Context, id, nonce, prompt string) (*domain.Session, error)
	// DeleteSession removes the session and closes any subscriptions.
	DeleteSession(ctx context.Context, id string) error

	// StartRun accepts a new run, marks it in-flight, and returns the run id.
	// Returns ErrRunActive when an existing run has not terminated.
	StartRun(ctx context.Context, id string, in RunInput) (runID string, err error)
	// CancelRun signals cancellation; the runner is expected to emit a
	// terminal event that FinishRun clears.
	CancelRun(ctx context.Context, sessionID, runID string) error
	// FinishRun releases the run slot after the runner emits a terminal event.
	FinishRun(ctx context.Context, sessionID, runID string) error
	// RunActive reports whether a run is currently in flight on the session.
	RunActive(ctx context.Context, sessionID string) (bool, error)

	// Publish appends an event to the session ring and fans it out to
	// subscribers. Kind must be one of the domain.Event* constants.
	Publish(ctx context.Context, sessionID string, ev domain.Event) error
	// Subscribe registers a live subscriber, optionally replaying events
	// after sinceID. Cancel is safe to call multiple times.
	Subscribe(ctx context.Context, sessionID string, sinceID uint64) (*Subscription, []domain.Event, error)
}
