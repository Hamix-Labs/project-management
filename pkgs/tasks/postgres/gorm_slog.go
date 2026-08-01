package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const defaultSlowQueryMS = 200

// slowQueryThresholdForGORM reads HAMIX_GORM_SLOW_QUERY_MS (milliseconds).
// Empty: defaultSlowQueryMS. "0": disable the slow-SQL Warn branch (non-slow SQL stays at Debug).
// Invalid or negative: defaultSlowQueryMS.
func slowQueryThresholdForGORM() time.Duration {
	slog.Debug("trace", "operation", "postgres.slowQueryThresholdForGORM")
	s := strings.TrimSpace(os.Getenv("HAMIX_GORM_SLOW_QUERY_MS"))
	if s == "" {
		return time.Duration(defaultSlowQueryMS) * time.Millisecond
	}
	ms, err := strconv.Atoi(s)
	if err != nil || ms < 0 {
		return time.Duration(defaultSlowQueryMS) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}

// SlowQueryThresholdMS returns the effective GORM slow-SQL threshold in milliseconds
// (HAMIX_GORM_SLOW_QUERY_MS; default 200; 0 means the slow-SQL warn branch is off).
func SlowQueryThresholdMS() int {
	slog.Debug("trace", "operation", "postgres.SlowQueryThresholdMS")
	d := slowQueryThresholdForGORM()
	if d <= 0 {
		return 0
	}
	return int(d / time.Millisecond)
}

// ConfigWithSlogLogger returns a GORM config that records each SQL round-trip through lg
// (typically slog.Default() after taskapi attaches the JSON log handler).
// Non-slow SQL logs at slog Debug so default Info stays sparse; statements slower than
// HAMIX_GORM_SLOW_QUERY_MS (default 200ms; 0 disables the Warn branch) log at Warn.
// ParameterizedQueries keeps bound values out of log lines.
//
// GORM's stock NewSlogLogger has no Debug level and emits non-slow SQL at Info; we use a
// thin adapter so Info remains representative of lifecycle events, not every query.
func ConfigWithSlogLogger(lg *slog.Logger) *gorm.Config {
	slog.Debug("trace", "operation", "postgres.ConfigWithSlogLogger")
	if lg == nil {
		return nil
	}
	return GORMConfigDefaults(&gorm.Config{
		Logger: newSlogGORMLogger(lg, gormlogger.Config{
			// Info enables SQL Trace in GORM's level gate; emission level is Debug (see Trace).
			LogLevel:                  gormlogger.Info,
			SlowThreshold:             slowQueryThresholdForGORM(),
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		}),
	})
}

// slogGORMLogger mirrors gormlogger.NewSlogLogger but logs non-slow SQL at slog.LevelDebug.
type slogGORMLogger struct {
	logger                    *slog.Logger
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	parameterized             bool
	ignoreRecordNotFoundError bool
}

func newSlogGORMLogger(logger *slog.Logger, config gormlogger.Config) gormlogger.Interface {
	slog.Debug("trace", "operation", "postgres.newSlogGORMLogger")
	return &slogGORMLogger{
		logger:                    logger,
		logLevel:                  config.LogLevel,
		slowThreshold:             config.SlowThreshold,
		parameterized:             config.ParameterizedQueries,
		ignoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
	}
}

//funclogmeasure:skip category=hot-path reason="GORM logger.Interface; Trace owns SQL emission."
func (l *slogGORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	n := *l
	n.logLevel = level
	return &n
}

//funclogmeasure:skip category=hot-path reason="GORM logger.Interface Info; rare vs Trace."
func (l *slogGORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.log(ctx, slog.LevelInfo, msg, slog.Any("data", data))
	}
}

//funclogmeasure:skip category=hot-path reason="GORM logger.Interface Warn; rare vs Trace."
func (l *slogGORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.log(ctx, slog.LevelWarn, msg, slog.Any("data", data))
	}
}

//funclogmeasure:skip category=hot-path reason="GORM logger.Interface Error; rare vs Trace."
func (l *slogGORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.log(ctx, slog.LevelError, msg, slog.Any("data", data))
	}
}

//funclogmeasure:skip category=hot-path reason="Per-query SQL boundary; emits slog Debug/Warn/Error."
func (l *slogGORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []slog.Attr{
		slog.String("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)),
		slog.String("sql", sql),
	}
	if rows != -1 {
		fields = append(fields, slog.Int64("rows", rows))
	}

	switch {
	case err != nil && (!l.ignoreRecordNotFoundError || !errors.Is(err, gormlogger.ErrRecordNotFound)):
		fields = append(fields, slog.String("error", err.Error()))
		l.log(ctx, slog.LevelError, "SQL executed", slog.Attr{
			Key:   "trace",
			Value: slog.GroupValue(fields...),
		})
	case l.slowThreshold != 0 && elapsed > l.slowThreshold:
		l.log(ctx, slog.LevelWarn, "SQL executed", slog.Attr{
			Key:   "trace",
			Value: slog.GroupValue(fields...),
		})
	case l.logLevel >= gormlogger.Info:
		l.log(ctx, slog.LevelDebug, "SQL executed", slog.Attr{
			Key:   "trace",
			Value: slog.GroupValue(fields...),
		})
	}
}

//funclogmeasure:skip category=hot-path reason="Pure params filter for GORM logger.Interface."
func (l *slogGORMLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.parameterized {
		return sql, nil
	}
	return sql, params
}

//funclogmeasure:skip category=hot-path reason="Internal slog emit helper for GORM Trace/Info/Warn/Error."
func (l *slogGORMLogger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	if !l.logger.Enabled(ctx, level) {
		return
	}
	r := slog.NewRecord(time.Now(), level, msg, utils.CallerFrame().PC)
	r.Add(args...)
	_ = l.logger.Handler().Handle(ctx, r)
}
