# pkgs/taskchecklist

Bounded context for **task checklists** — definition rows, verify commands, and per-subject completion.

## Layout

| Path | Role |
| --- | --- |
| `contract/` | HTTP wire types + `ChecklistStore` interface |
| `store/` | GORM facade; delegates to `internal/checklist` |
| `handler/` | `/tasks/{id}/checklist*` routes via `handler.Register` |

## Wiring

- `pkgs/tasks/store.Store` composes `taskchecklist/store.Store` and delegates through `facade_checklist.go`.
- `notifyUnblockedDependents` stays in the tasks facade when checklist completion unblocks dependents.
- `pkgs/tasks/handler/handler_routes.go` calls `taskchecklist/handler.Register` with `NotifyTaskUpdated` for enriched SSE.
- Domain types (`TaskChecklistItem`) remain in `pkgs/tasks/domain`; GORM models stay in `pkgs/tasks/store/model`.

## Docs

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/checklist*`
- [ADR-0051](../../docs/adr/ADR-0051-bounded-context-taskchecklist.md)
