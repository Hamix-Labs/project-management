package storekernel

import (
	"fmt"
	"strings"
)

// MapPayloadPersistenceError maps driver failures on JSON/JSONB payload columns
// to invalid so handlers return 400 with a readable message. Callers pass their
// BC invalid-input sentinel.
//
//funclogmeasure:skip category=hot-path reason="Pure error mapper without I/O; callers emit operation trace."
func MapPayloadPersistenceError(err, invalid error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "invalid input syntax for type json"):
		return fmt.Errorf("%w: payload could not be saved", invalid)
	case strings.Contains(msg, "malformed json"):
		return fmt.Errorf("%w: payload could not be saved", invalid)
	default:
		return err
	}
}
