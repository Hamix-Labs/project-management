# pkgs/taskcompose

Bounded context for **task drafts** and **task templates** — saved compose payloads and reusable blueprints.

## Layout

| Path | Role |
| --- | --- |
| `domain/` | `TaskDraft`, `TaskTemplate` domain types |
| `contract/` | HTTP wire types + `ComposeStore` interface |
| `store/` | GORM facade; delegates to `internal/drafts`, `internal/templates`, `internal/namedpayload` |
| `handler/` | `/task-drafts*` and `/task-templates*` routes via `handler.Register` |

## Wiring

- `internal/taskapi/composition.NewAPI(db)` constructs `taskcompose/store.Store` and exposes draft/template methods on the composition API ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- `pkgs/tasks/handler/handler_routes.go` calls `taskcompose/handler.Register` with `NormalizeCompose` and `InstantiateFromTemplate` callbacks for task create.
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Tests

```powershell
go test ./pkgs/taskcompose/... -count=1
```

## Docs

- [docs/api.md](../../docs/api.md) — `/task-drafts*`, `/task-templates*`
- [ADR-0048](../../docs/adr/ADR-0048-bounded-context-taskcompose.md)
