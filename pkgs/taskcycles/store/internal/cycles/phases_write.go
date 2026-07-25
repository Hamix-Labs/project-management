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
	"unicode/utf8"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// maxPhaseEventDetailRunes caps each string value inside the optional
// "details" object attached to phase_* audit events so a pathological
// runner payload cannot blow up SSE payloads.
const maxPhaseEventDetailRunes = 8192

// StartPhase appends a new phase row to a running cycle. PhaseSeq is
// assigned by the store (max + 1 within the cycle). Enforces:
//   - cycle exists and is itself running (terminal cycles are read-only);
//   - "at most one running phase per cycle" via in-TX guard;
//   - the requested next phase is allowed by the state machine in
//     cyclesdomain.ValidPhaseTransition, where the previous phase is the
//     highest-seq phase already on this cycle (empty if none).
//
// In the same SQL transaction the call appends an EventPhaseStarted
// mirror row to task_events and writes the assigned task_events.seq
// back into the phase row's event_seq column so the audit pointer is
// one-shot.
func StartPhase(ctx context.Context, db *gorm.DB, cycleID string, phase cyclesdomain.Phase, by taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error) {
	defer storekernel.DeferLatency(storekernel.OpStartCyclePhase)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.StartPhase")
	if err := taskcoredomain.ValidateActor(by); err != nil {
		return nil, err
	}
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return nil, fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	if !cyclesdomain.ValidPhase(phase) {
		return nil, fmt.Errorf("%w: phase", taskcoredomain.ErrInvalidInput)
	}
	var created *cyclesdomain.TaskCyclePhase
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cycle, err := loadByIDInTx(tx, cycleID)
		if err != nil {
			return err
		}
		if cyclesdomain.TerminalCycleStatus(cycle.Status) {
			return fmt.Errorf("%w: cycle is terminal", taskcoredomain.ErrInvalidInput)
		}
		if err := assertNoRunningPhaseForCycleInTx(tx, cycle.ID); err != nil {
			return err
		}
		last, err := lastPhaseForCycleInTx(tx, cycle.ID)
		if err != nil {
			return err
		}
		var prev cyclesdomain.Phase
		if last != nil {
			prev = last.Phase
		}
		if !cyclesdomain.ValidPhaseTransition(prev, phase) && !cyclesdomain.ValidInterruptResumeTransition(last, phase) && !cyclesdomain.ValidVerifyOnlyRetryTransition(last, phase) {
			return fmt.Errorf("%w: phase transition %q -> %q not allowed", taskcoredomain.ErrInvalidInput, prev, phase)
		}
		nextSeq, err := nextPhaseSeqInTx(tx, cycle.ID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		runCorrelationID := uuid.NewString()
		startDetails, err := json.Marshal(map[string]string{
			cyclesdomain.PhaseDetailsRunCorrelationID: runCorrelationID,
		})
		if err != nil {
			return fmt.Errorf("marshal phase start details: %w", err)
		}
		row := &cyclesdomain.TaskCyclePhase{
			ID:          uuid.NewString(),
			CycleID:     cycle.ID,
			Phase:       phase,
			PhaseSeq:    nextSeq,
			Status:      cyclesdomain.PhaseStatusRunning,
			StartedAt:   now,
			DetailsJSON: json.RawMessage(startDetails),
		}
		if err := tx.Create(model.FromDomainTaskCyclePhasePtr(row)).Error; err != nil {
			return fmt.Errorf("insert task_cycle_phase: %w", err)
		}
		evSeq, err := eventsaudit.NextEventSeq(tx, cycle.TaskID)
		if err != nil {
			return err
		}
		payload, err := phaseStartedPayload(cycle.ID, row)
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, cycle.TaskID, evSeq, taskeventsdomain.EventPhaseStarted, by, payload); err != nil {
			return err
		}
		if err := tx.Model(&model.TaskCyclePhase{}).Where("id = ?", row.ID).Update("event_seq", evSeq).Error; err != nil {
			return fmt.Errorf("backfill phase event_seq: %w", err)
		}
		row.EventSeq = &evSeq
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// CompletePhase moves a running phase to a terminal status. Rejected
// when the phase row is missing, already terminal, or the requested
// status is not terminal. Does not move the parent cycle into a
// terminal status — that is an explicit Terminate call so the caller
// controls when an attempt is declared finished.
//
// In the same SQL transaction the call appends an audit mirror to
// task_events (EventPhaseCompleted / EventPhaseFailed /
// EventPhaseSkipped depending on the terminal status) and writes the
// assigned task_events.seq back into the phase row's event_seq column,
// replacing the EventPhaseStarted pointer set at StartPhase time.
func CompletePhase(ctx context.Context, db *gorm.DB, in CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error) {
	defer storekernel.DeferLatency(storekernel.OpCompleteCyclePhase)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.CompletePhase")
	if err := taskcoredomain.ValidateActor(in.By); err != nil {
		return nil, err
	}
	cycleID := strings.TrimSpace(in.CycleID)
	if cycleID == "" {
		return nil, fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	if in.PhaseSeq <= 0 {
		return nil, fmt.Errorf("%w: phase_seq", taskcoredomain.ErrInvalidInput)
	}
	if !cyclesdomain.TerminalPhaseStatus(in.Status) {
		return nil, fmt.Errorf("%w: status must be a terminal phase status", taskcoredomain.ErrInvalidInput)
	}
	details, err := storekernel.NormalizeJSONObject(in.Details, "details", taskcoredomain.ErrInvalidInput)
	if err != nil {
		return nil, err
	}
	var out *cyclesdomain.TaskCyclePhase
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cycle, err := loadByIDInTx(tx, cycleID)
		if err != nil {
			return err
		}
		ph, err := loadPhaseByCycleSeqInTx(tx, cycleID, in.PhaseSeq)
		if err != nil {
			return err
		}
		if cyclesdomain.TerminalPhaseStatus(ph.Status) {
			return fmt.Errorf("%w: phase already terminal", taskcoredomain.ErrInvalidInput)
		}
		mergedDetails, err := mergePhaseDetailsJSON(ph.DetailsJSON, details)
		if err != nil {
			return fmt.Errorf("merge phase details: %w", err)
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"status":       in.Status,
			"ended_at":     now,
			"details_json": datatypes.JSON(mergedDetails),
		}
		if in.Summary != nil {
			updates["summary"] = *in.Summary
		}
		if err := tx.Model(&model.TaskCyclePhase{}).Where("id = ?", ph.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update task_cycle_phase: %w", err)
		}
		ph.Status = in.Status
		ph.EndedAt = &now
		ph.DetailsJSON = json.RawMessage(mergedDetails)
		if in.Summary != nil {
			s := *in.Summary
			ph.Summary = &s
		}
		evSeq, err := eventsaudit.NextEventSeq(tx, cycle.TaskID)
		if err != nil {
			return err
		}
		payload, err := phaseTerminatedPayload(cycle.ID, ph)
		if err != nil {
			return err
		}
		mirrorType := mirrorEventTypeForPhaseStatus(in.Status)
		if err := eventsaudit.AppendEvent(tx, cycle.TaskID, evSeq, mirrorType, in.By, payload); err != nil {
			return err
		}
		if err := tx.Model(&model.TaskCyclePhase{}).Where("id = ?", ph.ID).Update("event_seq", evSeq).Error; err != nil {
			return fmt.Errorf("backfill phase event_seq: %w", err)
		}
		ph.EventSeq = &evSeq
		out = ph
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PatchPhaseDetails shallow-merges patch into details_json for a
// non-terminal phase row. It preserves run_correlation_id and never
// overwrites an existing non-empty session_id (first-wins per ADR-0031
// so mid-run patches from the adapter do not clobber the id already
// stored). Rejects missing or already-terminal phase rows.
func PatchPhaseDetails(ctx context.Context, db *gorm.DB, cycleID string, phaseSeq int64, patch []byte) error {
	defer storekernel.DeferLatency(storekernel.OpCompleteCyclePhase)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.PatchPhaseDetails",
		"cycle_id", cycleID, "phase_seq", phaseSeq, "patch_bytes", len(patch))
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return fmt.Errorf("%w: cycle_id", taskcoredomain.ErrInvalidInput)
	}
	if phaseSeq <= 0 {
		return fmt.Errorf("%w: phase_seq", taskcoredomain.ErrInvalidInput)
	}
	normalized, err := storekernel.NormalizeJSONObject(patch, "patch", taskcoredomain.ErrInvalidInput)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ph, err := loadPhaseByCycleSeqInTx(tx, cycleID, phaseSeq)
		if err != nil {
			return err
		}
		if cyclesdomain.TerminalPhaseStatus(ph.Status) {
			return fmt.Errorf("%w: phase already terminal", taskcoredomain.ErrInvalidInput)
		}
		merged, err := patchPhaseDetailsJSON(ph.DetailsJSON, normalized)
		if err != nil {
			return fmt.Errorf("merge phase details: %w", err)
		}
		if err := tx.Model(&model.TaskCyclePhase{}).Where("id = ?", ph.ID).Update("details_json", datatypes.JSON(merged)).Error; err != nil {
			return fmt.Errorf("update task_cycle_phase details_json: %w", err)
		}
		return nil
	})
}

