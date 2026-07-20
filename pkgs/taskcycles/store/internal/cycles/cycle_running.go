package cycles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	"gorm.io/gorm"
)

// IsTaskCycleRunning reports whether taskID has a task_cycles row with status=running.
// When no running cycle exists, it still verifies the task row exists (ErrNotFound otherwise).
func IsTaskCycleRunning(ctx context.Context, db *gorm.DB, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.IsTaskCycleRunning")
	return isTaskCycleRunningInTx(db.WithContext(ctx), taskID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func isTaskCycleRunningInTx(tx *gorm.DB, taskID string) (bool, error) {
	var n int64
	if err := tx.Model(&cyclesmodel.TaskCycle{}).
		Where("task_id = ? AND status = ?", taskID, cyclesdomain.CycleStatusRunning).
		Count(&n).Error; err != nil {
		return false, fmt.Errorf("running cycle lookup: %w", err)
	}
	if n > 0 {
		return true, nil
	}
	var row taskmodel.Task
	if err := tx.Where("id = ?", taskID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, taskcoredomain.ErrNotFound
		}
		return false, fmt.Errorf("load task: %w", err)
	}
	return false, nil
}
