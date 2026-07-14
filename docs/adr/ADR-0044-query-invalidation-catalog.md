# ADR-0044: Query invalidation catalog

**Date:** 2026-07-08  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Project and git write paths each called `queryClient.invalidateQueries` inline with inconsistent key sets. SSE `project` / `project_context` frames in `decideSyncFrame` duplicated partial lists. Global git mutations repeated an eight-key `invalidateRepo` block. Project create omitted `projectsByRepo`; list mutations omitted `taskQueryKeys.listRoot()`.

`tasks/sync/` owns **SSE orchestration** for tasks ([ADR-0022](./ADR-0022-task-sync-policy.md)). Cross-vertical invalidation tables do not belong there.

## Decision

1. **Shared catalog** — Pure functions in `web/src/lib/queryInvalidation/`:
   - `decideProjectInvalidationKeys(scope)` — scopes: `list`, `detail`, `context`, `repositoryLink`
   - `decideGitInvalidationKeys(scope)` — scopes: `repositories`, `repository`, `legacyRepositories`, `legacyRepository`
   - `applyQueryInvalidations(queryClient, keys)` — loops `invalidateQueries`

2. **Vertical mutation modules** — Production `invalidateQueries` for projects and worktrees live only under:
   - `web/src/projects/mutations/`
   - `web/src/worktrees/mutations/`

3. **SSE alignment** — `decideSyncFrame` maps `project` and `project_context` frames through `decideProjectInvalidationKeys` (no inline key lists).

4. **Git SSE** — Out of scope; backend emits project hints only.

5. **CI gate** — `scripts/check-code-standards.ps1` fails when `invalidateQueries` appears in `projects/` or `worktrees/` outside `mutations/` (production files).

## Consequences

### Positive

- One truth table for project/git cache effects; SSE and mutations stay aligned
- Gaps closed: `projectsByRepo` on create/link; `listRoot` on list mutations; `detail` on context writes
- Table-driven unit tests lock the contract

### Negative / trade-offs

- `legacy*` git scopes remain in the catalog until callers are removed (legacy stack deleted in [#155](https://github.com/AlexsanderHamir/Hamix/pull/155) / [#156](https://github.com/AlexsanderHamir/Hamix/pull/156))
- Task create/bulk invalidation remains separate ([ADR-0025](./ADR-0025-frontend-data-coherence.md))

## See also

- [docs/web.md](../web.md) §Project/worktree mutation invalidation
- Phase 3 policy centralization — [#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203)
