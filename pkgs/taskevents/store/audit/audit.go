// Package audit owns transactional task_events seq allocation and row insert
// helpers used by peer BC stores (taskcore, taskcycles, taskchecklist) inside
// open GORM transactions. Kept as a leaf under taskevents/store so callers do
// not import store/internal.
package audit

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// taskIDRow is a minimal tasks-table shape for SELECT … FOR UPDATE locking
// without importing pkgs/taskcore/store/model (avoids migrate-graph cycles).
type taskIDRow struct {
	ID string `gorm:"column:id;primaryKey"`
}

func (taskIDRow) TableName() string { return "tasks" }

// NextEventSeq returns the next monotonic seq for taskID inside the open
// transaction tx. Used by every audit-emitting path (CRUD, checklist,
// cycles, phases, devmirror, public AppendTaskEvent).
//
// Concurrency: two transactions racing to append events for the same
// task previously both read MAX(seq) = N and both tried to insert at
// seq = N+1, hitting `task_events_pkey` (composite PK on
// (task_id, seq)) with SQLSTATE 23505 — observed in production from
// parallel POST /tasks/{id}/checklist/items requests fired by the
// create-task modal. We serialize writers per task by row-locking the
// parent `tasks` row (`SELECT ... FOR UPDATE`) before reading
// MAX(seq); the lock is held for the rest of the caller's
// transaction, so the AppendEvent that follows is guaranteed to be
// the only writer at this seq. Lock granularity is the single task
// row (same chokepoint already used by tasks/crud.Update and
// events/thread.MarkResponded), so concurrent appends to *different*
// tasks remain fully parallel.
//
// SQLite is excluded — it serializes all writers globally, so
// `FOR UPDATE` is unnecessary (and unsupported pre-3.45). Mirrors the
// dialect guard in events/thread.MarkResponded.
func NextEventSeq(tx *gorm.DB, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.audit.NextEventSeq")
	if tx.Dialector.Name() != "sqlite" {
		var locked taskIDRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").Where("id = ?", taskID).First(&locked).Error; err != nil {
			return 0, fmt.Errorf("lock task for event seq: %w", err)
		}
	}
	var max int64
	err := tx.Raw(`SELECT COALESCE(MAX(seq), 0) FROM task_events WHERE task_id = ?`, taskID).Scan(&max).Error
	if err != nil {
		return 0, fmt.Errorf("next event seq: %w", err)
	}
	return max + 1, nil
}

// AppendEvent inserts one task_events row inside the open transaction tx.
//
// data is normalized through storekernel.NormalizeJSONObject so the on-disk
// shape of task_events.data_json honours the documented "always a JSON object"
// invariant (see docs/api.md GET /tasks/{id}/events). nil, empty,
// whitespace-only, or the literal "null" all collapse to "{}" so downstream
// consumers (handler readers, SSE fan-out, /events keyset paging) never
// observe SQL NULL or a JSON null literal even if a future caller forgets
// the chokepoint at its own boundary. Non-object payloads (string / number
// / array / bool / malformed) surface as taskcoredomain.ErrInvalidInput so the bug
// is caught at the writing call site instead of leaking past the read-side
// normalizeJSONObjectForResponse defense.
func AppendEvent(tx *gorm.DB, taskID string, seq int64, typ taskeventsdomain.EventType, by taskcoredomain.Actor, data []byte) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.audit.AppendEvent")
	normalized, err := storekernel.NormalizeJSONObject(data, "data", taskcoredomain.ErrInvalidInput)
	if err != nil {
		return err
	}
	data = normalized
	ev := eventsmodel.TaskEvent{
		TaskID: taskID,
		Seq:    seq,
		At:     time.Now().UTC(),
		Type:   typ,
		By:     by,
		Data:   datatypes.JSON(data),
	}
	if err := tx.Create(&ev).Error; err != nil {
		return fmt.Errorf("insert task_event: %w", err)
	}
	return nil
}
