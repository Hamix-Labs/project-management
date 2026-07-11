package domain

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
)

func scanStringEnum[T ~string](dst *T, value any) error {
	slog.Debug("trace", "operation", "domain.scanStringEnum")
	if value == nil {
		var zero T
		*dst = zero
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*dst = T(string(v))
	case string:
		*dst = T(v)
	default:
		return fmt.Errorf("taskevents: scan %T into enum", value)
	}
	return nil
}

func valueStringEnum[T ~string](s T) (driver.Value, error) {
	slog.Debug("trace", "operation", "domain.valueStringEnum")
	return string(s), nil
}

func (e *EventType) Scan(value any) error { return scanStringEnum(e, value) }
func (e EventType) Value() (driver.Value, error) {
	return valueStringEnum(e)
}
