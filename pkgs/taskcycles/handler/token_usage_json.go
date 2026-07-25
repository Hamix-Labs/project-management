package handler

import (
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

type tokenUsageProjection struct {
	ConsumedTokens        int64 `json:"consumed_tokens"`
	ExecuteConsumedTokens int64 `json:"execute_consumed_tokens"`
	VerifyConsumedTokens  int64 `json:"verify_consumed_tokens"`
	InputTokens           int64 `json:"input_tokens"`
	OutputTokens          int64 `json:"output_tokens"`
	CacheReadTokens       int64 `json:"cache_read_tokens"`
	CacheWriteTokens      int64 `json:"cache_write_tokens"`
	Known                 bool  `json:"known"`
}

type taskTokenUsageAttempt struct {
	CycleID        string               `json:"cycle_id"`
	AttemptSeq     int64                `json:"attempt_seq"`
	TokenUsage     tokenUsageProjection `json:"token_usage"`
	ShareOfTaskPct *float64             `json:"share_of_task_pct"`
}

type taskTokenUsageResponse struct {
	TaskID     string                  `json:"task_id"`
	TokenUsage tokenUsageProjection    `json:"token_usage"`
	Attempts   []taskTokenUsageAttempt `json:"attempts"`
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func projectTokenUsage(exec, verify cyclesdomain.TokenUsage, known bool) tokenUsageProjection {
	if !known {
		return tokenUsageProjection{Known: false}
	}
	total := cyclesdomain.AddTokenUsage(exec, verify)
	return tokenUsageProjection{
		ConsumedTokens:        total.Consumed(),
		ExecuteConsumedTokens: exec.Consumed(),
		VerifyConsumedTokens:  verify.Consumed(),
		InputTokens:           total.InputTokens,
		OutputTokens:          total.OutputTokens,
		CacheReadTokens:       total.CacheReadTokens,
		CacheWriteTokens:      total.CacheWriteTokens,
		Known:                 true,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func projectTokenUsageFromRows(rows []cyclesdomain.PhaseUsageRow) tokenUsageProjection {
	exec, verify := cyclesdomain.SumPhaseUsageByKind(rows)
	return projectTokenUsage(exec, verify, len(rows) > 0)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func projectCycleTokenUsage(rows []cyclesdomain.PhaseUsageRow, cycleID string) tokenUsageProjection {
	filtered := make([]cyclesdomain.PhaseUsageRow, 0, len(rows))
	for _, r := range rows {
		if r.CycleID == cycleID {
			filtered = append(filtered, r)
		}
	}
	return projectTokenUsageFromRows(filtered)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func groupUsageRowsByCycleID(rows []cyclesdomain.PhaseUsageRow) map[string][]cyclesdomain.PhaseUsageRow {
	out := make(map[string][]cyclesdomain.PhaseUsageRow, len(rows))
	for _, r := range rows {
		out[r.CycleID] = append(out[r.CycleID], r)
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func phaseUsageRowsFromPhases(cycleID string, attemptSeq int64, phases []cyclesdomain.TaskCyclePhase) []cyclesdomain.PhaseUsageRow {
	out := make([]cyclesdomain.PhaseUsageRow, 0, len(phases))
	for i := range phases {
		p := &phases[i]
		u, ok := cyclesdomain.TokenUsageFromDetailsJSON(p.DetailsJSON)
		if !ok {
			continue
		}
		out = append(out, cyclesdomain.PhaseUsageRow{
			CycleID:    cycleID,
			AttemptSeq: attemptSeq,
			Phase:      p.Phase,
			Usage:      u,
		})
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func shareOfTaskPct(cycleConsumed, taskConsumed int64, known bool) *float64 {
	if !known || taskConsumed == 0 || cycleConsumed == 0 {
		return nil
	}
	pct := float64(cycleConsumed) / float64(taskConsumed) * 100
	return &pct
}
