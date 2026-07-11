# pkgs/taskcycles

Bounded context for **execution cycles** — attempts, phases, stream events, verdict reports, and indexed commits.

## Layout

| Path | Role |
| --- | --- |
| `contract/` | `CycleStore` interface + cycle wire input types |
| `store/` | GORM facade; delegates to `internal/{cycles,reports,commits}` |
| `handler/` | `/tasks/{id}/cycles*`, commits, `/tasks/cycle-failures` REST routes |

## Wiring

- `pkgs/tasks/store.Store` composes `taskcycles/store.Store` and delegates through `facade_cycles.go`, `facade_commits.go`, and `facade_reports.go`.
- HTTP routes register via `pkgs/taskcycles/handler.Register` from `pkgs/tasks/handler/handler_routes.go`.
- Domain types (`TaskCycle`, `TaskCyclePhase`, etc.) remain in `pkgs/tasks/domain`; GORM models stay in `pkgs/tasks/store/model`.
- `FailureSurfaceMessage` is exported from `taskcycles/store` for `pkgs/tasks/store/internal/stats`.

## Docs

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/cycles*`, commits, verdict endpoints
- [ADR-0053](../../docs/adr/ADR-0053-bounded-context-taskcycles.md)