// patchPhaseDetailsJSON shallow-merges an incoming patch object into the
// existing details_json. run_correlation_id is preserved (never
// overwritten) and an existing non-empty session_id wins over a
// different/empty value in the patch (first-wins, ADR-0031).
//
//funclogmeasure:skip category=hot-path reason="Pure merge helper without I/O; operation trace is emitted by PatchPhaseDetails."
func patchPhaseDetailsJSON(existing json.RawMessage, incoming []byte) ([]byte, error) {
	base := map[string]any{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &base); err != nil {
			return nil, fmt.Errorf("unmarshal existing details: %w", err)
		}
	}
	preservedCorrelationID := cyclesdomain.RunCorrelationIDFromDetailsJSON(existing)
	preservedSessionID := cyclesdomain.SessionIDFromDetailsJSON(existing)

	patch := map[string]any{}
	if len(incoming) > 0 {
		if err := json.Unmarshal(incoming, &patch); err != nil {
			return nil, fmt.Errorf("unmarshal incoming patch: %w", err)
		}
	}
	for k, v := range patch {
		base[k] = v
	}
	if preservedCorrelationID != "" {
		base[cyclesdomain.PhaseDetailsRunCorrelationID] = preservedCorrelationID
	}
	if preservedSessionID != "" {
		if incomingID, _ := patch[cyclesdomain.PhaseDetailsSessionID].(string); strings.TrimSpace(incomingID) != preservedSessionID {
			slog.Warn("taskcycles: PatchPhaseDetails preserved existing session_id",
				"cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.patchPhaseDetailsJSON.session_id_conflict",
				"incoming_session_id_present", strings.TrimSpace(incomingID) != "",
			)
		}
		base[cyclesdomain.PhaseDetailsSessionID] = preservedSessionID
	}
	out, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("marshal merged details: %w", err)
	}
	return out, nil
}

