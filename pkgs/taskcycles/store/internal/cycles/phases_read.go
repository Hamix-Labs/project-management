package cycles

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	"gorm.io/gorm"
)

// ListPhasesForCycle returns phases for cycleID in execution order
// (phase_seq ASC). The cycle must exist; an empty result for an
// existing cycle (no phases started yet) is not an error.
func ListPhasesForCycle(ctx context.Context, db *gorm.DB, cycleID string) ([]cyclesdomain.TaskCyclePhase, error) {
	defer storekernel.DeferLatency(storekernel.OpListCyclePhases)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.ListPhasesForCycle")
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return nil, fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	var rows []model.TaskCyclePhase
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := loadByIDInTx(tx, cycleID); err != nil {
			return err
		}
		if err := tx.Where("cycle_id = ?", cycleID).Order("phase_seq ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("list task_cycle_phases: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return model.ToDomainTaskCyclePhases(rows), nil
}

// LastSessionID returns the session_id from the latest terminal phase row
// of the given phase type in cycleID. Empty string means no usable id.
func LastSessionID(ctx context.Context, db *gorm.DB, cycleID string, phase cyclesdomain.Phase) (string, error) {
	defer storekernel.DeferLatency(storekernel.OpListCyclePhases)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.LastSessionID",
		"cycle_id", cycleID, "phase", string(phase))
	phases, err := ListPhasesForCycle(ctx, db, cycleID)
	if err != nil {
		return "", err
	}
	for i := len(phases) - 1; i >= 0; i-- {
		p := phases[i]
		if p.Phase != phase {
			continue
		}
		if !cyclesdomain.TerminalPhaseStatus(p.Status) {
			continue
		}
		if id := cyclesdomain.SessionIDFromDetailsJSON(p.DetailsJSON); id != "" {
			return id, nil
		}
	}
	return "", nil
}
