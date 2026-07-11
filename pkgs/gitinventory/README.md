# `pkgs/gitinventory`

Bounded context for registered git repositories, worktrees, branches, live inventory probes, and reconcile/relocate orchestration. Extracted from `pkgs/tasks` per [ADR-0046](../../docs/adr/ADR-0046-bounded-context-gitinventory.md).

HTTP routes (`/git/repositories`, `/git/worktrees`, …) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/gitinventory/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `GitRepository`, `GitWorktree`, `GitBranch`, git error codes — stdlib only |
| Store | [`store/`](./store/) | GORM persistence, reconcile, live worktree inventory; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/git/*` REST handlers and wire DTOs |

## Wiring

- **`cmd/taskapi`** still constructs `pkgs/tasks/store.Store` as the composition root.
- `tasks/store.Store` holds `*gitinventorystore.Store` and implements `contract.GitReadStore` / `contract.GitWriteStore` via delegation ([`facade_git.go`](../tasks/store/facade_git.go)).
- Harness, worker, and agent reconcile load git rows through the tasks store facade; contract types use `pkgs/gitinventory/domain`.
- `handler.Register` receives `contract.ProjectStore` for `GET /git/repositories/{repoId}/projects`.

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | `pkgs/tasks/*`, GORM |
| `store` | `gitinventory/domain`, GORM, `pkgs/tasks/kernel`, `pkgs/tasks/contract`, `pkgs/gitwork` | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `handler` | `gitinventory/domain`, `pkgs/projects/domain`, `pkgs/tasks/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx`, `pkgs/gitwork` | `pkgs/tasks/store` facade, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_gitinventory_boundary`.

## Tests

```powershell
go test ./pkgs/gitinventory/... -count=1
go test ./pkgs/tasks/store/... -run Git -count=1
```

Integration coverage for git CRUD, reconcile, and facade delegation lives in [`pkgs/tasks/store/facade_git_test.go`](../tasks/store/facade_git_test.go) and [`pkgs/gitinventory/store/`](../gitinventory/store/) tests.

## See also

- [docs/api.md](../../docs/api.md) — `/git/*` contract
- [docs/domain/worktrees-and-branches.md](../../docs/domain/worktrees-and-branches.md) — operator model
- [pkgs/tasks/contract/git_read.go](../tasks/contract/git_read.go) — `GitReadStore` interface
- [pkgs/tasks/contract/git_write.go](../tasks/contract/git_write.go) — `GitWriteStore` interface