func loadPhaseByCycleSeqInTx(tx *gorm.DB, cycleID string, phaseSeq int64) (*cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.loadPhaseByCycleSeqInTx")
	var p model.TaskCyclePhase
	if err := tx.Where("cycle_id = ? AND phase_seq = ?", cycleID, phaseSeq).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, taskcoredomain.ErrNotFound
		}
		return nil, fmt.Errorf("load task_cycle_phase: %w", err)
	}
	return model.ToDomainTaskCyclePhasePtr(&p), nil
}

// mergePhaseDetailsJSON shallow-merges incoming details over existing
// details_json. run_correlation_id from the existing row is never dropped
// (ADR-0030).
//
//funclogmeasure:skip category=hot-path reason="Pure merge helper without I/O; operation trace is emitted by CompletePhase."
func mergePhaseDetailsJSON(existing json.RawMessage, incoming []byte) ([]byte, error) {
	base := map[string]any{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &base); err != nil {
			return nil, fmt.Errorf("unmarshal existing details: %w", err)
		}
	}
	preservedID := cyclesdomain.RunCorrelationIDFromDetailsJSON(existing)

	patch := map[string]any{}
	if len(incoming) > 0 {
		if err := json.Unmarshal(incoming, &patch); err != nil {
			return nil, fmt.Errorf("unmarshal incoming details: %w", err)
		}
	}
	for k, v := range patch {
		base[k] = v
	}
	if preservedID != "" {
		base[cyclesdomain.PhaseDetailsRunCorrelationID] = preservedID
	}
	out, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("marshal merged details: %w", err)
	}
	return out, nil
}

func assertNoRunningPhaseForCycleInTx(tx *gorm.DB, cycleID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.assertNoRunningPhaseForCycleInTx")
	var n int64
	if err := tx.Model(&model.TaskCyclePhase{}).Where("cycle_id = ? AND status = ?", cycleID, cyclesdomain.PhaseStatusRunning).Count(&n).Error; err != nil {
		return fmt.Errorf("running phase lookup: %w", err)
	}
	if n > 0 {
		return fmt.Errorf("%w: cycle already has a running phase", taskcoredomain.ErrInvalidInput)
	}
	return nil
}

