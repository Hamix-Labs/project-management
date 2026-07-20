# `pkgs/gitinventory`

Bounded context for registered git repositories, worktrees, branches, live inventory probes, and reconcile/relocate orchestration. Extracted from `pkgs/tasks` per [ADR-0046](../../docs/adr/ADR-0046-bounded-context-gitinventory.md).

HTTP routes (`/git/repositories`, `/git/worktrees`, …) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/gitinventory/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `GitRepository`, `GitWorktree`, `GitBranch`, git error codes — stdlib only |
| Contract | [`contract/`](./contract/) | `GitInventoryStore`, `GitWriteStore` + reconcile wire types |
| Store | [`store/`](./store/) | GORM persistence, reconcile, live worktree inventory; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/git/*` REST handlers and wire DTOs |

## Wiring

- **`cmd/taskapi`** constructs `internal/taskapi/composition.API` via `composition.NewAPI(db)` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- Composition holds `*gitinventorystore.Store` and satisfies `contract.GitInventoryStore` / `contract.GitWriteStore`.
- Harness, worker, and agent reconcile load git rows through composition; contract types use `pkgs/gitinventory/domain`.
- `handler.Register` receives `contract.ProjectStore` for `GET /git/repositories/{repoId}/projects`.
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | `pkgs/tasks/*`, GORM |
| `store` | `gitinventory/domain`, `gitinventory/contract`, GORM, `pkgs/storekernel`, `pkgs/gitwork`, `pkgs/obs/calltrace` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `handler` | `gitinventory/domain`, `pkgs/projects/domain`, `gitinventory/contract`, `pkgs/tasks/apijson`, `pkgs/obs/calltrace`, `pkgs/tasks/logctx`, `pkgs/gitwork` | `internal/taskapi/composition`, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_gitinventory_boundary`.

## Tests

```powershell
go test ./pkgs/gitinventory/... -count=1
```

## See also

- [docs/api.md](../../docs/api.md) — `/git/*` contract
- [docs/domain/worktrees-and-branches.md](../../docs/domain/worktrees-and-branches.md) — operator model
- [pkgs/gitwork/README.md](../gitwork/README.md) — which git package?
- [pkgs/gitinventory/contract/read.go](./contract/read.go) — `GitInventoryStore` interface
- [pkgs/gitinventory/contract/write.go](./contract/write.go) — `GitWriteStore` interface
