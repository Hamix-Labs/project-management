package handler

import (
	"encoding/json"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// Wire shapes for full-mux cycle contract tests. Mirrors pkgs/taskcycles/handler
// response DTOs without importing that package (tests stay black-box on the mux).

const defaultCycleListLimit = 50

type taskCycleResponse struct {
	ID            string              `json:"id"`
	TaskID        string              `json:"task_id"`
	AttemptSeq    int64               `json:"attempt_seq"`
	Status        domain.CycleStatus  `json:"status"`
	StartedAt     time.Time           `json:"started_at"`
	EndedAt       *time.Time          `json:"ended_at,omitempty"`
	TriggeredBy   domain.Actor        `json:"triggered_by"`
	ParentCycleID *string             `json:"parent_cycle_id,omitempty"`
	Meta          json.RawMessage     `json:"meta"`
	CycleMeta     cycleMetaProjection `json:"cycle_meta"`
}

type cycleMetaProjection struct {
	Runner               string `json:"runner"`
	RunnerVersion        string `json:"runner_version"`
	CursorModel          string `json:"cursor_model"`
	CursorModelEffective string `json:"cursor_model_effective"`
	PromptHash           string `json:"prompt_hash"`
}

type taskCyclePhaseResponse struct {
	ID        string             `json:"id"`
	CycleID   string             `json:"cycle_id"`
	Phase     domain.Phase       `json:"phase"`
	PhaseSeq  int64              `json:"phase_seq"`
	Status    domain.PhaseStatus `json:"status"`
	StartedAt time.Time          `json:"started_at"`
	EndedAt   *time.Time         `json:"ended_at,omitempty"`
	Summary   *string            `json:"summary,omitempty"`
	Details   json.RawMessage    `json:"details"`
	EventSeq  *int64             `json:"event_seq,omitempty"`
}

type taskCyclesListResponse struct {
	TaskID               string              `json:"task_id"`
	Cycles               []taskCycleResponse `json:"cycles"`
	Limit                int                 `json:"limit"`
	HasMore              bool                `json:"has_more"`
	NextBeforeAttemptSeq *int64              `json:"next_before_attempt_seq,omitempty"`
}

type taskCycleDetailResponse struct {
	ID            string                   `json:"id"`
	TaskID        string                   `json:"task_id"`
	AttemptSeq    int64                    `json:"attempt_seq"`
	Status        domain.CycleStatus       `json:"status"`
	StartedAt     time.Time                `json:"started_at"`
	EndedAt       *time.Time               `json:"ended_at,omitempty"`
	TriggeredBy   domain.Actor             `json:"triggered_by"`
	ParentCycleID *string                  `json:"parent_cycle_id,omitempty"`
	Meta          json.RawMessage          `json:"meta"`
	CycleMeta     cycleMetaProjection      `json:"cycle_meta"`
	Phases        []taskCyclePhaseResponse `json:"phases"`
}

type taskCycleStreamEventResponse struct {
	ID        string          `json:"id"`
	TaskID    string          `json:"task_id"`
	CycleID   string          `json:"cycle_id"`
	PhaseSeq  int64           `json:"phase_seq"`
	StreamSeq int64           `json:"stream_seq"`
	At        time.Time       `json:"at"`
	Source    string          `json:"source"`
	Kind      string          `json:"kind"`
	Subtype   string          `json:"subtype,omitempty"`
	Message   string          `json:"message,omitempty"`
	Tool      string          `json:"tool,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type taskCycleStreamListResponse struct {
	TaskID       string                         `json:"task_id"`
	CycleID      string                         `json:"cycle_id"`
	Events       []taskCycleStreamEventResponse `json:"events"`
	Limit        int                            `json:"limit"`
	HasMore      bool                           `json:"has_more"`
	NextAfterSeq *int64                         `json:"next_after_seq,omitempty"`
}
