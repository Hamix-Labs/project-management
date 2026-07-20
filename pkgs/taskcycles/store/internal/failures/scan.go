package failures

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/cycles"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RecentFailureLimit caps the recent_failures slice on /tasks/stats.
const RecentFailureLimit = 25

// CycleFailure is the operator-facing cycle failure projection.
type CycleFailure = contract.CycleFailure

// cycleFailedRow is the raw projection from task_events for one
// cycle_failed mirror; we unmarshal data_json in Go (rather than
// jsonb-extract in SQL) so the scanner stays portable across Postgres
// and SQLite.
type cycleFailedRow struct {
	TaskID string
	Seq    int64
	At     time.Time
	Data   datatypes.JSON `gorm:"column:data_json"`
}

// cycleFailedPayload mirrors the keys terminatedPayload writes for
// EventCycleFailed in pkgs/taskcycles/store/internal/cycles/cycles.go.
type cycleFailedPayload struct {
	CycleID        string `json:"cycle_id"`
	AttemptSeq     int64  `json:"attempt_seq"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	FailureSummary string `json:"failure_summary,omitempty"`
}

// ScanRecent returns the last `limit` cycle_failed mirror rows ordered
// by event timestamp descending (newest first).
func ScanRecent(ctx context.Context, db *gorm.DB, limit int) ([]CycleFailure, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.failures.ScanRecent",
		"limit", limit)
	if limit <= 0 || limit > RecentFailureLimit {
		limit = RecentFailureLimit
	}
	var rows []cycleFailedRow
	if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
		Select("task_id, seq, at, data_json").
		Where("type = ?", string(taskeventsdomain.EventCycleFailed)).
		Order("at DESC, seq DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("recent failures: %w", err)
	}
	out := decodeCycleFailedRows(rows)
	enrichFromPhaseEvents(ctx, db, out)
	return out, nil
}

func decodeCycleFailedRows(rows []cycleFailedRow) []CycleFailure {
	out := make([]CycleFailure, 0, len(rows))
	for _, r := range rows {
		var p cycleFailedPayload
		if err := json.Unmarshal(r.Data, &p); err != nil {
			slog.Debug("recent failure decode skipped",
				"cmd", calltrace.LogCmd,
				"operation", "taskcycles.store.failures.decodeCycleFailedRows",
				"task_id", r.TaskID, "seq", r.Seq, "err", err)
			continue
		}
		out = append(out, CycleFailure{
			TaskID:     r.TaskID,
			EventSeq:   r.Seq,
			At:         r.At,
			CycleID:    p.CycleID,
			AttemptSeq: p.AttemptSeq,
			Status:     p.Status,
			Reason:     resolveReason(p.FailureSummary, p.Reason),
		})
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func resolveReason(failureSummary, cycleReason string) string {
	if s := strings.TrimSpace(failureSummary); s != "" {
		return s
	}
	return cycleReason
}

type phaseFailedMirrorPayload struct {
	CycleID string         `json:"cycle_id"`
	Summary string         `json:"summary,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

func enrichFromPhaseEvents(ctx context.Context, db *gorm.DB, failures []CycleFailure) {
	if len(failures) == 0 {
		return
	}
	needed := make(map[string]struct{}, len(failures))
	taskSeen := make(map[string]struct{})
	var taskIDs []string
	for _, f := range failures {
		needed[f.TaskID+"\x00"+f.CycleID] = struct{}{}
		if _, ok := taskSeen[f.TaskID]; !ok {
			taskSeen[f.TaskID] = struct{}{}
			taskIDs = append(taskIDs, f.TaskID)
		}
	}
	type row struct {
		TaskID string
		Seq    int64
		Data   datatypes.JSON `gorm:"column:data_json"`
	}
	var rows []row
	if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
		Select("task_id, seq, data_json").
		Where("type = ?", string(taskeventsdomain.EventPhaseFailed)).
		Where("task_id IN ?", taskIDs).
		Order("seq DESC").
		Limit(5000).
		Scan(&rows).Error; err != nil {
		slog.Debug("recent failures phase enrich skipped", "cmd", calltrace.LogCmd,
			"operation", "taskcycles.store.failures.enrichFromPhaseEvents",
			"err", err)
		return
	}
	phaseByKey := make(map[string]*phaseFailedMirrorPayload)
	for _, r := range rows {
		var p phaseFailedMirrorPayload
		if err := json.Unmarshal(r.Data, &p); err != nil || p.CycleID == "" {
			continue
		}
		k := r.TaskID + "\x00" + p.CycleID
		if _, want := needed[k]; !want {
			continue
		}
		if _, have := phaseByKey[k]; have {
			continue
		}
		pp := p
		phaseByKey[k] = &pp
		if len(phaseByKey) == len(needed) {
			break
		}
	}
	for i := range failures {
		k := failures[i].TaskID + "\x00" + failures[i].CycleID
		ph := phaseByKey[k]
		if r := observabilityReasonFromPhaseAndCycle(failures[i].Reason, ph); r != "" {
			failures[i].Reason = r
		}
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func observabilityReasonFromPhaseAndCycle(cycleReason string, phase *phaseFailedMirrorPayload) string {
	if phase == nil {
		return ""
	}
	return cycles.FailureSurfaceMessage(true, cycleReason, strings.TrimSpace(phase.Summary), phase.Details)
}
