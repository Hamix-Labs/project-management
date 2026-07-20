package storekernel

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// MapNotFound translates gorm.ErrRecordNotFound into notFound so handlers can
// use errors.Is without importing gorm. Callers pass their BC sentinel
// (e.g. taskcore/domain.ErrNotFound or projects/domain.ErrNotFound).
//
//funclogmeasure:skip category=hot-path reason="Pure error mapper without I/O; callers emit operation trace."
func MapNotFound(err, notFound error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFound
	}
	return fmt.Errorf("db: %w", err)
}

// MapWriteError maps driver write failures to domain errors. duplicateDetail
// is appended after conflict for unique/duplicate violations. Callers pass
// their BC conflict and invalid-input sentinels.
//
//funclogmeasure:skip category=hot-path reason="Pure error mapper without I/O; callers emit operation trace."
func MapWriteError(err error, duplicateDetail string, conflict, invalid error) error {
	if err == nil {
		return nil
	}
	if IsDuplicateKey(err) {
		return fmt.Errorf("%w: %s", conflict, duplicateDetail)
	}
	if IsCheckConstraintViolation(err) || IsForeignKeyViolation(err) {
		return fmt.Errorf("%w: %v", invalid, err)
	}
	return fmt.Errorf("db: %w", err)
}
