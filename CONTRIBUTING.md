# Contributing to Hamix

Set up the repo, verify your change, and find the right documentation for learning or editing.

| | |
| --- | --- |
| **Applies to** | First-time setup, pull requests, finding docs |
| **Audience** | Human contributors and agents |
| **Prerequisite** | None — start here after cloning |

## In this article

- [Requirements](#requirements)
- [Setup](#setup)
- [Before you open a PR](#before-you-open-a-pr)
- [Where to go next](#where-to-go-next)
- [Stuck?](#stuck)
- [Security](#security)
- [See also](#see-also)

## Requirements

- **Database** — `DATABASE_URL` in repo-root `.env` (always)
- **Go** 1.25+ and **Node** 20+ (npm/npx included; for the web UI)
- **Never commit** `.env` or secrets

> **Warning** — Workspace repo path, agent worker settings, cursor binary, and run timeout are configured in the SPA **Settings** page (`/settings`), not in `.env`. See [docs/configuration.md](docs/configuration.md).

### Desktop app (optional)

The downloadable host is `cmd/hamix-desktop` (Wails). Postgres URL is configured in the UI on first launch (or Settings → Database), stored under `{UserConfigDir}/hamix/desktop.json`. `DATABASE_URL` still overrides for local dev. See [ADR-0095](docs/adr/ADR-0095-desktop-wails-host.md) and [cmd/hamix-desktop/README.md](cmd/hamix-desktop/README.md).

```bash
./scripts/dev-desktop.sh        # Unix — build SPA + run desktop window
.\scripts\dev-desktop.ps1       # Windows
```

Optional: `-Migrate` / `--migrate` (same sugar as `dev.*`). `-Live` / `--live` uses `wails dev` (Vite HMR; requires Wails CLI).

## Setup

1. Copy `.env.example` to `.env` and set `DATABASE_URL`.

2. Migrate (once per schema change or first setup):

```bash
./scripts/migrate.sh       # Unix — chmod +x once if needed
.\scripts\migrate.ps1      # Windows
```

3. Run API + web (servers only):

```bash
./scripts/dev.sh        # Unix
.\scripts\dev.ps1       # Windows
```

Optional convenience: `./scripts/dev.sh --migrate` / `.\scripts\dev.ps1 -Migrate` runs migrate then servers.

**Cloud Postgres:** remote `DATABASE_URL` adds latency to migrate; daily dev uses step 3 only. Run step 2 after pulls that change schema. If you skip migrate, taskapi stderr and `GET /health/ready` report `schema` pending.

API: `http://127.0.0.1:8080` · Web: `http://localhost:5173`

Manual migrate only: `go run ./cmd/dbcheck -migrate` — [Schema migrations in configuration.md](docs/configuration.md).

Taskapi does **not** migrate on dev startup by default. See [Schema migrations in docs/configuration.md](docs/configuration.md).

## Before you open a PR

Verification steps live in `scripts/check-go.sh` / `scripts/check-web.sh` (and PowerShell twins). CI runs those leaf scripts directly — not duplicated commands in `.github/workflows/ci.yml`.

Each **check-script step** (e.g. `web (test-unit)`, `go-tests (core)`) must finish within **2 minutes** (120s). Exceeding that fails the check with a clear budget error (install / `npm ci` is excluded). Split or slim the suite — do not raise the budget casually. Actions plumbing (checkout, setup-node) is not budgeted.

| I want to… | Command |
|------------|---------|
| Run everything locally | `./scripts/check.sh` or `.\scripts\check.ps1` |
| First run / lockfile changed | `./scripts/check.sh --install` or `.\scripts\check.ps1 -Install` |
| Same as CI web deps seed | `./scripts/check-web.sh --install-only --verbose` or `.\scripts\check-web.ps1 -InstallOnly -Verbose` |
| Same as CI Go lint | `./scripts/check-go.sh --lint-only --verbose` or `.\scripts\check-go.ps1 -LintOnly -Verbose` |
| Same as CI Go tests (one group) | `./scripts/check-go.sh --tests-only --group=core --verbose` (groups: `core`, `tasks`, `handlertest`, `task-bcs`, `agents`, `harness`) — includes [coverage gate](docs/domain/testing.md#coverage-floors) |
| Same as CI Go (full local bar) | `./scripts/check-go.sh --verbose` or `.\scripts\check-go.ps1 -Verbose` |
| Same as CI web matrix cell | `./scripts/check-web.sh --verbose --group=test-unit` (after install / `node_modules` present); groups match CI: `lint`, `build`, `test-unit`, `test-components`, … |
| Full local web bar | `./scripts/check-web.sh --verbose` or `.\scripts\check-web.ps1 -Verbose` (`-Install` if needed) |
| Go only (fast) | `./scripts/check.sh --go-only` or `.\scripts\check.ps1 -GoOnly` |
| Full logs | add `--verbose` / `-Verbose` |

Quiet by default: one line per step on success; full tool output only on failure. Each script accepts `--help` / `-Help` for its step list and flags.

Example (quiet success):

```text
Hamix check (Go)
[1/5] gofmt                  ok 6s
...
check OK  5/5 passed  35s

Hamix check (web)
[1/4] web test               ok 22s
...
check OK  4/4 passed  33s
```

Also:

- [ ] Changed an API endpoint → update [docs/api.md](docs/api.md) in the same PR
- [ ] New behavior → add or update a test — see [docs/domain/testing.md](docs/domain/testing.md)
- [ ] User-visible change → update the relevant doc

Coding conventions (where to put API calls, how the live UI updates, etc.): [docs/agent-map.md](docs/agent-map.md), [docs/web.md](docs/web.md).

## Where to go next

Pick **one** row. Do not read the whole tree.

| I want to… | Start here |
| --- | --- |
| **Learn the project** — how docs fit together | [docs/guide.md](docs/guide.md) |
| **Use Hamix** — create tasks, write checklist criteria | [docs/execute-and-verify.md](docs/execute-and-verify.md) |
| **Edit code** — find a subsystem path | [docs/agent-map.md](docs/agent-map.md) |
| **Look up routes, schema, or env vars** | [docs/api.md](docs/api.md), [docs/data-model.md](docs/data-model.md), [docs/configuration.md](docs/configuration.md) |
| **Find any doc by topic** | [docs/README.md](docs/README.md) |

Vertical slice (domain → store → handler → optional web): [docs/agent-map.md](docs/agent-map.md), then `pkgs/tasks/handler/README.md` and [docs/domain/persistence.md](docs/domain/persistence.md).

## Stuck?

| Symptom | Fix |
| --- | --- |
| Full reload on `/tasks/<id>` shows raw JSON | Restart Vite; see `web/vite.config.ts` HTML bypass for `/tasks` proxy |
| SSE connected but Updates timeline empty | `HAMIX_SSE_TEST=1` in `.env`, restart `taskapi` — [docs/configuration.md](docs/configuration.md) |
| Fetch / EventSource errors | Confirm `taskapi` on `:8080` and dev script running |
| No repository for file search / `@`-mentions | Register a git repo on **Repositories** (`/repositories`, or `/repositories?register=1`; legacy `/worktrees` redirects) — [docs/domain/worktrees-and-branches.md](docs/domain/worktrees-and-branches.md) |
| Tests fail with database errors | Use `internal/tasktestdb/` (SQLite); gate real Postgres with `//go:build integration` |
| Match API error to logs | `request_id` in JSON body / `X-Request-ID` header on access logs |
| Still failing local checks | Use scoped groups: `.\scripts\check-go.ps1 -TestsOnly -Group <core\|tasks\|handlertest\|task-bcs\|agents\|harness>` (same as CI). Full bar: `.\scripts\check.ps1 -GoOnly`. Avoid `go test ./...` — it pulls in `web/node_modules` test packages and can flake on parallel SQLite. |
| File-size CI fails on “new red-zone file” | Split the file under [CODE_STANDARDS](.cursor/rules/CODE_STANDARDS.mdc) limits, or (legacy only) add the path to `scripts/code-standards-size-baseline.txt` and burn it down later |

## Security

For **undisclosed vulnerabilities**, use [SECURITY.md](SECURITY.md) (private GitHub advisory, not a public issue).

## See also

- [README.md](README.md) — product overview and quick start
- [docs/guide.md](docs/guide.md) — documentation map and learning paths
- [docs/agent-map.md](docs/agent-map.md) — subsystem code paths
