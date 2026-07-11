package storekernel

import (
	"fmt"
	"log/slog"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

// MapPayloadPersistenceError maps driver failures on JSON/JSONB payload columns
// to taskcoredomain.ErrInvalidInput so handlers return 400 with a readable message.
func MapPayloadPersistenceError(err error) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.MapPayloadPersistenceError")
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "invalid input syntax for type json"):
		return fmt.Errorf("%w: payload could not be saved", taskcoredomain.ErrInvalidInput)
	case strings.Contains(msg, "malformed json"):
		return fmt.Errorf("%w: payload could not be saved", taskcoredomain.ErrInvalidInput)
	default:
		return err
	}
}
