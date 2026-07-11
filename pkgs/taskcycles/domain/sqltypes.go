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
		return fmt.Errorf("taskcycles: scan %T into enum", value)
	}
	return nil
}

func valueStringEnum[T ~string](s T) (driver.Value, error) {
	slog.Debug("trace", "operation", "domain.valueStringEnum")
	return string(s), nil
}

func (p *Phase) Scan(value any) error              { return scanStringEnum(p, value) }
func (p Phase) Value() (driver.Value, error)       { return valueStringEnum(p) }
func (s *CycleStatus) Scan(value any) error        { return scanStringEnum(s, value) }
func (s CycleStatus) Value() (driver.Value, error) { return valueStringEnum(s) }
func (s *PhaseStatus) Scan(value any) error        { return scanStringEnum(s, value) }
func (s PhaseStatus) Value() (driver.Value, error) { return valueStringEnum(s) }
