package reports

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommandRunEntry is one verify-phase shell command execution row.
type CommandRunEntry struct {
	CriterionID string
	CommandSeq  int64
	ExitCode    int
	MetaPath    string
}

// UpsertCommandRuns persists command run metadata for one verify attempt.
func UpsertCommandRuns(ctx context.Context, db *gorm.DB, cycleID string, attemptSeq int64, entries []CommandRunEntry) error {
	defer storekernel.DeferLatency(storekernel.OpUpsertCommandRuns)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.reports.UpsertCommandRuns",
		"cycle_id", cycleID, "attempt_seq", attemptSeq, "entry_count", len(entries))
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	if attemptSeq <= 0 {
		return fmt.Errorf("%w: attempt_seq must be positive", taskcoredomain.ErrInvalidInput)
	}
	if len(entries) == 0 {
		return nil
	}
	now := time.Now().UTC()
	rows := make([]cyclesdomain.TaskCycleCommandRun, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, cyclesdomain.TaskCycleCommandRun{
			ID:          uuid.NewString(),
			CycleID:     cycleID,
			AttemptSeq:  attemptSeq,
			CriterionID: e.CriterionID,
			CommandSeq:  e.CommandSeq,
			ExitCode:    e.ExitCode,
			MetaPath:    e.MetaPath,
			WrittenAt:   now,
		})
	}
	modelRows := model.FromDomainTaskCycleCommandRuns(rows)
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "cycle_id"},
			{Name: "attempt_seq"},
			{Name: "criterion_id"},
			{Name: "command_seq"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"exit_code", "meta_path", "written_at"}),
	}).Create(&modelRows).Error
}

// ListCommandRunsForCycle returns command run rows for cycleID ordered by
// (attempt_seq ASC, criterion_id ASC, command_seq ASC).
func ListCommandRunsForCycle(ctx context.Context, db *gorm.DB, cycleID string) ([]cyclesdomain.TaskCycleCommandRun, error) {
	defer storekernel.DeferLatency(storekernel.OpListCommandRunsForCycle)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.reports.ListCommandRunsForCycle", "cycle_id", cycleID)
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return nil, fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	var rows []model.TaskCycleCommandRun
	err := db.WithContext(ctx).
		Where("cycle_id = ?", cycleID).
		Order("attempt_seq ASC, criterion_id ASC, command_seq ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list command runs: %w", err)
	}
	return model.ToDomainTaskCycleCommandRuns(rows), nil
}
