# ADR-0080: Task invalidation catalog

**Date:** 2026-07-12  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

[ADR-0044](./ADR-0044-query-invalidation-catalog.md) centralized project and git cache invalidation with `decideProjectInvalidationKeys` / `decideGitInvalidationKeys` and CI gates for `projects/` and `worktrees/`. Task-scoped mutations still called `queryClient.invalidateQueries` inline in hooks (`useTaskPatchFlow`, `useTaskDetailMutations`, `useTaskDeleteFlow`), checklist helpers, and create-flow hooks — with inconsistent scopes (`taskQueryKeys.all` vs `listRoot` + `stats`).

`tasks/sync/` owns SSE-driven invalidation orchestration ([ADR-0022](./ADR-0022-task-sync-policy.md)); mutation post-success paths belong in the catalog + `tasks/mutations/` facade.

## Decision

1. **Catalog** — Add `decideTaskInvalidationKeys(scope)` in `web/src/lib/queryInvalidation/` with scopes:
   - `listStats` — `listRoot()` + `stats()`
   - `detail` — `detail(taskId)`
   - `checklist` — `checklist(taskId)` + `detail(taskId)`
   - `events` — `eventsRoot(taskId)`
   - `drafts` — `drafts()`
   - `templates` — `templates()` prefix (all template list queries)

2. **Mutation facade** — `invalidateTaskCache(queryClient, ...scopes)` in `web/src/tasks/mutations/` dedupes keys and calls `applyQueryInvalidations`.

3. **Callers** — Production task hooks and checklist code import the facade; no direct `invalidateQueries` outside `tasks/mutations/`, `tasks/sync/`, and `lib/queryInvalidation/`.

4. **Narrow list coherence** — Task row mutations use `listStats` instead of `taskQueryKeys.all` where ADR-0025 list coherence applies.

5. **CI gate** — Extend `scripts/check-code-standards.ps1` to fail when `invalidateQueries` appears under `web/src/tasks/` outside `mutations/` and `sync/` (production files).

## Consequences

### Positive

- One truth table for task mutation cache effects; aligns with ADR-0044 pattern
- Table-driven tests lock scopes
- CI prevents regression to inline invalidation in hooks

### Negative / trade-offs

- `tasks/sync/` may still call `invalidateQueries` via `applySyncEffects` until PR #5 dedup
- Template invalidation uses prefix key `["task-templates"]` (partial match)

## See also

- [ADR-0044](./ADR-0044-query-invalidation-catalog.md), [ADR-0025](./ADR-0025-frontend-data-coherence.md)
- Phase 3 policy train — [#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203)
