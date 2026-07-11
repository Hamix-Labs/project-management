package domain

import "errors"

var (
	// ErrInvalidInput is returned when input fails validation. Handlers map
	// this to HTTP 400.
	ErrInvalidInput = errors.New("settings: invalid input")
)
