package domain

import "errors"

// Sentinel errors used by store and handlers for HTTP status mapping.
var (
	// ErrNotFound: session or run id does not exist.
	ErrNotFound = errors.New("draftassist: not found")
	// ErrInvalidInput: request payload failed validation.
	ErrInvalidInput = errors.New("draftassist: invalid input")
	// ErrNonceMismatch: MCP tool called with a stale/wrong nonce.
	ErrNonceMismatch = errors.New("draftassist: nonce mismatch")
	// ErrRunActive: cannot start a new run while one is in flight.
	ErrRunActive = errors.New("draftassist: run active")
	// ErrUnavailable: tool needs a capability not available in this host.
	ErrUnavailable = errors.New("draftassist: unavailable")
	// ErrUnauthorized: MCP tool rejected by taskapi (nonce mismatch or missing).
	ErrUnauthorized = errors.New("draftassist: unauthorized")
)
