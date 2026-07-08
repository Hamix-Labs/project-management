package tasktestdb

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// postgresTestDSN returns HAMIX_TEST_POSTGRES_URL, then DATABASE_URL, or "".
//
//funclogmeasure:skip category=hot-path reason="Test-only env helper; OpenPostgres emits the operation trace."
func postgresTestDSN() string {
	if v := strings.TrimSpace(os.Getenv("HAMIX_TEST_POSTGRES_URL")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("DATABASE_URL"))
}

// OpenPostgres returns a GORM DB connected to Postgres with schema migrated.
// Skips the test when no DSN is configured so default CI (SQLite-only) stays green.
func OpenPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := postgresTestDSN()
	if dsn == "" {
		t.Skip("HAMIX_TEST_POSTGRES_URL or DATABASE_URL not set; skipping Postgres store test")
	}
	slog.Debug("trace", "operation", "tasktestdb.OpenPostgres")
	db, err := postgres.Open(dsn, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("postgres sql db: %v", err)
	}
	if err := sqlDB.PingContext(context.Background()); err != nil {
		_ = sqlDB.Close()
		t.Skipf("postgres ping failed: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close postgres: %v", err)
		}
	})
	if err := postgres.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}
	return db
}
