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

- **`cmd/taskapi`** constructs `internal/taskapi/composition.API` via `composition.NewAPI(db)` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- Composition holds `*taskeventsstore.Store` and exposes event read/append methods on the composition API.
- Event **append** from CRUD, cycles, checklist, and devmirror still uses `storekernel.NextEventSeq` + `storekernel.AppendEvent` inside those subpackages (dual-write); this context owns **reads**, standalone `AppendTaskEvent`, and thread messages.
- `pkgs/tasks/handler/handler_routes.go` calls `taskevents/handler.Register` with `NotifyTaskEventChanged` for SSE after PATCH.
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib, `pkgs/taskcore/domain` (Actor alias) | GORM, `pkgs/tasks/*` |
| `contract` | `taskevents/domain`, `pkgs/taskcore/contract` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `store` | `taskevents/domain`, `taskevents/contract`, `taskevents/store/model`, GORM, `pkgs/storekernel`, `pkgs/taskcore/domain`, `pkgs/obs/calltrace` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `handler` | `taskevents/domain`, `taskevents/contract`, `pkgs/tasks/handlerhttp`, `pkgs/tasks/apijson`, `pkgs/obs/calltrace`, `pkgs/tasks/logctx`, `pkgs/taskcore/domain`, `pkgs/tasks/handler/readpolicy` | `internal/taskapi/composition`, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_taskevents_boundary`.

## Tests

```powershell
go test ./pkgs/taskevents/... -count=1
```

## See also

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/events*`
- [docs/domain/task-events.md](../../docs/domain/task-events.md) — dual-write invariant
- [pkgs/taskevents/contract/events.go](./contract/events.go) — `TaskEventStore` interface
- [ADR-0057](../../docs/adr/ADR-0057-taskevents-domain-model.md) — domain/model ownership
