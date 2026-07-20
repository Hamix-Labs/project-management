package ready

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"gorm.io/gorm"
)

// DeferredPickup is a ready task with pickup_not_before still in the future.
type DeferredPickup = taskcorecontract.DeferredPickup

// DeferredPickupCursor is a keyset cursor for ListDeferredReadyPickups pagination.
type DeferredPickupCursor = taskcorecontract.DeferredPickupCursor

// ListDeferredReadyPickups returns ready tasks whose pickup_not_before is
// strictly after `now`, ordered by pickup time then id. Used to hydrate the
// pickup wake scheduler at startup. When after is non-nil, resumes after that
// keyset position (exclusive).
func ListDeferredReadyPickups(ctx context.Context, db *gorm.DB, now time.Time, limit int, after *DeferredPickupCursor) ([]DeferredPickup, error) {
	defer storekernel.DeferLatency(storekernel.OpListReadyQueue)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ready.ListDeferredReadyPickups")
	if limit <= 0 {
		limit = 10_000
	}
	if limit > 50_000 {
		limit = 50_000
	}
	var rows []struct {
		ID              string
		PickupNotBefore time.Time
	}
	q := db.WithContext(ctx).Model(&model.Task{}).
		Select("id", "pickup_not_before").
		Where("status = ?", domain.StatusReady).
		Where("pickup_not_before IS NOT NULL").
		Where("pickup_not_before > ?", now.UTC())
	if after != nil && after.ID != "" && !after.NotBefore.IsZero() {
		q = q.Where(
			"(pickup_not_before > ?) OR (pickup_not_before = ? AND id > ?)",
			after.NotBefore.UTC(), after.NotBefore.UTC(), after.ID,
		)
	}
	q = q.Order("pickup_not_before ASC, id ASC").Limit(limit)
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list deferred ready pickups: %w", err)
	}
	out := make([]DeferredPickup, 0, len(rows))
	for i := range rows {
		out = append(out, DeferredPickup{
			ID:              rows[i].ID,
			PickupNotBefore: rows[i].PickupNotBefore,
		})
	}
	return out, nil
}
