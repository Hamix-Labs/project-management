package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SchemaRevision is bumped in the same PR as any change to domain models or
// idempotent post-AutoMigrate steps in Migrate.
//
// Rev 2 (ADR-0037 expand phase): adds worktree_branches, git_worktrees
// .active_branch_id, projects.repository_id, tasks.worktree_branch_id and their
// idempotent backfill.
//
// Rev 4 (ADR-0039): git_worktrees.branch_id, tasks.worktree_id; drops
// worktree_branches, tasks.worktree_branch_id, git_worktrees.active_branch_id.
//
// Rev 5: git_repositories.git_common_dir (unique repo identity); main path normalization.
//
// Rev 6 (ADR-0042): projects.is_default; per-repo default projects; remove global default row.
//
// Rev 7: backfill worktree_id in task_templates / task_drafts compose payloads.
//
// Rev 8 (Tier 3 BC blueprint): domain types colocated to taskchecklist,
// taskevents, taskcycles, and taskcore; BC domain packages are canonical.
// No SQL or post-AutoMigrate behavior change.
//
// Rev 9 (ADR-0079): delete pkgs/tasks/store facade; postgres.Migrate calls
// BC store models via migrate_models.go. No SQL or post-AutoMigrate change.
//
// Rev 10: peel one-shot migrate steps into pkgs/tasks/postgres/migrate.
// No SQL or post-AutoMigrate behavior change.
//
// Rev 11: purge projects whose repository_id no longer exists (orphans left
// when repositories were deleted without cascading project rows).
//
// Rev 12: project_context_items.description — short selection blurb for
// memory nodes (optional; empty string default for existing rows).
//
// Rev 13: attach non-default projects with null repository_id to the sole
// registered repository so create-task repo-scoped pickers list them.
//
// Rev 14: PendingRetry polish fields (flagged/new criterion IDs, skip_verify)
// in taskcore domain JSON. No SQL or post-AutoMigrate behavior change.
//
// Rev 15: app_settings.verify_model — optional Cursor --model for PhaseVerify
// on the same chat as execute (empty inherits execute effective model).
//
// Rev 16: task_events composite index (type, at, seq) for GET /tasks/activity
// cross-task filtered pagination (status_changed, phase_failed, approval_granted).
//
// Rev 17: drop app_settings.stream_idle_stuck_seconds (and its check
// constraint) — Hamix no longer runs its own stream-idle kill; runners
// own their own timeouts. Only max_run_duration_seconds remains as the
// wall-clock cap.
//
// Rev 18: tasks.number + projects.next_task_number; backfill dense
// per-project numbers for existing tasks (display ref #N).
//
// Rev 19: tasks.status CHECK includes closed (close/reopen replaces hard delete).
//
// Rev 20 (ADR-0086): app_settings.verify_chat_mode (same_chat | different_chat
// default) and tasks.verify_chat_mode (empty inherits settings).
//
// Rev 21: project_context_items.tag replaces kind for UI grouping; migrate
// legacy role kinds (note/decision/constraint/handoff) to tag=General.
//
// Rev 22: remove project memory — drop project_context_items/edges,
// task_context_snapshots, tasks.project_context_item_ids, projects.context_summary.
//
// Rev 23: app_settings.agent_mcp_enabled — Hamix agent MCP tool-only report
// submit (default true; false is emergency kill-switch to legacy freeform Write).
//
// Rev 24 (ADR-0090): drop app_settings.verify_max_retries,
// app_settings.verify_chat_mode, and tasks.verify_chat_mode — command-only
// verify is one-shot; operators no longer configure retry loops or chat mode.
//
// Rev 25: app_settings.agent_task_parallelism — max parallel tasks across
// different worktrees (replaces HAMIX_AGENT_WORKER_CONCURRENCY env).
const SchemaRevision = 25

const schemaMetaRowID = 1

// SchemaDriftRemediation tells operators how to apply pending schema changes.
const SchemaDriftRemediation = "run dbcheck -migrate or scripts/migrate.*"

// SchemaDriftStatus classifies code vs database schema revision alignment.
type SchemaDriftStatus string

const (
	SchemaDriftOK        SchemaDriftStatus = "ok"
	SchemaDriftPending   SchemaDriftStatus = "pending"
	SchemaDriftDowngrade SchemaDriftStatus = "downgrade"
)

