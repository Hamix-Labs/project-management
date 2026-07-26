# `pkgs/projects`

Bounded context for first-class projects (CRUD, per-repo defaults, task membership). Extracted from `pkgs/tasks` per [ADR-0045](../../docs/adr/ADR-0045-bounded-context-projects.md). Project memory/context nodes were removed in [ADR-0087](../../docs/adr/ADR-0087-remove-project-context.md).

HTTP routes (`/projects`, `/projects/{id}`) and JSON shapes are registered via `pkgs/projects/handler.Register` from `pkgs/tasks/handler`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `Project`, status enums, sentinel errors — stdlib only |
| Contract | [`contract/`](./contract/) | `ProjectStore` interface + store wire input types |
| Store | [`store/`](./store/) | GORM persistence facade; `internal/` holds CRUD; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/projects*` REST handlers and wire DTOs |

## Wiring

- **`cmd/taskapi`** constructs `internal/taskapi/composition.API` via `composition.NewAPI(db)` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- Composition holds `*projectsstore.Store` and satisfies `contract.ProjectStore` for HTTP callers.
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | `pkgs/tasks/*`, GORM |
| `store` | `projects/domain`, `projects/contract`, GORM, `pkgs/storekernel`, `pkgs/taskcore/store/model` (task FK checks), `pkgs/obs/calltrace` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `handler` | `projects/domain`, `projects/contract`, `pkgs/tasks/handlerhttp`, `pkgs/tasks/apijson`, `pkgs/obs/calltrace`, `pkgs/tasks/logctx`, `pkgs/tasks/realtime` | `internal/taskapi/composition`, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_projects_boundary`.

## Tests

```powershell
go test ./pkgs/projects/... -count=1
```

## See also

- [docs/api.md](../../docs/api.md) — `/projects*` contract
- [pkgs/projects/contract/project.go](./contract/project.go) — `ProjectStore` interface
- [ADR-0055](../../docs/adr/ADR-0055-contract-colocation.md) — contract colocation
- [ADR-0087](../../docs/adr/ADR-0087-remove-project-context.md) — project memory teardown
