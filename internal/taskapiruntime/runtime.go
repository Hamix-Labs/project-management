// Package taskapiruntime builds the shared Hamix API stack for hosts
// (cmd/taskapi, cmd/hamix-desktop). Lives beside internal/taskapi to avoid an
// import cycle with agentworker (which imports parent taskapi for metrics).
package taskapiruntime

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/agentworker"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"gorm.io/gorm"
)

// Options configures Start. Hosts (cmd/taskapi, cmd/hamix-desktop) load env
// and resolve DatabaseURL before calling Start.
type Options struct {
	// DatabaseURL is required (Postgres DSN).
	DatabaseURL string
	// Migrate runs GORM AutoMigrate + backfill before serving when true.
	Migrate bool
	// CmdName is the slog "cmd" field; empty defaults to calltrace.LogCmd.
	CmdName string
}

// Runtime is the shared Hamix API stack: composition store, SSE hub, agent
// worker, and the REST+SSE http.Handler. Hosts wrap Handler (e.g. /metrics,
// AssetServer) and call Close on shutdown.
type Runtime struct {
	Handler     http.Handler
	SchemaDrift postgres.SchemaDriftReport
	Store       *composition.API
	Hub         *realtime.SSEHub
	AgentQueue  *agents.MemoryQueue
	AgentWorker *agentworker.Supervisor

	db         *gorm.DB
	stopAgents context.CancelFunc
	closeHTTP  func() error
	cmdName    string
}

// Start opens the database, optionally migrates, wires agents, and builds
// NewHTTPHandler. The returned Runtime.Handler is the API stack only — hosts
// must not assume a catch-all mux that owns "/".
func Start(ctx context.Context, opts Options) (*Runtime, error) {
	cmd := opts.CmdName
	if cmd == "" {
		cmd = calltrace.LogCmd
	}
	slog.Debug("trace", "cmd", cmd, "operation", "taskapiruntime.Start", "migrate", opts.Migrate)
	if opts.DatabaseURL == "" {
		return nil, fmt.Errorf("%s: DATABASE_URL is empty", cmd)
	}

	dbStartup, err := openDatabase(opts.DatabaseURL, opts.Migrate, cmd)
	if err != nil {
		if dbStartup.db != nil {
			_ = closeSQLDBOrLog(dbStartup.db, cmd)
		}
		return nil, err
	}

	rt, err := buildRuntime(ctx, dbStartup.db, dbStartup.schemaDrift, cmd)
	if err != nil {
		_ = closeSQLDBOrLog(dbStartup.db, cmd)
		return nil, err
	}
	return rt, nil
}

func buildRuntime(ctx context.Context, db *gorm.DB, drift postgres.SchemaDriftReport, cmd string) (*Runtime, error) {
	slog.Debug("trace", "cmd", cmd, "operation", "taskapi.buildRuntime")
	store := composition.NewAPI(db)
	hub := realtime.NewSSEHubWith(realtime.DefaultSSEHubOptions())
	logHandlerMiddlewareConfig(cmd)

	stopAgents, q, aw, err := startReadyTaskAgents(ctx, store, hub, cmd)
	if err != nil {
		return nil, err
	}

	api, closeHTTP := taskapi.NewHTTPHandler(store, hub, nil, aw, drift, taskapi.DefaultDraftAssistHost())
	return &Runtime{
		Handler:     api,
		SchemaDrift: drift,
		Store:       store,
		Hub:         hub,
		AgentQueue:  q,
		AgentWorker: aw,
		db:          db,
		stopAgents:  stopAgents,
		closeHTTP:   closeHTTP,
		cmdName:     cmd,
	}, nil
}

// Close drains the agent worker, stops reconcile/provisioner loops, and
// closes the database pool. Safe to call once; order matches cmd/taskapi
// historical shutdown so in-flight audit writes finish before the pool closes.
func (r *Runtime) Close() error {
	if r == nil {
		return nil
	}
	slog.Debug("trace", "cmd", r.cmdName, "operation", "taskapiruntime.Close")
	if r.AgentWorker != nil {
		r.AgentWorker.Drain()
	}
	if r.stopAgents != nil {
		r.stopAgents()
		r.stopAgents = nil
	}
	if r.closeHTTP != nil {
		if err := r.closeHTTP(); err != nil {
			slog.Warn("taskapi http close failed", "cmd", r.cmdName, "operation", "taskapiruntime.Close", "err", err)
		}
		r.closeHTTP = nil
	}
	if r.db != nil {
		ok := closeSQLDBOrLog(r.db, r.cmdName)
		r.db = nil
		if !ok {
			return fmt.Errorf("%s: database close failed", r.cmdName)
		}
	}
	return nil
}

// ReconcileGitOnStartup runs the optional git inventory reconcile when
// HAMIX_GIT_RECONCILE_ON_STARTUP (or equivalent) is configured by the host.
func (r *Runtime) ReconcileGitOnStartup(ctx context.Context) {
	if r == nil || r.Store == nil {
		return
	}
	slog.Debug("trace", "cmd", r.cmdName, "operation", "taskapiruntime.ReconcileGitOnStartup")
	r.Store.ReconcileGitRepositoriesOnStartup(ctx)
}

func logHandlerMiddlewareConfig(cmd string) {
	rlim := middleware.RateLimitPerMinuteConfigured()
	slog.Info("rate limit config", "cmd", cmd, "operation", "taskapi.rate_limit",
		"enabled", rlim > 0, "per_ip_per_min", rlim)
	slog.Info("api auth config", "cmd", cmd, "operation", "taskapi.api_auth",
		"enabled", middleware.APIAuthEnabled())

	mb := middleware.MaxRequestBodyBytesConfigured()
	slog.Info("max request body config", "cmd", cmd, "operation", "taskapi.max_body",
		"enabled", mb > 0, "max_bytes", mb)
	reqTimeout := middleware.RequestTimeout()
	reqTimeoutSec := int(reqTimeout / time.Second)
	if reqTimeout > 0 && reqTimeoutSec == 0 {
		reqTimeoutSec = 1
	}
	slog.Info("request timeout config", "cmd", cmd, "operation", "taskapi.request_timeout",
		"enabled", reqTimeout > 0, "timeout_sec", reqTimeoutSec)
	idemTTL := middleware.IdempotencyTTL()
	idemMaxEntries, idemMaxBytes := middleware.IdempotencyCacheLimits()
	idemSec := int(idemTTL / time.Second)
	if idemTTL > 0 && idemSec == 0 {
		idemSec = 1
	}
	slog.Info("idempotency config", "cmd", cmd, "operation", "taskapi.idempotency",
		"enabled", idemTTL > 0, "ttl_sec", idemSec,
		"max_entries", idemMaxEntries, "max_bytes", idemMaxBytes)
}