func nextPhaseSeqInTx(tx *gorm.DB, cycleID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.nextPhaseSeqInTx")
	var max int64
	if err := tx.Raw(`SELECT COALESCE(MAX(phase_seq), 0) FROM task_cycle_phases WHERE cycle_id = ?`, cycleID).Scan(&max).Error; err != nil {
		return 0, fmt.Errorf("next phase_seq: %w", err)
	}
	return max + 1, nil
}

// lastPhaseForCycleInTx returns the highest-seq phase row in this cycle,
// or nil when none exist.
func lastPhaseForCycleInTx(tx *gorm.DB, cycleID string) (*cyclesdomain.TaskCyclePhase, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.lastPhaseForCycleInTx")
	var p model.TaskCyclePhase
	err := tx.Where("cycle_id = ?", cycleID).Order("phase_seq DESC").Limit(1).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("last phase lookup: %w", err)
	}
	return model.ToDomainTaskCyclePhasePtr(&p), nil
}

// phaseStartedPayload builds the data_json payload for the
// EventPhaseStarted audit mirror.
func phaseStartedPayload(cycleID string, p *cyclesdomain.TaskCyclePhase) ([]byte, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.phaseStartedPayload")
	out := map[string]any{
		"cycle_id":  cycleID,
		"phase":     string(p.Phase),
		"phase_seq": p.PhaseSeq,
	}
	if id := cyclesdomain.RunCorrelationIDFromDetailsJSON(p.DetailsJSON); id != "" {
		out[cyclesdomain.PhaseDetailsRunCorrelationID] = id
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal phase_started payload: %w", err)
	}
	return b, nil
}

// phaseTerminatedPayload builds the data_json payload for the
// EventPhaseCompleted / EventPhaseFailed / EventPhaseSkipped audit
// mirror.
func phaseTerminatedPayload(cycleID string, p *cyclesdomain.TaskCyclePhase) ([]byte, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.phaseTerminatedPayload")
	out := map[string]any{
		"cycle_id":  cycleID,
		"phase":     string(p.Phase),
		"phase_seq": p.PhaseSeq,
		"status":    string(p.Status),
	}
	if p.Summary != nil && *p.Summary != "" {
		out["summary"] = *p.Summary
	}
	if d := phaseDetailsForEventPayload(p.DetailsJSON); len(d) > 0 {
		out["details"] = d
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal phase_terminated payload: %w", err)
	}
	return b, nil
}

// phaseDetailsForEventPayload returns a deep-copied, size-clamped JSON object
// suitable for task_events.data_json so the timeline can show stderr tails,
// token usage, etc., without a separate cycle-phase fetch.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func phaseDetailsForEventPayload(detailsJSON json.RawMessage) map[string]any {
	if len(detailsJSON) == 0 {
		return nil
	}
	var obj map[string]any
	if err := json.Unmarshal(detailsJSON, &obj); err != nil || len(obj) == 0 {
		return nil
	}
	out := truncatePhaseEventDetailValue(obj, maxPhaseEventDetailRunes)
	if m, ok := out.(map[string]any); ok {
		return m
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncatePhaseEventDetailValue(v any, maxRunes int) any {
	switch x := v.(type) {
	case string:
		return truncateStringRunes(x, maxRunes)
	case map[string]any:
		for k, vv := range x {
			x[k] = truncatePhaseEventDetailValue(vv, maxRunes)
		}
		return x
	case []any:
		for i, vv := range x {
			x[i] = truncatePhaseEventDetailValue(vv, maxRunes)
		}
		return x
	default:
		return v
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncateStringRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= maxRunes {
			b.WriteRune('…')
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// mirrorEventTypeForPhaseStatus picks which audit row type to write
// when a phase reaches the given terminal status.
func mirrorEventTypeForPhaseStatus(s cyclesdomain.PhaseStatus) taskeventsdomain.EventType {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.cycles.mirrorEventTypeForPhaseStatus")
	switch s {
	case cyclesdomain.PhaseStatusSucceeded:
		return taskeventsdomain.EventPhaseCompleted
	case cyclesdomain.PhaseStatusFailed:
		return taskeventsdomain.EventPhaseFailed
	case cyclesdomain.PhaseStatusSkipped:
		return taskeventsdomain.EventPhaseSkipped
	default:
		return taskeventsdomain.EventPhaseFailed
	}
}
