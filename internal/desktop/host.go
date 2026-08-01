// Package desktop is thin glue for cmd/hamix-desktop: AssetServer handler
// chaining (API + SPA fallback) and runtime lifecycle. No Wails imports —
// the cmd package owns the WebView host.
package desktop

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/AlexsanderHamir/Hamix/internal/desktopconfig"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiruntime"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
)

const cmdName = "hamix-desktop"

// Host owns optional API runtime and serves AssetServer fallback requests.
type Host struct {
	mu     sync.RWMutex
	rt     *taskapiruntime.Runtime
	paths  desktopconfig.Paths
	assets fs.FS // root containing index.html (already sub'd if needed)
	cancel context.CancelFunc
	appCtx context.Context
}

// NewHost constructs a Host. assets must contain index.html at its root
// (typically fs.Sub of the embed after FindPath). paths may be Default().
func NewHost(paths desktopconfig.Paths, assets fs.FS) *Host {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.NewHost")
	return &Host{paths: paths, assets: assets}
}

// StartRuntime resolves DSN and starts taskapiruntime when configured.
// Returns ErrNotConfigured when no DSN (setup UI mode).
func (h *Host) StartRuntime(parent context.Context) error {
	dsn, src, err := h.paths.RequireDSN()
	if err != nil {
		return err
	}
	slog.Info("desktop database resolved", "cmd", cmdName, "operation", "desktop.dsn",
		"source", string(src))

	ctx, cancel := context.WithCancel(parent)
	migrate := taskapiconfig.MigrateEnabled(false)
	rt, err := taskapiruntime.Start(ctx, taskapiruntime.Options{
		DatabaseURL: dsn,
		Migrate:     migrate,
		CmdName:     cmdName,
	})
	if err != nil {
		cancel()
		return err
	}
	if mode := taskapiconfig.GitReconcileOnStartupMode(); mode != "" {
		rt.ReconcileGitOnStartup(ctx)
	}

	h.mu.Lock()
	h.cancel = cancel
	h.appCtx = ctx
	h.rt = rt
	h.mu.Unlock()
	return nil
}

// Close stops the API runtime if running.
func (h *Host) Close() error {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.Close")
	h.mu.Lock()
	rt := h.rt
	cancel := h.cancel
	h.rt = nil
	h.cancel = nil
	h.appCtx = nil
	h.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if rt != nil {
		return rt.Close()
	}
	return nil
}

// Handler returns the AssetServer fallback handler (API + SPA HTML).
func (h *Host) Handler() http.Handler {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.Handler")
	return http.HandlerFunc(h.serve)
}

func (h *Host) serve(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.serve", "path", r.URL.Path, "method", r.Method)
	if isAPIPath(r.URL.Path) {
		h.serveAPI(w, r)
		return
	}
	if r.Method == http.MethodGet && wantsHTML(r) {
		if serveIndexHTML(w, h.assets) {
			return
		}
	}
	if isAPIPath(r.URL.Path) || r.Method != http.MethodGet {
		h.serveAPI(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *Host) serveAPI(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.serveAPI", "path", r.URL.Path)
	h.mu.RLock()
	rt := h.rt
	h.mu.RUnlock()
	if rt == nil || rt.Handler == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		// Prefer Quit on StartRuntime failure; this path is defense in depth.
		// Never claim "not configured" when a DSN is already resolved.
		msg := `{"error":"database not configured"}`
		if _, src, err := h.paths.Resolve(); err == nil && src != desktopconfig.SourceNone {
			msg = `{"error":"api runtime unavailable"}`
		}
		_, _ = w.Write([]byte(msg))
		return
	}
	rt.Handler.ServeHTTP(w, r)
}

// Configured reports whether a DSN is available (env or file).
func (h *Host) Configured() bool {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.Configured")
	_, src, err := h.paths.Resolve()
	return err == nil && src != desktopconfig.SourceNone
}

// GetDatabaseURL returns the effective or file-stored URL for the UI
// (env overrides are reflected; empty when unconfigured).
func (h *Host) GetDatabaseURL() (url string, source string, err error) {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.GetDatabaseURL")
	dsn, src, err := h.paths.Resolve()
	if err != nil {
		return "", "", err
	}
	return dsn, string(src), nil
}

// SaveDatabaseURL writes desktop.json (does not apply until restart).
func (h *Host) SaveDatabaseURL(url string) error {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.SaveDatabaseURL")
	return h.paths.Save(desktopconfig.File{DatabaseURL: strings.TrimSpace(url)})
}

// TestDatabaseURL opens Postgres with the given DSN and pings.
func (h *Host) TestDatabaseURL(ctx context.Context, url string) error {
	slog.Debug("trace", "cmd", cmdName, "operation", "desktop.TestDatabaseURL")
	url = strings.TrimSpace(url)
	if url == "" {
		return desktopconfig.ErrNotConfigured
	}
	db, err := postgres.Open(url, postgres.ConfigWithSlogLogger(slog.Default()))
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()
	pingCtx, cancel := context.WithTimeout(ctx, postgres.DefaultPingTimeout)
	defer cancel()
	return sqlDB.PingContext(pingCtx)
}

// Paths exposes config paths for tests.
//
//funclogmeasure:skip category=hot-path reason="Pure accessor; no I/O boundary."
func (h *Host) Paths() desktopconfig.Paths { return h.paths }
