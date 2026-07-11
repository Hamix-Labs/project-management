// Package taskload reads task rows inside open store transactions.
// Split from storekernel so event append paths do not pull in
// pkgs/tasks/store/model (migrate hub cycles with BC model packages).
package taskload

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"errors"
	"fmt"
	"log/slog"

	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"gorm.io/gorm"
)

// LoadTask reads one domain.Task by id inside the open transaction tx and
// maps gorm.ErrRecordNotFound to domain.ErrNotFound.
func LoadTask(tx *gorm.DB, id string) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "storekernel.taskload.LoadTask")
	var row taskmodel.Task
	if err := tx.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("load task: %w", err)
	}
	t := taskmodel.ToDomainTask(row)
	return &t, nil
}
