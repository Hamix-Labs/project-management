package cycles

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/taskload"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Start creates a new TaskCycle row with status=running. Enforces "at
// most one running cycle per task" via an in-TX guard (portable across
// Postgres + SQLite); concurrent attempts surface as
// taskcoredomain.ErrInvalidInput. AttemptSeq is assigned by the store
// (max(attempt_seq) + 1). ParentCycleID, when non-nil, must reference
// a cycle row on the same task; cross-task lineage is rejected.
//
// In the same SQL transaction the call appends an EventCycleStarted
// mirror row to task_events so GET /tasks/{id}/events stays a complete
// witness of cycle activity. If the mirror insert fails, the cycle row
// is rolled back.
func Start(ctx context.Context, db *gorm.DB, in StartCycleInput) (*cyclesdomain.TaskCycle, error) {
	defer storekernel.DeferLatency(storekernel.OpStartCycle)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.Start")
	if err := taskcoredomain.ValidateActor(in.TriggeredBy); err != nil {
		return nil, err
	}
	taskID := strings.TrimSpace(in.TaskID)
	if taskID == "" {
		return nil, fmt.Errorf("%w: task_id", taskcoredomain.ErrInvalidInput)
	}
	meta, err := storekernel.NormalizeJSONObject(in.Meta, "meta", taskcoredomain.ErrInvalidInput)
	if err != nil {
		return nil, err
	}
	var created *cyclesdomain.TaskCycle
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := taskload.LoadTask(tx, taskID); err != nil {
			return err
		}
		if err := assertNoRunningCycleForTaskInTx(tx, taskID); err != nil {
			return err
		}
		if in.ParentCycleID != nil {
			parentID := strings.TrimSpace(*in.ParentCycleID)
			if parentID == "" {
				return fmt.Errorf("%w: parent_cycle_id", taskcoredomain.ErrInvalidInput)
			}
			parent, err := loadByIDInTx(tx, parentID)
			if err != nil {
				return err
			}
			if parent.TaskID != taskID {
				return fmt.Errorf("%w: parent_cycle_id does not belong to this task", taskcoredomain.ErrInvalidInput)
			}
			in.ParentCycleID = &parent.ID
		}
		nextAttempt, err := nextAttemptSeqInTx(tx, taskID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		row := &cyclesdomain.TaskCycle{
			ID:            uuid.NewString(),
			TaskID:        taskID,
			AttemptSeq:    nextAttempt,
			Status:        cyclesdomain.CycleStatusRunning,
			StartedAt:     now,
			TriggeredBy:   string(in.TriggeredBy),
			ParentCycleID: in.ParentCycleID,
			MetaJSON:      json.RawMessage(meta),
		}
		if err := tx.Create(model.FromDomainTaskCyclePtr(row)).Error; err != nil {
			return fmt.Errorf("insert task_cycle: %w", err)
		}
		seq, err := eventsaudit.NextEventSeq(tx, taskID)
		if err != nil {
			return err
		}
		payload, err := startedPayload(row)
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, taskID, seq, taskeventsdomain.EventCycleStarted, in.TriggeredBy, payload); err != nil {
			return err
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Terminate moves a running cycle into a terminal state. Rejected when
// the cycle is already terminal, when the requested status is not
// terminal, or when the cycle still has a running phase row.
//
// In the same SQL transaction the call appends a mirror row to
// task_events: EventCycleCompleted for CycleStatusSucceeded;
// EventCycleFailed for CycleStatusFailed and CycleStatusAborted (the
// mirror payload's status field preserves the distinction between
// failed and aborted). reason, if non-empty, is included in the mirror
// payload.
func Terminate(ctx context.Context, db *gorm.DB, cycleID string, status cyclesdomain.CycleStatus, reason string, by taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error) {
	defer storekernel.DeferLatency(storekernel.OpTerminateCycle)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.Terminate")
	if err := taskcoredomain.ValidateActor(by); err != nil {
		return nil, err
	}
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return nil, fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	if !cyclesdomain.TerminalCycleStatus(status) {
		return nil, fmt.Errorf("%w: status must be a terminal cycle status", taskcoredomain.ErrInvalidInput)
	}
	reason = strings.TrimSpace(reason)
	var out *cyclesdomain.TaskCycle
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cycle, err := loadByIDInTx(tx, cycleID)
		if err != nil {
			return err
		}
		if cyclesdomain.TerminalCycleStatus(cycle.Status) {
			return fmt.Errorf("%w: cycle already terminal", taskcoredomain.ErrInvalidInput)
		}
		if err := assertNoRunningPhaseForCycleInTx(tx, cycle.ID); err != nil {
			return err
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"status":   status,
			"ended_at": now,
		}
		if err := tx.Model(&model.TaskCycle{}).Where("id = ?", cycle.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update task_cycle: %w", err)
		}
		cycle.Status = status
		cycle.EndedAt = &now

		failureSummary := ""
		if mirrorEventTypeForCycleStatus(status) == taskeventsdomain.EventCycleFailed {
			var lastFailedExecute model.TaskCyclePhase
			q := tx.Where("cycle_id = ? AND phase = ? AND status = ?", cycle.ID, cyclesdomain.PhaseExecute, cyclesdomain.PhaseStatusFailed).
				Order("phase_seq DESC")
			if err := q.First(&lastFailedExecute).Error; err == nil {
				lastPhase := model.ToDomainTaskCyclePhase(lastFailedExecute)
				var details map[string]any
				if len(lastPhase.DetailsJSON) > 0 {
					_ = json.Unmarshal(lastPhase.DetailsJSON, &details)
				}
				sum := ""
				if lastPhase.Summary != nil {
					sum = *lastPhase.Summary
				}
				failureSummary = FailureSurfaceMessage(true, reason, sum, details)
			}
		}

		seq, err := eventsaudit.NextEventSeq(tx, cycle.TaskID)
		if err != nil {
			return err
		}
		payload, err := terminatedPayload(cycle, reason, failureSummary)
		if err != nil {
			return err
		}
		mirrorType := mirrorEventTypeForCycleStatus(status)
		if err := eventsaudit.AppendEvent(tx, cycle.TaskID, seq, mirrorType, by, payload); err != nil {
			return err
		}
		out = cycle
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one cycle by id; ErrNotFound when missing.
func Get(ctx context.Context, db *gorm.DB, cycleID string) (*cyclesdomain.TaskCycle, error) {
	defer storekernel.DeferLatency(storekernel.OpGetCycle)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.Get")
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return nil, fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	return loadByIDInTx(db.WithContext(ctx), cycleID)
}

// ListForTask returns cycles for a task ordered by attempt_seq DESC
// (newest first). limit is clamped to [1, 200]; the task must exist.
func ListForTask(ctx context.Context, db *gorm.DB, taskID string, limit int) ([]cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.ListForTask")
	return ListForTaskBefore(ctx, db, taskID, 0, limit)
}

// ListForTaskBefore is the keyset-paginated form of ListForTask. When
// beforeAttemptSeq > 0 the result is restricted to cycles whose
// attempt_seq is strictly less than beforeAttemptSeq (i.e. the next page
// of older cycles past a cursor the caller already saw). beforeAttemptSeq
// <= 0 is equivalent to ListForTask (no cursor / first page). Ordering
// and limit clamping match ListForTask exactly so the two callers share
// the same storekernel.OpListCyclesForTask Prometheus label and the same
// envelope shape on the wire.
//
// Cursor semantics: the page is always sorted attempt_seq DESC (newest
// first) so handlers paginate by handing the *last* (oldest) row's
// attempt_seq back as the next beforeAttemptSeq. Strict < (rather than
// <=) keeps the cursor row from being repeated across pages and matches
// the existing /events `before_seq` convention.
func ListForTaskBefore(ctx context.Context, db *gorm.DB, taskID string, beforeAttemptSeq int64, limit int) ([]cyclesdomain.TaskCycle, error) {
	defer storekernel.DeferLatency(storekernel.OpListCyclesForTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.ListForTaskBefore")
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("%w: task_id", taskcoredomain.ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var rows []model.TaskCycle
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := taskload.LoadTask(tx, taskID); err != nil {
			return err
		}
		q := tx.Where("task_id = ?", taskID)
		if beforeAttemptSeq > 0 {
			q = q.Where("attempt_seq < ?", beforeAttemptSeq)
		}
		if err := q.Order("attempt_seq DESC").Limit(limit).Find(&rows).Error; err != nil {
			return fmt.Errorf("list task_cycles: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return model.ToDomainTaskCycles(rows), nil
}

// loadByIDInTx fetches one cycle by id with gorm errors mapped to the
// domain sentinels. Shared with the phase code via the in-TX scope.
func loadByIDInTx(tx *gorm.DB, cycleID string) (*cyclesdomain.TaskCycle, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.loadByIDInTx")
	var c model.TaskCycle
	if err := tx.Where("id = ?", cycleID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, taskcoredomain.ErrNotFound
		}
		return nil, fmt.Errorf("load task_cycle: %w", err)
	}
	return model.ToDomainTaskCyclePtr(&c), nil
}

func assertNoRunningCycleForTaskInTx(tx *gorm.DB, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.assertNoRunningCycleForTaskInTx")
	var n int64
	if err := tx.Model(&model.TaskCycle{}).Where("task_id = ? AND status = ?", taskID, cyclesdomain.CycleStatusRunning).Count(&n).Error; err != nil {
		return fmt.Errorf("running cycle lookup: %w", err)
	}
	if n > 0 {
		return fmt.Errorf("%w: task already has a running cycle", taskcoredomain.ErrInvalidInput)
	}
	return nil
}

func nextAttemptSeqInTx(tx *gorm.DB, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.nextAttemptSeqInTx")
	var max int64
	if err := tx.Raw(`SELECT COALESCE(MAX(attempt_seq), 0) FROM task_cycles WHERE task_id = ?`, taskID).Scan(&max).Error; err != nil {
		return 0, fmt.Errorf("next attempt_seq: %w", err)
	}
	return max + 1, nil
}

// startedPayload builds the data_json payload for the EventCycleStarted
// audit mirror. Keys are stable (asserted by the dual-write invariant test).
func startedPayload(c *cyclesdomain.TaskCycle) ([]byte, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.startedPayload")
	out := map[string]any{
		"cycle_id":     c.ID,
		"attempt_seq":  c.AttemptSeq,
		"triggered_by": string(c.TriggeredBy),
	}
	if c.ParentCycleID != nil && *c.ParentCycleID != "" {
		out["parent_cycle_id"] = *c.ParentCycleID
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal cycle_started payload: %w", err)
	}
	return b, nil
}

// terminatedPayload builds the data_json payload for the
// EventCycleCompleted / EventCycleFailed audit mirror.
func terminatedPayload(c *cyclesdomain.TaskCycle, reason, failureSummary string) ([]byte, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.terminatedPayload")
	out := map[string]any{
		"cycle_id":    c.ID,
		"attempt_seq": c.AttemptSeq,
		"status":      string(c.Status),
	}
	if reason != "" {
		out["reason"] = reason
	}
	if strings.TrimSpace(failureSummary) != "" {
		out["failure_summary"] = strings.TrimSpace(failureSummary)
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal cycle_terminated payload: %w", err)
	}
	return b, nil
}

// mirrorEventTypeForCycleStatus picks which audit row type to write
// when a cycle reaches the given terminal status. CycleStatusAborted
// folds into EventCycleFailed; the payload's status field preserves
// the distinction.
func mirrorEventTypeForCycleStatus(s cyclesdomain.CycleStatus) taskeventsdomain.EventType {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.mirrorEventTypeForCycleStatus")
	if s == cyclesdomain.CycleStatusSucceeded {
		return taskeventsdomain.EventCycleCompleted
	}
	return taskeventsdomain.EventCycleFailed
}
