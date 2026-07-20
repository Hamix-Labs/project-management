package stats

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"gorm.io/gorm"
)

// RecentFailureLimit caps the recent_failures slice on the wire so the
// /tasks/stats payload stays bounded under load.
const RecentFailureLimit = cyclesstore.RecentFailureLimit

// RecentFailure is one row in the recent_failures slice on /tasks/stats.
type RecentFailure = contract.RecentFailure

// scanRecentFailures returns the last `limit` cycle_failed mirror rows
// via the taskcycles-owned failure scanner.
func scanRecentFailures(ctx context.Context, db *gorm.DB, limit int) ([]RecentFailure, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.stats.scanRecentFailures",
		"limit", limit)
	return cyclesstore.ScanRecentFailures(ctx, db, limit)
}
