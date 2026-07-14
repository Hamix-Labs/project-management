package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres/migrate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

// DefaultMigrateTimeout is the recommended upper bound for [Migrate] from operators
// (taskapi startup, dbcheck -migrate). Tests and fast local runs may use a shorter deadline or
// [context.Background] when AutoMigrate is expected to finish quickly.
const DefaultMigrateTimeout = 2 * time.Minute

// DefaultPingTimeout is the recommended upper bound for the first successful [database/sql.DB.PingContext]
// from operator CLIs (dbcheck). Long-lived servers may omit an explicit ping or use their own probe policy.
const DefaultPingTimeout = 30 * time.Second

func configureSQLPool(sqldb *sql.DB) {
	slog.Debug("trace", "operation", "postgres.configureSQLPool")
	sqldb.SetMaxOpenConns(defaultMaxOpenConns)
	sqldb.SetMaxIdleConns(defaultMaxIdleConns)
	sqldb.SetConnMaxLifetime(defaultConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(defaultConnMaxIdleTime)
}

// Open returns a GORM DB connected to PostgreSQL using the given DSN.
func Open(dsn string, cfg *gorm.Config) (*gorm.DB, error) {
	slog.Debug("trace", "operation", "postgres.Open")
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("postgres open: %w", errEmptyDSN)
	}
	dsn = ensureQueryExecModeSimpleProtocol(dsn)
	if cfg == nil {
		cfg = &gorm.Config{}
	}
	cfg = GORMConfigDefaults(cfg)
	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	sqldb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm sql db: %w", err)
	}
	configureSQLPool(sqldb)
	return db, nil
}

// Migrate runs AutoMigrate for store persistence models (works with any GORM dialector).
func Migrate(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.Migrate")
	db = db.Session(&gorm.Session{
		Logger: gormlogger.NewSlogLogger(slog.Default(), gormlogger.Config{
			LogLevel:                  gormlogger.Warn,
			SlowThreshold:             slowQueryThresholdForGORM(),
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		}),
	})
	if err := migrate.Run(ctx, db, migrate.Deps{
		AutoMigrateSchemaMeta: func(ctx context.Context, db *gorm.DB) error {
			return db.WithContext(ctx).AutoMigrate(&SchemaMeta{})
		},
	}); err != nil {
		return err
	}
	if err := RecordSchemaRevision(ctx, db, time.Now().UTC()); err != nil {
		return fmt.Errorf("record schema revision: %w", err)
	}
	return nil
}

var errEmptyDSN = errors.New("database DSN is empty")

// ensureQueryExecModeSimpleProtocol appends pgx's default_query_exec_mode when
// absent. Without this, ALTER TABLE / AutoMigrate can change the row type of
// SELECT * while pooled connections still hold cached prepared statements,
// producing PostgreSQL error 0A000 "cached plan must not change result type".
// Simple protocol avoids server-side plan caching for that failure mode.
//
// See pgx ParseConfig: default_query_exec_mode=simple_protocol.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ensureQueryExecModeSimpleProtocol(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return dsn
	}
	if strings.Contains(dsn, "default_query_exec_mode=") {
		return dsn
	}
	const param = "default_query_exec_mode=simple_protocol"
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		if strings.Contains(dsn, "?") {
			return dsn + "&" + param
		}
		return dsn + "?" + param
	}
	return dsn + " " + param
}
