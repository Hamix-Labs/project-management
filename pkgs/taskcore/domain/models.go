package domain

import (
	"encoding/json"
	"time"
)

type Task struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	InitialPrompt string   `json:"initial_prompt"`
	Status        Status   `json:"status"`
	Priority      Priority `json:"priority"`
	ProjectID     *string  `json:"project_id,omitempty"`
	// RepositoryID is the registered repo for managed-worktree allocate
	// (ADR-0083 / ADR-0094). Required on create for the allocate path; persisted
	// so reconcile works when the project is the global Default (null repo).
	RepositoryID *string `json:"repository_id,omitempty"`
	// Number is the per-project human-facing task ref (#N). Assigned when
	// project_id is set; immutable thereafter. Null when the task has no project.
	Number    *int      `json:"number,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Milestone *string   `json:"milestone,omitempty"`
	Gate      *TaskGate `json:"gate,omitempty"`
	// DependsOn is hydrated from task_dependencies on read; not a database column.
	DependsOn []DependencyEdge `json:"depends_on,omitempty"`
	// Runner is the agent runner id for this task (e.g. "cursor"). Set at
	// create time from the request or app defaults; must match the worker's
	// configured runner when the task runs.
	Runner string `json:"runner"`
	// CursorModel is forwarded to cursor-agent as --model when non-empty;
	// empty means omit the flag for this task (same semantics as app settings).
	CursorModel string `json:"cursor_model"`
	// RunnerConfig stores per-task runner config overrides as a JSON blob.
	// When non-empty, the worker merges this with the global runner config
	// from app_settings.runner_configs for the matching runner ID.
	RunnerConfig json.RawMessage `json:"runner_config,omitempty"`
	// PickupNotBefore defers agent dequeue until this instant (UTC). NULL means
	// eligible as soon as status is ready (legacy rows and zero-delay creates).
	PickupNotBefore *time.Time `json:"pickup_not_before,omitempty"`
	// CriteriaSatisfiedAt is set when every inherited checklist item has a
	// verified completion row; cleared when any item becomes unchecked.
	// Maintained in checklist completion TX for SQL queue parity.
	CriteriaSatisfiedAt *time.Time `json:"criteria_satisfied_at,omitempty"`
	// PendingRetry holds operator retry/polish intent between POST /retry or
	// POST /polish and worker pickup. Not exposed on the public task API (json:"-").
	PendingRetry *PendingRetry `json:"-"`
	// CreatedAt is hydrated from the seq=1 task_created audit row on read;
	// not a tasks-table column.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// WorktreeID binds the task to a registered git worktree row.
	WorktreeID *string `json:"worktree_id,omitempty"`
}

// TaskDependency is a directed edge: task_id depends on depends_on_task_id completing first.
type TaskDependency struct {
	TaskID          string              `json:"task_id"`
	DependsOnTaskID string              `json:"depends_on_task_id"`
	Satisfies       DependencySatisfies `json:"satisfies"`
	CreatedAt       time.Time           `json:"created_at"`
}