// SchemaMeta records the schema revision last applied by postgres.Migrate.
type SchemaMeta struct {
	ID        int       `gorm:"primaryKey"`
	Revision  int       `gorm:"not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// SchemaDriftReport is the result of comparing code SchemaRevision to schema_meta.
type SchemaDriftReport struct {
	Status       SchemaDriftStatus
	CodeRevision int
	DBRevision   int
}

// OperatorMessage is plain-language guidance for stderr and fatal startup errors.
//
//funclogmeasure:skip category=hot-path reason="Pure string formatter; drift is traced at taskapi startup."
func (r SchemaDriftReport) OperatorMessage() string {
	switch r.Status {
	case SchemaDriftPending:
		if r.DBRevision == 0 {
			return "Database schema has not been migrated yet. Apply schema migrate before starting taskapi."
		}
		return "Database schema is out of date for this build. Apply schema migrate before starting taskapi."
	case SchemaDriftDowngrade:
		return "This taskapi build is older than the database schema. Deploy a matching release before starting taskapi."
	default:
		return ""
	}
}

// Remediation returns operator guidance when drift fails readiness.
//
//funclogmeasure:skip category=hot-path reason="Pure constant accessor; drift is traced at taskapi startup and GET /health/ready."
func (r SchemaDriftReport) Remediation() string {
	switch r.Status {
	case SchemaDriftDowngrade:
		return "deploy a taskapi build that matches the database schema"
	default:
		return SchemaDriftRemediation
	}
}

// RemediationCLI returns a one-line shell command for stderr when migrate applies.
//
//funclogmeasure:skip category=hot-path reason="Pure string formatter; drift is traced at taskapi startup."
func (r SchemaDriftReport) RemediationCLI() string {
	switch r.Status {
	case SchemaDriftDowngrade:
		return "Deploy a taskapi build that matches the database schema."
	default:
		return "Run: .\\scripts\\migrate.ps1   or   go run ./cmd/dbcheck -migrate"
	}
}

// FailsReadiness reports whether GET /health/ready should return 503 for schema.
//
//funclogmeasure:skip category=hot-path reason="Pure predicate; readiness handler traces the HTTP boundary."
func (r SchemaDriftReport) FailsReadiness() bool {
	return r.Status == SchemaDriftPending || r.Status == SchemaDriftDowngrade
}

// DefaultDevStartupGrace is added to DefaultMigrateTimeout for dev script port waits
// when migrate runs before taskapi (scripts/dev.* --migrate sugar).
const DefaultDevStartupGrace = 30 * time.Second

// DefaultDevReadinessTimeout returns how long dev scripts wait for taskapi to listen.
//
//funclogmeasure:skip category=hot-path reason="Pure duration helper for devconfig CLI; no I/O boundary."
func DefaultDevReadinessTimeout() time.Duration {
	return DefaultMigrateTimeout + DefaultDevStartupGrace
}

// CheckSchemaDrift compares SchemaRevision to the revision stored in schema_meta.
func CheckSchemaDrift(ctx context.Context, db *gorm.DB) (SchemaDriftReport, error) {
	slog.Debug("trace", "operation", "postgres.CheckSchemaDrift")
	report := SchemaDriftReport{
		Status:       SchemaDriftPending,
		CodeRevision: SchemaRevision,
		DBRevision:   0,
	}
	if err := db.WithContext(ctx).AutoMigrate(&SchemaMeta{}); err != nil {
		return report, fmt.Errorf("automigrate schema_meta: %w", err)
	}
	var meta SchemaMeta
	err := db.WithContext(ctx).First(&meta, schemaMetaRowID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return report, nil
	}
	if err != nil {
		return report, fmt.Errorf("read schema_meta: %w", err)
	}
	report.DBRevision = meta.Revision
	switch {
	case meta.Revision < SchemaRevision:
		report.Status = SchemaDriftPending
	case meta.Revision > SchemaRevision:
		report.Status = SchemaDriftDowngrade
	default:
		report.Status = SchemaDriftOK
	}
	return report, nil
}

// RecordSchemaRevision upserts schema_meta after a successful Migrate.
func RecordSchemaRevision(ctx context.Context, db *gorm.DB, at time.Time) error {
	slog.Debug("trace", "operation", "postgres.RecordSchemaRevision", "revision", SchemaRevision)
	if err := db.WithContext(ctx).AutoMigrate(&SchemaMeta{}); err != nil {
		return fmt.Errorf("automigrate schema_meta: %w", err)
	}
	row := SchemaMeta{
		ID:        schemaMetaRowID,
		Revision:  SchemaRevision,
		AppliedAt: at.UTC(),
	}
	if err := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"revision", "applied_at"}),
	}).Create(&row).Error; err != nil {
		return fmt.Errorf("upsert schema_meta: %w", err)
	}
	return nil
}
