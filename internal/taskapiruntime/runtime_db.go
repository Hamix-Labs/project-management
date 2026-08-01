package taskapiruntime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
	"gorm.io/gorm"
)

type dbStartupResult struct {
	db          *gorm.DB
	schemaDrift postgres.SchemaDriftReport
}

func migrateDBAndRegisterMetrics(db *gorm.DB, cmd string) error {
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), postgres.DefaultMigrateTimeout)
	defer migrateCancel()
	if err := postgres.Migrate(migrateCtx, db); err != nil {
		return fmt.Errorf("migrate (timeout %ds, deadline_exceeded=%t): %w",
			int(postgres.DefaultMigrateTimeout/time.Second),
			errors.Is(err, context.DeadlineExceeded),
			err)
	}
	if err := composition.BackfillCriteriaSatisfiedAt(migrateCtx, db); err != nil {
		return fmt.Errorf("backfill criteria_satisfied_at: %w", err)
	}
	slog.Info("migrate ok", "cmd", cmd, "operation", "taskapi.migrate",
		"timeout_sec", int(postgres.DefaultMigrateTimeout/time.Second),
		"schema_revision", postgres.SchemaRevision)
	return nil
}

func registerDBMetrics(db *gorm.DB, cmd string) {
	postgres.LogStartupDBConfig(slog.Default(), cmd, db)
	taskapi.RegisterSQLDBPoolCollector(db)
}

func emitSchemaDriftAlerts(report postgres.SchemaDriftReport, cmd string) {
	if report.Status != postgres.SchemaDriftPending && report.Status != postgres.SchemaDriftDowngrade {
		return
	}
	fmt.Fprintf(os.Stderr,
		"%s: %s\n"+
			"         %s\n",
		cmd, report.OperatorMessage(), report.RemediationCLI())
	slog.Error("schema migrate required", "cmd", cmd, "operation", "taskapi.schema_drift",
		"status", string(report.Status),
		"code_revision", report.CodeRevision,
		"db_revision", report.DBRevision,
		"message", report.OperatorMessage(),
		"remediation", report.Remediation())
}

func openDatabase(databaseURL string, migrateEnabled bool, cmd string) (dbStartupResult, error) {
	var out dbStartupResult

	db, err := postgres.Open(
		databaseURL,
		postgres.ConfigWithSlogLogger(slog.Default()),
	)
	if err != nil {
		return out, err
	}
	out.db = db

	checkCtx, checkCancel := context.WithTimeout(context.Background(), postgres.DefaultPingTimeout)
	defer checkCancel()
	drift, err := postgres.CheckSchemaDrift(checkCtx, db)
	if err != nil {
		return out, fmt.Errorf("schema drift check: %w", err)
	}

	if migrateEnabled {
		if err := migrateDBAndRegisterMetrics(db, cmd); err != nil {
			return out, err
		}
		drift, err = postgres.CheckSchemaDrift(checkCtx, db)
		if err != nil {
			return out, fmt.Errorf("schema drift check after migrate: %w", err)
		}
		if drift.FailsReadiness() {
			emitSchemaDriftAlerts(drift, cmd)
			_ = closeSQLDBOrLog(db, cmd)
			out.db = nil
			return out, fmt.Errorf("%s", drift.OperatorMessage())
		}
	} else {
		slog.Info("migrate skipped", "cmd", cmd, "operation", "taskapi.migrate",
			"reason", "not_requested",
			"hint", "run scripts/migrate.* or pass -migrate / set HAMIX_MIGRATE=1")
		if drift.FailsReadiness() {
			emitSchemaDriftAlerts(drift, cmd)
			_ = closeSQLDBOrLog(db, cmd)
			out.db = nil
			return out, fmt.Errorf("%s", drift.OperatorMessage())
		}
	}

	out.schemaDrift = drift
	registerDBMetrics(db, cmd)
	return out, nil
}

// closeSQLDBOrLog closes the GORM-owned *sql.DB pool and logs the outcome.
// Callers take the returned bool as the success signal and never re-log.
func closeSQLDBOrLog(db *gorm.DB, cmd string) (dbClosed bool) {
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("database close skipped", "cmd", cmd, "operation", "taskapi.db_close", "err", err)
		return false
	}
	if err := sqlDB.Close(); err != nil {
		slog.Error("database close", "cmd", cmd, "operation", "taskapi.db_close", "err", err)
		return false
	}
	slog.Info("database pool closed", "cmd", cmd, "operation", "taskapi.shutdown", "phase", "db_done")
	return true
}
