# hamix-desktop

Wails v2 host for the Hamix SPA + shared API runtime ([ADR-0095](../../docs/adr/ADR-0095-desktop-wails-host.md)).

## Bind allowlist

Only these methods are exposed to the WebView (`window.go.main.App`):

- `GetDatabaseConfig`
- `SaveDatabaseConfig`
- `TestDatabaseConnection`
- `QuitApp`

Do **not** add domain APIs (tasks, projects, git, …). The SPA uses same-origin HTTP + SSE.

## Develop

From the repo root (same `.env` / `DATABASE_URL` as browser dev):

```powershell
.\scripts\dev-desktop.ps1          # Windows — build SPA + open desktop window
.\scripts\dev-desktop.ps1 -Migrate # optional migrate sugar
.\scripts\dev-desktop.ps1 -Live    # wails dev (Vite HMR; needs Wails CLI)
```

```bash
./scripts/dev-desktop.sh
./scripts/dev-desktop.sh --migrate
./scripts/dev-desktop.sh --live
```

Manual steps (what the default script does):

```powershell
cd web
npm ci
npm run build
cd ../cmd/hamix-desktop
node ./scripts/copy-web-dist.mjs
cd ../..
.\scripts\build-desktop.ps1   # go build -tags desktop,production (required)
.\hamix-desktop-dev.exe
```

```bash
./scripts/build-desktop.sh
./hamix-desktop-dev
```

**Build tags:** Untagged `go build ./cmd/hamix-desktop` is invalid — Wails ships a stub that shows an error dialog. Always use `scripts/build-desktop.*` (`-tags desktop,production`) or `wails dev` / `wails build` (CLI sets tags). See [Wails manual builds](https://wails.io/docs/guides/manual-builds/).

Install the Wails CLI once for `-Live`: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build

Preferred for local/dev binaries (Go + Node only, no Wails CLI):

```powershell
.\scripts\build-desktop.ps1
```

Optional packaged release via Wails CLI:

```powershell
wails build
```

Produces `build/bin/hamix-desktop` (platform-specific).

## Config

DSN precedence (no global env required for local dev):

1. `DATABASE_URL` from the process env — `dev-desktop.*` loads the **repo `.env`** (working directory = checkout root)
2. Else `{UserConfigDir}/hamix/desktop.json`
3. Else in-app setup UI

Schema migrate is a **separate** step (`.\scripts\migrate.ps1`), same as browser `dev.ps1`. Desktop does not AutoMigrate on start; schema drift / DB start failure prints remediation and quits.

## Logging

- JSONL + stderr share `HAMIX_LOG_LEVEL` (default `info`) via `internal/applog`
- Per-query SQL is **Debug**; slow SQL is **Warn** — default Info shows startup/access, not every query
- Set `HAMIX_LOG_LEVEL=debug` when you want the SQL firehose

See [docs/configuration.md](../../docs/configuration.md).
