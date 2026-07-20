// Package taskload reads task rows inside open store transactions.
// Split from storekernel so event append paths do not pull in
// pkgs/tasks/store/model (migrate hub cycles with BC model packages).
package taskload

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"errors"
	"fmt"
	"log/slog"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"gorm.io/gorm"
)

// LoadTask reads one taskcoredomain.Task by id inside the open transaction tx and
// maps gorm.ErrRecordNotFound to taskcoredomain.ErrNotFound.
func LoadTask(tx *gorm.DB, id string) (*taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "storekernel.taskload.LoadTask")
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
