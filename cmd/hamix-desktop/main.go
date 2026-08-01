package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/AlexsanderHamir/Hamix/internal/desktop"
	"github.com/AlexsanderHamir/Hamix/internal/desktopconfig"
	"github.com/AlexsanderHamir/Hamix/internal/envload"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var embeddedAssets embed.FS

const cmdName = "hamix-desktop"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmdName, err)
		os.Exit(1)
	}
}

func run() error {
	slog.Debug("trace", "cmd", cmdName, "operation", "hamix-desktop.run")
	// Optional repo-root .env for local desktop dev (DATABASE_URL override).
	if _, err := envload.OverloadDotenvIfPresent(""); err != nil {
		return fmt.Errorf("preload .env: %w", err)
	}

	assetFS, err := fs.Sub(embeddedAssets, "frontend/dist")
	if err != nil {
		return fmt.Errorf("embed assets: %w", err)
	}

	host := desktop.NewHost(desktopconfig.Default(), assetFS)
	app := NewApp(host)

	return wails.Run(&options.App{
		Title:  "Hamix",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets:  embeddedAssets,
			Handler: host.Handler(),
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
}

// App is the Wails bind surface. Allowlist: database config + quit only.
// Do not add domain APIs here (ADR-0095).
type App struct {
	ctx  context.Context
	host *desktop.Host
}

// NewApp wraps the desktop host for Wails.
func NewApp(host *desktop.Host) *App {
	slog.Debug("trace", "cmd", cmdName, "operation", "hamix-desktop.NewApp")
	return &App{host: host}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.host.StartRuntime(ctx); err != nil {
		if errors.Is(err, desktopconfig.ErrNotConfigured) {
			slog.Info("desktop setup required", "cmd", cmdName, "operation", "desktop.startup")
			return
		}
		slog.Error("desktop runtime start failed", "cmd", cmdName, "operation", "desktop.startup", "err", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if err := a.host.Close(); err != nil {
		slog.Error("desktop runtime close", "cmd", cmdName, "operation", "desktop.shutdown", "err", err)
	}
}

// DatabaseConfigDTO is returned to the SPA setup/settings UI.
type DatabaseConfigDTO struct {
	URL        string `json:"url"`
	Source     string `json:"source"`
	Configured bool   `json:"configured"`
	NeedsSetup bool   `json:"needsSetup"`
}

// GetDatabaseConfig returns the resolved DSN metadata for the UI.
func (a *App) GetDatabaseConfig() (DatabaseConfigDTO, error) {
	slog.Debug("trace", "cmd", cmdName, "operation", "hamix-desktop.GetDatabaseConfig")
	url, source, err := a.host.GetDatabaseURL()
	if err != nil {
		return DatabaseConfigDTO{}, err
	}
	configured := source != string(desktopconfig.SourceNone) && url != ""
	return DatabaseConfigDTO{
		URL:        url,
		Source:     source,
		Configured: configured,
		NeedsSetup: !configured,
	}, nil
}

// SaveDatabaseConfig persists database_url to desktop.json.
func (a *App) SaveDatabaseConfig(url string) error {
	slog.Debug("trace", "cmd", cmdName, "operation", "hamix-desktop.SaveDatabaseConfig")
	return a.host.SaveDatabaseURL(url)
}

// TestDatabaseConnection pings Postgres with the given URL.
func (a *App) TestDatabaseConnection(url string) error {
	slog.Debug("trace", "cmd", cmdName, "operation", "hamix-desktop.TestDatabaseConnection")
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.host.TestDatabaseURL(ctx, url)
}

// QuitApp closes the window (after save, SPA asks the user to reopen or calls this).
func (a *App) QuitApp() {
	slog.Debug("trace", "cmd", cmdName, "operation", "hamix-desktop.QuitApp")
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}
