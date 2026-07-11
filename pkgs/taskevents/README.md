# `pkgs/taskevents`

Bounded context for **task audit events** — the append-only `task_events` timeline, keyset paging, approval-pending probe, and response-thread append. Extracted from `pkgs/tasks` per [ADR-0054](../../docs/adr/ADR-0054-bounded-context-taskevents.md); domain and GORM models colocated per [ADR-0057](../../docs/adr/ADR-0057-taskevents-domain-model.md).

HTTP routes (`/tasks/{id}/events*`) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/taskevents/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `TaskEvent`, `EventType`, `Actor`, response thread — stdlib only |
| Contract | [`contract/`](./contract/) | `TaskEventStore` interface + wire DTOs |
| Store | [`store/`](./store/) | GORM persistence facade; `internal/events/` holds reads/thread append; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/tasks/{id}/events*` REST handlers |

## Wiring

- **`cmd/taskapi`** still constructs `pkgs/tasks/store.Store` as the composition root.
- `tasks/store.Store` embeds `taskevents/store.Store` and delegates through [`facade_events.go`](../tasks/store/facade_events.go).
- Event **append** from CRUD, cycles, checklist, and devmirror still uses `storekernel.NextEventSeq` + `storekernel.AppendEvent` inside those subpackages (dual-write); this context owns **reads**, standalone `AppendTaskEvent`, and thread messages.
- `pkgs/tasks/handler/handler_routes.go` calls `taskevents/handler.Register` with `NotifyTaskEventChanged` for SSE after PATCH.
- `pkgs/tasks/store/model/migrate_models.go` registers `taskevents/store/model` types in FK-safe order.

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | GORM, `pkgs/tasks/*` |
| `contract` | `taskevents/domain`, `pkgs/tasks/domain` (`Task`) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `store` | `taskevents/domain`, `taskevents/contract`, `taskevents/store/model`, GORM, `pkgs/storekernel`, `pkgs/tasks/domain`, `pkgs/tasks/store/model` (task rows only) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `handler` | `taskevents/domain`, `taskevents/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx`, `pkgs/tasks/domain` | `pkgs/tasks/store` facade, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_taskevents_boundary`.

## Tests

```powershell
go test ./pkgs/taskevents/... -count=1
go test ./pkgs/tasks/store/... -run Event -count=1
go test ./pkgs/tasks/handler/... -run Events -count=1
```

Cross-route HTTP contract tests for events remain in [`pkgs/tasks/handler/handler_http_events_*_test.go`](../pkgs/tasks/handler/).

## See also

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/events*`
- [docs/domain/task-events.md](../../docs/domain/task-events.md) — dual-write invariant
- [pkgs/taskevents/contract/events.go](./contract/events.go) — `TaskEventStore` interface
- [ADR-0057](../../docs/adr/ADR-0057-taskevents-domain-model.md) — domain/model ownership
