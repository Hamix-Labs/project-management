// Package taskload reads task rows inside open store transactions.
// Owned by taskcore so peer BC stores can load a task without importing
// taskcore/store/internal.
package taskload

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"gorm.io/gorm"
)

// LoadTask reads one taskcoredomain.Task by id inside the open transaction tx and
// maps gorm.ErrRecordNotFound to taskcoredomain.ErrNotFound.
func LoadTask(tx *gorm.DB, id string) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.taskload.LoadTask")
	var row taskmodel.Task
	if err := tx.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, taskcoredomain.ErrNotFound
		}
		return nil, fmt.Errorf("load task: %w", err)
	}
	t := taskmodel.ToDomainTask(row)
	return &t, nil
}
