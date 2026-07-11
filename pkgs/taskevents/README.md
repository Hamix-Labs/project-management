# pkgs/taskevents

Bounded context for **task audit events** — the append-only `task_events` timeline, keyset paging, approval-pending probe, and response-thread append.

## Layout

| Path | Role |
| --- | --- |
| `contract/` | `TaskEventStore`, `TaskEventsPage`, `TaskGetter` |
| `store/` | GORM facade; delegates to `internal/events` for reads and thread append |
| `handler/` | `/tasks/{id}/events*` routes via `handler.Register` |

## Wiring

- `pkgs/tasks/store.Store` composes `pkgs/taskevents/store.Store` and delegates through `facade_events.go`.
- Event **append** from CRUD, cycles, checklist, and devmirror still uses `storekernel.NextEventSeq` + `storekernel.AppendEvent` inside those subpackages (dual-write); this context owns **reads**, standalone `AppendTaskEvent`, and thread messages.
- `pkgs/tasks/handler/handler_routes.go` calls `taskevents/handler.Register` with `NotifyTaskEventChanged` for SSE after PATCH.
- Domain types (`TaskEvent`, `EventType`) remain in `pkgs/tasks/domain`; GORM models stay in `pkgs/tasks/store/model`.

## Docs

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/events*`
- [ADR-0054](../../docs/adr/ADR-0054-bounded-context-taskevents.md)
