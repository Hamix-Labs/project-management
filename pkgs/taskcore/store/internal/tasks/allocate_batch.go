package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// AllocateNextTaskNumbers locks the project once and returns k contiguous
// task numbers for batch instantiate.
func AllocateNextTaskNumbers(ctx context.Context, db *gorm.DB, projectID string, k int) ([]int, error) {
	defer storekernel.DeferLatency(storekernel.OpCreateTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.AllocateNextTaskNumbers",
		"project_id", projectID, "count", k)
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("%w: project id", domain.ErrInvalidInput)
	}
	var out []int
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		nums, err := allocateNextTaskNumbers(tx, projectID, k)
		if err != nil {
			return err
		}
		out = nums
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
