# ADR-0049: Extract `/repo/*` HTTP to `pkgs/repo/handler`

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Workspace search, file preview, line-range validation, and commit diff lived in `pkgs/tasks/handler/repo_handlers.go` with provider wiring in `repo_provider.go`. Path resolution and mention validation already live in `pkgs/repo/`; only HTTP adapters and worktree-scoped root resolution were mixed into the tasks handler package.

Prior bounded-context extractions (projects, gitinventory, settings — ADR-0045–0047) established the pattern: register sibling routes from `handler_routes.go`, keep contract tests on the full mux in `pkgs/tasks/handler`, and move provider types to the domain package they serve.

## Decision

1. **HTTP only** — `pkgs/repo/handler` owns the four `/repo/*` routes. No new store layer; `pkgs/repo` keeps path/search/mention logic.

2. **Provider colocation** — `RepoProvider`, `GitWorktreeResolver`, reason constants, and constructors move to `pkgs/repo/provider.go`.

3. **Route registration** — `repohandler.Register(m, repohandler.Deps{Provider: h.repoProv})` from `pkgs/tasks/handler/handler_routes.go`. No URL or JSON shape changes.

4. **Tasks handler compat** — `pkgs/tasks/handler/repo_compat.go` type-aliases `RepoProvider` and re-exports constructors/constants so `WithRepoProvider`, mention validation (`handler_task_git_binding.go`), and existing tests compile without churn.

5. **Contract tests** — `handler_http_repo_test.go` and settings-provider repo tests stay in `pkgs/tasks/handler` (full mux dependency; same approach as settings/gitinventory SSE cross-tests).

6. **Import gate** — CI rejects `pkgs/repo/handler` importing `pkgs/tasks/handler` (`scripts/check-go.sh` → `step_repo_handler_boundary`).

## Consequences

### Positive

- `/repo/*` ownership aligns with `pkgs/repo/` path and mention code
- Smaller tasks handler surface; provider reusable without HTTP imports
- No web or `docs/api.md` contract changes

### Negative / trade-offs

- HTTP helpers (`writeJSON`, invalid-input stripping) duplicated lightly vs `pkgs/gitinventory/handler` until a shared thin helper package exists
- Mention validation still reads `h.repoProv` on the tasks handler (intentional — task create/patch owns the gate)

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Full bounded context with store | No persistence beyond git worktree rows (already gitinventory) |
| Keep handlers in tasks, provider-only move | Leaves route ownership split from `pkgs/repo` |
| Move contract tests to `pkgs/repo/handler` | Would import tasks handler or duplicate mux wiring |

## See also

- [pkgs/repo/handler/README.md](../../pkgs/repo/handler/README.md)
- [docs/domain/workspace-repo.md](../domain/workspace-repo.md)
- [ADR-0047](./ADR-0047-bounded-context-settings.md) — prior extraction pattern
