# ADR-0095: Desktop Wails host (thin shell)

**Date:** 2026-08-01  
**Status:** Accepted  
**Deciders:** Hamix maintainers  

## Context

Hamix is a Go + React product with a browser/dev path (`cmd/taskapi` + Vite). Users also need a downloadable desktop app. The stack must stay hireable: one product contract, no Wails-colored domain packages, and a clear answer to “where does desktop end?”

Postgres remains the store. The connection string cannot live in DB-backed `app_settings` (chicken-and-egg). Download users should not be required to set shell env vars.

## Decision

1. **Host:** Wails v2 AssetServer embeds `web/dist` and mounts the shared API `http.Handler` from `internal/taskapiruntime` for non-asset requests (same-origin REST + SSE).
2. **Domain IPC:** Do **not** expose tasks/projects/git/etc. as Wails `Bind` methods. The SPA keeps relative `fetch` and `EventSource`.
3. **Bind allowlist:** window lifecycle + desktop database config only (`Get` / `Save` / `Test` connection URL, optional restart).
4. **DSN storage:** `{UserConfigDir}/hamix/desktop.json` field `database_url` via `internal/desktopconfig`. Precedence: `DATABASE_URL` env → file → setup UI.
5. **One SPA:** same `web/` for browser and desktop; single bridge module for WebView IPC.
6. **Wails import boundary:** only `cmd/hamix-desktop` (+ optional `internal/desktop` glue). Never import Wails from `pkgs/*`.
7. **Browser path stays first-class:** contributors keep `taskapi` + Vite; desktop does not replace local web development.
8. **Wails build tags:** Host binaries must be built with `-tags desktop,production` (or via `wails dev` / `wails build`). Untagged `go build ./cmd/hamix-desktop` selects Wails’ stub and is invalid. The SSOT for flags is `scripts/build-desktop.ps1` / `scripts/build-desktop.sh`; `dev-desktop.*` only orchestrates and must call those scripts.

## Consequences

### Positive

- One HTTP+SSE contract and one SPA for all hosts
- New hires can own desktop work without learning a second API
- Config package is unit-testable without a WebView

### Negative / trade-offs

- Changing the DSN requires app restart (no live reconnect in v1)
- Users must bring their own Postgres (local, Docker, or cloud)
- Production WebView builds are typically per-OS (not a single Windows cross-compile)
- Contributors must not invent ad-hoc `go build` for this binary; CI asserts `build-desktop.*` / `dev-desktop.*` keep the tag contract

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Electron / Tauri | Heavier or less natural with an existing Go API stack |
| Rewrite domain APIs as Wails bindings | Dual clients, dual contracts, hire nightmare |
| Ship SQLite for desktop | Large concurrency/parity surface; not needed if URL is user-configured |
| Shell-only / `.env` for download users | Poor desktop UX |
| Store DSN in Postgres `app_settings` | Unreachable before connect |

## Forbid list

- Wails bindings that mirror `/tasks`, `/events`, …
- `web-desktop/` or duplicated API clients
- `DATABASE_URL` fields on Postgres-backed settings
- Desktop-only forks of middleware, SSE, or composition
- Build tags that hide half of `pkgs/` behind `//go:build desktop`
- Documenting or scripting untagged `go build ./cmd/hamix-desktop` (use `scripts/build-desktop.*`)

## See also

- [docs/configuration.md](../configuration.md) (desktop DSN)
- [internal/desktopconfig](../../internal/desktopconfig)
- [internal/taskapiruntime](../../internal/taskapiruntime)
- [ADR-0081](./ADR-0081-hamix-managed-worktrees.md) (same `{UserConfigDir}/hamix` root)
