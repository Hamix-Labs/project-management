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

- `pkgs/tasks/store.Store` composes `taskcompose/store.Store` and re-exports draft/template methods via `facade_compose.go`.
- `pkgs/tasks/handler/handler_routes.go` calls `taskcompose/handler.Register` with `NormalizeCompose` and `InstantiateFromTemplate` callbacks for task create.

## Docs

- [docs/api.md](../../docs/api.md) — `/task-drafts*`, `/task-templates*`
- [ADR-0048](../../docs/adr/ADR-0048-bounded-context-taskcompose.md)
