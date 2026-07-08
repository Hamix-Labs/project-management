# ADR-0043: Compose git assignment model

**Date:** 2026-07-07  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Template and task compose forms let operators pick a registered repository, project, and worktree. Bugs (worktree revert on save, infinite refetch loops, misleading select labels) traced to **split state**: parent form fields, local component mirror state, and modal callbacks that cleared unrelated fields on mount.

The wire payload also omitted `repository_id`, forcing async `getProject` hydration on template edit.

## Decision

1. **Single assignment model** — `ComposeGitAssignment { repositoryId, projectId, worktreeId }` in `web/src/lib/composeGitAssignment.ts`. All transitions are pure functions: `initFreshAssignment`, `hydrateAssignmentFromPayload`, `selectRepository`, `selectProject`, `selectWorktree`, `applyRepoScopedDefaults`.

2. **Thin React shell** — `useComposeGitAssignment` in `web/src/hooks/useComposeGitAssignment.ts` loads git/project data and dispatches reducer actions. `TaskCreateAssignmentFields` is presentational (no assignment `useEffect`).

3. **Explicit init paths** — Fresh create runs `initFreshAssignment` once when all ids are empty. Edit/template hydrate sets three ids from payload without inferring mode from empty strings.

4. **Strict compose payload** — Templates and drafts require `repository_id`, `project_id`, and `worktree_id` on save. Server validates project belongs to repository and worktree binding rules (`validateComposeGitBinding`).

5. **Destructive cascades live in named actions** — `selectRepository` clears project and worktree; modal wiring must not duplicate that logic inline on every callback.

## Consequences

### Positive

- One file answers “where is git assignment logic?”
- Reducer tests are the behavioral spec; UI tests cover user journeys only
- Template edit round-trips repo/worktree without extra API calls

### Negative / trade-offs

- Legacy templates without `repository_id` must be re-saved or reset in dev (no new DB backfill)
- Compose validation is stricter than task PATCH (which still uses project/worktree binding only)

## Anti-patterns eliminated

| Pattern | Replacement |
| --- | --- |
| Mirror `selectedRepoId` + sync effects | Parent-owned assignment props |
| Modal `onRepositoryChange` clearing three fields inline | Reducer `selectRepository` + adapter |
| CustomSelect first-option fallback | Placeholder when value unknown |
| Derive repo via `getProject` on template open | `repository_id` on payload |

## See also

- [docs/web.md](../web.md) §Task create flow
- [ADR-0042](./ADR-0042-repo-default-projects.md) — per-repo default projects
- [ADR-0039](./ADR-0039-fixed-worktree-branch.md) — worktree binding
