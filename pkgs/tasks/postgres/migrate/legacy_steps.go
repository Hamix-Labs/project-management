package migrate

import (
	"context"
	"fmt"
	"log/slog"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	"gorm.io/gorm"
)

// dropLegacyGoalStepTables removes project_goals and project_steps after the
// flat task hierarchy migration. Idempotent - safe on fresh and upgraded DBs.
func dropLegacyGoalStepTables(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "migrate.dropLegacyGoalStepTables")
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS project_steps CASCADE`).Error; err != nil {
			return fmt.Errorf("drop project_steps: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS project_goals CASCADE`).Error; err != nil {
			return fmt.Errorf("drop project_goals: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP COLUMN IF EXISTS project_step_id`).Error; err != nil {
			return fmt.Errorf("drop tasks.project_step_id: %w", err)
		}
		return nil
	}
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS project_steps`).Error; err != nil {
		return fmt.Errorf("drop project_steps: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS project_goals`).Error; err != nil {
		return fmt.Errorf("drop project_goals: %w", err)
	}
	return nil
}

// backfillLegacyChecklistCompletions marks pre-V1.1 completion rows so
// ValidateCanMarkDoneInTx continues to accept them after evidence columns ship.
func backfillLegacyChecklistCompletions(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "migrate.backfillLegacyChecklistCompletions")
	res := db.WithContext(ctx).Exec(`
UPDATE task_checklist_completions
   SET verified_by = ?
 WHERE (verified_by IS NULL OR verified_by = '')
   AND (evidence IS NULL OR evidence = '')`,
		string(checklistdomain.VerifierLegacy),
	)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// migrateChecklistCheckToText merges legacy shell-check commands into criterion
// text, then drops the check column and app_settings.check_command_timeout_seconds.
// Postgres only; SQLite test DBs rely on AutoMigrate after the domain field removal.
// Idempotent: skips the merge when the column was already dropped on a prior boot.
func migrateChecklistCheckToText(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "migrate.migrateChecklistCheckToText")
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}
	hasCheck, err := postgresTableHasColumn(ctx, db, "task_checklist_items", "check")
	if err != nil {
		return fmt.Errorf("probe task_checklist_items.check: %w", err)
	}
	if hasCheck {
		if err := db.WithContext(ctx).Exec(`
UPDATE task_checklist_items
   SET text = text || ' (verification: ' || trim("check") || ')'
 WHERE trim("check") != ''
   AND text NOT LIKE '%(verification:%'`).Error; err != nil {
			return fmt.Errorf("merge checklist check into text: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE task_checklist_items DROP COLUMN IF EXISTS "check"`).Error; err != nil {
			return fmt.Errorf("drop task_checklist_items.check: %w", err)
		}
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS chk_app_settings_check_timeout`).Error; err != nil {
		return fmt.Errorf("drop app_settings check timeout constraint: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN IF EXISTS check_command_timeout_seconds`).Error; err != nil {
		return fmt.Errorf("drop app_settings.check_command_timeout_seconds: %w", err)
	}
	return nil
}

// migrateDropPromptAutomations removes the prompt-automations feature schema.
// Postgres only; SQLite test DBs rely on AutoMigrate after domain field removal.
// Idempotent: safe on fresh and upgraded DBs.
func migrateDropPromptAutomations(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "migrate.migrateDropPromptAutomations")
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP COLUMN IF EXISTS automation_selections`).Error; err != nil {
		return fmt.Errorf("drop tasks.automation_selections: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS automations`).Error; err != nil {
		return fmt.Errorf("drop automations table: %w", err)
	}
	return nil
}
