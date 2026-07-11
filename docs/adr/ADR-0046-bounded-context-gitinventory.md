# ADR-0046: Extract `pkgs/gitinventory` bounded context

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Git repository, worktree, and branch inventory — plus reconcile/relocate orchestration — lived inside `pkgs/tasks/handler` and `pkgs/tasks/store`. The domain was already isolated behind `contract.GitReadStore`, `contract.GitWriteStore`, and `store/internal/git/`, but HTTP routes and GORM models still mixed with tasks, projects, and settings. ADR-0045 proved the bounded-context extraction pattern; git inventory was the next candidate (see ADR-0045 future extractions and ADR-0037 global git model).

Tasks and harness reference git rows only via contract types and the composed tasks store facade — no import cycle if gitinventory never imports task CRUD or `pkgs/tasks/handler`.

## Decision

1. **New bounded context** — `pkgs/gitinventory/{domain,store,handler}` owns git entity types, GORM persistence, reconcile helpers, and `/git/*` HTTP routes.

2. **Composition root unchanged** — `cmd/taskapi` still wires `pkgs/tasks/store.Store`; that facade embeds/delegates `contract.GitReadStore` and `contract.GitWriteStore` to `pkgs/gitinventory/store.Store`.

3. **Route registration** — `pkgs/gitinventory/handler.Register` mounts the same URLs from `pkgs/tasks/handler/handler_routes.go`. No path or JSON shape changes.

4. **Contract kernel** — `pkgs/tasks/contract` keeps `GitReadStore`, `GitWriteStore`, and shared DTOs (`ReconcileGitInput`, etc.). Handlers depend on contract interfaces, not the tasks store facade.

5. **Cross-context reads** — `GET /git/repositories/{repoId}/projects` uses `contract.ProjectStore.ListProjectsByRepository` injected at registration time (same database, no projects import from gitinventory store).

6. **Domain purity** — `pkgs/gitinventory/domain` imports stdlib only. Git sentinel errors and codes live in `pkgs/gitinventory/domain/git.go`.

7. **Import gate** — CI rejects `pkgs/gitinventory` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (`scripts/check-go.sh` → `step_gitinventory_boundary`).

8. **Postgres migrate** — `pkgs/tasks/postgres` and `pkgs/tasks/store/model` AutoMigrate import GORM models from `pkgs/gitinventory/store/model`; one database, one binary.

9. **Shared HTTP helpers** — `pkgs/gitinventory/handler` exports `WriteGitStoreError` / `GitErrHTTP` for tasks handlers that still map git errors on non-`/git` routes (e.g. task create with project/worktree binding).

## Consequences

### Positive

- Git-focused tests, logs, and file ownership narrow to `pkgs/gitinventory`
- Template for a future `pkgs/settings/` extraction
- Legacy project-scoped git deletion (Phase 2 dead-code) can proceed against a stable global handler surface
- Web and operator docs unchanged at the HTTP contract layer

### Negative / trade-offs

- Harness and worker still reach git methods through the composed tasks store facade (not a separate binary)
- `pkgs/gitinventory/store` may import `pkgs/tasks/kernel` and `pkgs/tasks/store/model` for FK checks until `pkgs/storekernel/` lands (same debt as projects)
- `git_store_adapter.go` remains in `pkgs/tasks/handler` as a thin compile-time shim for optional service wiring

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Type aliases in `pkgs/tasks/domain` indefinitely | Hides ownership; blocks import lint |
| Split `taskapi` binary | Out of scope; composition root stays in tasks store |
| Move contract git interfaces into gitinventory | Breaks harness/worker imports that already use `pkgs/tasks/contract` |
| Keep handlers in tasks, store-only extraction | Leaves largest god-handler surface and mixed route ownership |

## Future extractions (not this ADR)

| Context | Trigger |
| --- | --- |
| `pkgs/settings/` | Auth / multi-tenant prep |
| `pkgs/storekernel/` | Third store extraction needs shared kernel |
| Legacy project-scoped git HTTP | Phase 2 dead-code after global path is sole surface |

## See also

- [pkgs/gitinventory/README.md](../../pkgs/gitinventory/README.md)
- [docs/agent-map.md](../agent-map.md)
- [ADR-0045](./ADR-0045-bounded-context-projects.md) — prior extraction template
- [ADR-0037](./ADR-0037-global-repos-project-tree.md) — global git model
