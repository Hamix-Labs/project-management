package domain

import "errors"

var (
	// ErrNotFound is returned by store methods when a requested project row
	// does not exist. Handlers map this to HTTP 404.
	ErrNotFound = errors.New("projects: not found")

	// ErrInvalidInput is returned when input fails validation. Handlers map
	// this to HTTP 400.
	ErrInvalidInput = errors.New("projects: invalid input")

	// ErrConflict is returned when the request conflicts with persisted state.
	// Handlers map this to HTTP 409.
	ErrConflict = errors.New("projects: conflict")
)
