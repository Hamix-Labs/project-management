package cycles

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/taskload"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type phaseUsageScanRow struct {
	CycleID     string             `gorm:"column:cycle_id"`
	AttemptSeq  int64              `gorm:"column:attempt_seq"`
	Phase       cyclesdomain.Phase `gorm:"column:phase"`
	DetailsJSON datatypes.JSON     `gorm:"column:details_json"`
}

// ListPhaseTokenUsageForTask returns one row per phase with parseable usage
// under details_json for every cycle on the task. Phases without usage are
// omitted. Rows are ordered by attempt_seq ASC, phase_seq ASC.
func ListPhaseTokenUsageForTask(ctx context.Context, db *gorm.DB, taskID string) ([]cyclesdomain.PhaseUsageRow, error) {
	defer storekernel.DeferLatency(storekernel.OpListCyclePhases)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.ListPhaseTokenUsageForTask")
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("%w: task_id", taskcoredomain.ErrInvalidInput)
	}
	var rows []phaseUsageScanRow
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := taskload.LoadTask(tx, taskID); err != nil {
			return err
		}
		if err := tx.Model(&model.TaskCyclePhase{}).
			Select(
				"task_cycles.id AS cycle_id, "+
					"task_cycles.attempt_seq AS attempt_seq, "+
					"task_cycle_phases.phase AS phase, "+
					"task_cycle_phases.details_json AS details_json").
			Joins("INNER JOIN task_cycles ON task_cycles.id = task_cycle_phases.cycle_id").
			Where("task_cycles.task_id = ?", taskID).
			Order("task_cycles.attempt_seq ASC, task_cycle_phases.phase_seq ASC").
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("list phase token usage: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]cyclesdomain.PhaseUsageRow, 0, len(rows))
	for _, r := range rows {
		u, ok := cyclesdomain.TokenUsageFromDetailsJSON(json.RawMessage(r.DetailsJSON))
		if !ok {
			continue
		}
		out = append(out, cyclesdomain.PhaseUsageRow{
			CycleID:    r.CycleID,
			AttemptSeq: r.AttemptSeq,
			Phase:      r.Phase,
			Usage:      u,
		})
	}
	return out, nil
}
