package tasks

import (
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// allocateNextTaskNumber returns the next per-project task number and bumps
// projects.next_task_number under the caller's transaction. Mirrors
// taskevents NextEventSeq: SELECT … FOR UPDATE on Postgres; SQLite relies on
// global write serialization.
func allocateNextTaskNumber(tx *gorm.DB, projectID string) (int, error) {
	nums, err := allocateNextTaskNumbers(tx, projectID, 1)
	if err != nil {
		return 0, err
	}
	return nums[0], nil
}

// allocateNextTaskNumbers locks the project once, returns k contiguous numbers
// starting at next_task_number, and bumps the counter by k.
func allocateNextTaskNumbers(tx *gorm.DB, projectID string, k int) ([]int, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.allocateNextTaskNumbers",
		"project_id", projectID, "count", k)
	if k < 1 {
		return nil, fmt.Errorf("allocate task numbers: count must be >= 1")
	}
	q := tx.Select("id", "next_task_number").Where("id = ?", projectID)
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var proj projectmodel.Project
	if err := q.First(&proj).Error; err != nil {
		return nil, fmt.Errorf("lock project for task number: %w", err)
	}
	n := proj.NextTaskNumber
	if n < 1 {
		n = 1
	}
	if err := tx.Model(&projectmodel.Project{}).Where("id = ?", projectID).
		Update("next_task_number", n+k).Error; err != nil {
		return nil, fmt.Errorf("bump project task number: %w", err)
	}
	out := make([]int, k)
	for i := 0; i < k; i++ {
		out[i] = n + i
	}
	return out, nil
}
