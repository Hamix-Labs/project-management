# `pkgs/projects`

Bounded context for first-class projects, curated project context (items + edges), and immutable task context snapshots. Extracted from `pkgs/tasks` per [ADR-0045](../../docs/adr/ADR-0045-bounded-context-projects.md).

HTTP routes (`/projects`, `/projects/{id}/context`, …) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/projects/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `Project`, context types, enums, sentinel errors — stdlib only |
| Store | [`store/`](./store/) | GORM persistence facade; `internal/` holds CRUD; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/projects*` REST handlers and wire DTOs |

## Wiring

- **`cmd/taskapi`** still constructs `pkgs/tasks/store.Store` as the composition root.
- `tasks/store.Store` holds `*projectsstore.Store` and implements `contract.ProjectStore` via delegation ([`facade_projects.go`](../tasks/store/facade_projects.go)).
- Harness and worker load project context through the tasks store facade; contract types use `pkgs/projects/domain`.

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | `pkgs/tasks/*`, GORM |
| `store` | `projects/domain`, GORM, `pkgs/tasks/kernel`, `pkgs/tasks/contract`, `pkgs/tasks/store/model` (task FK checks) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `handler` | `projects/domain`, `pkgs/tasks/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx` | `pkgs/tasks/store` facade, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_projects_boundary`.

## Tests

```powershell
go test ./pkgs/projects/... -count=1
go test ./pkgs/tasks/store/... -run Project -count=1
```

Integration coverage for project CRUD, context graph, and snapshots lives in [`pkgs/tasks/store/facade_projects_test.go`](../tasks/store/facade_projects_test.go) (tasks facade delegates to this package).

## See also

- [docs/api.md](../../docs/api.md) — `/projects*` contract
- [docs/domain/project-context.md](../../docs/domain/project-context.md) — harness context selection
- [pkgs/tasks/contract/project.go](../tasks/contract/project.go) — `ProjectStore` interface
