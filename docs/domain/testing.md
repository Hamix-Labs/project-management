# Testing and CI

Contributor playbook for Go verification, coverage gates, test tiers, and anti-patterns. Harness-specific recipes live in [harness-testing.md](harness-testing.md).

| | |
| --- | --- |
| **Applies to** | Adding or fixing tests, reproducing CI failures, raising coverage |
| **Audience** | Contributors and agents |
| **Prerequisite** | Repo checkout; [CONTRIBUTING.md](../../CONTRIBUTING.md) setup |

## In this article

- [Verification ladder](#verification-ladder)
- [Six CI test groups](#six-ci-test-groups)
- [Coverage floors](#coverage-floors)
- [Test tiers](#test-tiers)
- [Where to add tests](#where-to-add-tests)
- [Anti-patterns](#anti-patterns)
- [See also](#see-also)

## Verification ladder

Work from narrow to full bar — same order CI expects.

| Step | Command (PowerShell) | When |
| --- | --- | --- |
| 1. Scoped group | `.\scripts\check-go.ps1 -TestsOnly -Group <core\|tasks\|handlertest\|task-bcs\|agents\|harness> -Verbose` | Iterating on one area |
| 2. Full Go | `.\scripts\check.ps1 -GoOnly` | Before pushing Go-only changes |
| 3. Full repo | `.\scripts\check.ps1` | Before opening a PR (add `-Install` when `web/package-lock.json` changed) |

Unix: `./scripts/check-go.sh --tests-only --group=harness --verbose`, etc.

**Do not** use `go test ./...` as the primary loop. It includes unrelated packages (e.g. `web/node_modules`), runs more than CI, and can flake on parallel SQLite. Use the check scripts — they mirror [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml).

## Six CI test groups

Package lists are defined in [`scripts/test-groups.sh`](../../scripts/test-groups.sh) / [`scripts/test-groups.ps1`](../../scripts/test-groups.ps1). Lint job asserts every repo package belongs to exactly one group.

| Group | Owns | Examples |
| --- | --- | --- |
| `core` | Binaries, `internal/`, repo/git helpers | `cmd/taskapi`, `internal/taskapi`, `pkgs/repo` |
| `tasks` | Task HTTP/store kernel | `pkgs/tasks/handler`, `pkgs/storekernel` |
| `handlertest` | Handler integration (events/cycles/sse) | `internal/handlertest/events` |
| `handlertest-ui` | Handler integration (compose/checklist) | `internal/handlertest/compose` |
| `task-bcs` | Task-adjacent BCs | `pkgs/taskcore`, `pkgs/draftassist`, `pkgs/settings`, `pkgs/projects` |
| `agents` | Agent worker, runner, queue, reconcile e2e | `pkgs/agents/worker`, `runner`, `internal/taskapi/agentreconcile` |
| `harness` | Agent harness only | `pkgs/agents/harness/...` |

CI runs `go-tests` as a **matrix** — one job per group. Failures name the group in the workflow log.

## Coverage floors

After tests pass, each group job runs the **coverage gate** (`scripts/coverage-gate.sh`) with floors from [`scripts/coverage-baselines.json`](../../scripts/coverage-baselines.json):

| Group | Floor (statement %) |
| --- | ---: |
| `core` | 30 |
| `tasks` | 27 |
| `handlertest` | 27 |
| `handlertest-ui` | 27 |
| `task-bcs` | 27 |
| `agents` | 62 |
| `harness` | 56 |

Run locally (reuses the last test cover profile when invoked via `check-go`):

```powershell
.\scripts\check-go.ps1 -TestsOnly -Group harness -Verbose
```

Standalone gate (runs `go test` + compare):

```powershell
.\scripts\coverage-gate.ps1 -Group harness
```

**Ratchet:** raise floors deliberately in `coverage-baselines.json` (+1–2% per group per quarter) when the team agrees; document the change in the PR.

Calibrate on **Linux CI** if Windows and Linux totals diverge by more than ~1%.

## Test tiers

| Tier | Use | Dependencies |
| --- | --- | --- |
| **Pure** | FSM, policy tables, formatting | stdlib + domain only |
| **Contract** | Handler HTTP via `httptest`, harness via `harness.Run` + fakes | `tasktestdb` / `storefake`, `runnerfake` |
| **Store facade** | `*store.Store` on SQLite | `internal/tasktestdb` |
| **Store facade (Postgres)** | `*store.Store` on Postgres | `tasktestdb.OpenPostgres` — skips unless `HAMIX_TEST_POSTGRES_URL` or `DATABASE_URL` is set |
| **Git wrapper** | Real `git` in `t.TempDir()` | git binary + fakes for store |
| **Opt-in real Cursor** | Runner integration | `HAMIX_TEST_REAL_CURSOR=1` (local only; never CI) |

Handler **policy** are pure packages — unit test without HTTP.

Taskapi assembly smoke: `internal/taskapi/server_smoke_test.go` (`NewHTTPHandler` + `/health`, `/v1/bootstrap`).

## Where to add tests

| Change | Start here |
| --- | --- |
| REST route / JSON | `pkgs/tasks/handler/handler_*_test.go`, `internal/handlertest/` |
| Scheduling rules | `pkgs/tasks/scheduling/*_test.go` (parity with domain) |
| Harness cycle / verify | [harness-testing.md](harness-testing.md) — `harness.Run`, not `worker.Worker` |
| Worker admission / queue | `pkgs/agents/worker/` |
| SSE publish shape | `policy` publish tests + handler SSE tests |
| Bootstrap limits | `policy` read-limit tests + `handler_bootstrap` tests |

## Anti-patterns

- **`worker.Worker` in harness tests** — tests worker pickup, not harness logic; use `harness.Run` (see harness-testing.md).
- **`go test ./...`** — not the CI bar; use `check-go` groups.
- **Direct `tasktestdb` in harness `*_test.go`** — use `storefake`; `import_guard_test.go` enforces.
- **New production funcs without funclogmeasure** — run `check-go` with funclogmeasure; add `//funclogmeasure:skip` only with a valid reason on test fakes and hot-path pure helpers.
- **Polling GitHub Actions** — run `.\scripts\check.ps1` locally instead of `gh run watch`.

## See also

- [harness-testing.md](harness-testing.md) — harness tiers and `harness.Run` recipe
- [CONTRIBUTING.md](../../CONTRIBUTING.md) — PR checklist and Stuck section
- [agent-map.md](../agent-map.md) — subsystem code paths
- [ADR-0026](adr/ADR-0026-backend-data-coherence.md) — SSE / bootstrap coherence
