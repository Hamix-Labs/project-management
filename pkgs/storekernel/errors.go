package storekernel

import (
	"errors"
	"fmt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// MapNotFound translates gorm.ErrRecordNotFound into taskcoredomain.ErrNotFound so
// handlers can use errors.Is without importing gorm.
//
//funclogmeasure:skip category=hot-path reason="Pure error mapper without I/O; callers emit operation trace."
func MapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return taskcoredomain.ErrNotFound
	}
	return fmt.Errorf("db: %w", err)
}

// MapWriteError maps driver write failures to domain errors. duplicateDetail
// is appended after taskcoredomain.ErrConflict for unique/duplicate violations.
//
//funclogmeasure:skip category=hot-path reason="Pure error mapper without I/O; callers emit operation trace."
func MapWriteError(err error, duplicateDetail string) error {
	if err == nil {
		return nil
	}
	if IsDuplicateKey(err) {
		return fmt.Errorf("%w: %s", taskcoredomain.ErrConflict, duplicateDetail)
	}
	if IsCheckConstraintViolation(err) || IsForeignKeyViolation(err) {
		return fmt.Errorf("%w: %v", taskcoredomain.ErrInvalidInput, err)
	}
	return fmt.Errorf("db: %w", err)
}
