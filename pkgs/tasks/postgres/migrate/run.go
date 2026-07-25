package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// Deps supplies orchestration hooks that must stay in package postgres to avoid
// an import cycle (SchemaMeta lives next to RecordSchemaRevision).
type Deps struct {
	AutoMigrateSchemaMeta func(ctx context.Context, db *gorm.DB) error
}

// Run executes the ordered one-shot migrate steps and AutoMigrate for store
// models. Callers record schema revision after Run returns.
func Run(ctx context.Context, db *gorm.DB, deps Deps) error {
	slog.Debug("trace", "operation", "migrate.Run")
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		if err := db.WithContext(ctx).Exec(`ALTER TABLE project_context_items DROP CONSTRAINT IF EXISTS chk_project_context_kind`).Error; err != nil {
			return fmt.Errorf("drop project context kind constraint: %w", err)
		}
	}
	if err := migrateExpandFixedWorktreeBranch(ctx, db); err != nil {
		return fmt.Errorf("expand fixed worktree branch: %w", err)
	}
	if err := autoMigrateStoreModels(db.WithContext(ctx)); err != nil {
		return fmt.Errorf("automigrate store models: %w", err)
	}
	if deps.AutoMigrateSchemaMeta != nil {
		if err := deps.AutoMigrateSchemaMeta(ctx, db); err != nil {
			return fmt.Errorf("automigrate schema meta: %w", err)
		}
	}
	if err := migrateRemoveSubtasks(ctx, db); err != nil {
		return fmt.Errorf("migrate remove subtasks: %w", err)
	}
	if err := migrateRemoveTaskType(ctx, db); err != nil {
		return fmt.Errorf("migrate remove task type: %w", err)
	}
	if err := migrateRemoveDraftEvaluations(ctx, db); err != nil {
		return fmt.Errorf("migrate remove draft evaluations: %w", err)
	}
	if err := backfillLegacyChecklistCompletions(ctx, db); err != nil {
		return fmt.Errorf("backfill checklist completions: %w", err)
	}
	if err := migrateChecklistCheckToText(ctx, db); err != nil {
		return fmt.Errorf("migrate checklist check column: %w", err)
	}
	if err := migrateDropPromptAutomations(ctx, db); err != nil {
		return fmt.Errorf("migrate drop prompt automations: %w", err)
	}
	if err := dropLegacyGoalStepTables(ctx, db); err != nil {
		return fmt.Errorf("drop legacy goal/step tables: %w", err)
	}
	if err := migrateRepoRootToGitRepository(ctx, db); err != nil {
		return fmt.Errorf("migrate repo_root to git repository: %w", err)
	}
	if err := migrateDropRepoRootColumn(ctx, db); err != nil {
		return fmt.Errorf("drop app_settings.repo_root: %w", err)
	}
	if err := migrateDropStreamIdleStuckColumn(ctx, db); err != nil {
		return fmt.Errorf("drop app_settings.stream_idle_stuck_seconds: %w", err)
	}
	if err := migrateSeedWorktreeBranchTree(ctx, db); err != nil {
		return fmt.Errorf("seed worktree-branch tree: %w", err)
	}
	if err := migrateContractGitTree(ctx, db); err != nil {
		return fmt.Errorf("contract git tree: %w", err)
	}
	if err := migrateFixedWorktreeBranch(ctx, db); err != nil {
		return fmt.Errorf("fixed worktree branch: %w", err)
	}
	if err := migrateGitCommonDir(ctx, db); err != nil {
		return fmt.Errorf("git common dir: %w", err)
	}
	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		return fmt.Errorf("repo default projects: %w", err)
	}
	if err := migrateComposePayloadWorktree(ctx, db); err != nil {
		return fmt.Errorf("compose payload worktree: %w", err)
	}
	if err := migrateOrphanRepoProjects(ctx, db); err != nil {
		return fmt.Errorf("orphan repo projects: %w", err)
	}
	if err := migrateUnlinkedProjects(ctx, db); err != nil {
		return fmt.Errorf("unlinked projects: %w", err)
	}
	if err := migrateBackfillTaskNumbers(ctx, db); err != nil {
		return fmt.Errorf("backfill task numbers: %w", err)
	}
	return nil
}
