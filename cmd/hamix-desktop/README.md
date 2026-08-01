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
go build -o hamix-desktop-dev.exe ./cmd/hamix-desktop
.\hamix-desktop-dev.exe
```

Install the Wails CLI once for `-Live`: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build

```powershell
wails build
```

Produces `build/bin/hamix-desktop` (platform-specific).

## Config

DSN precedence: `DATABASE_URL` env → `{UserConfigDir}/hamix/desktop.json` → setup UI. See [docs/configuration.md](../../docs/configuration.md).
