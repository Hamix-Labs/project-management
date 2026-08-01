# ADR-0094: Global default project

**Date:** 2026-08-01  
**Status:** Accepted  
**Deciders:** Hamix maintainers  
**Supersedes:** [ADR-0042](./ADR-0042-repo-default-projects.md)

## Context

[ADR-0042](./ADR-0042-repo-default-projects.md) seeded a system `"Default"` project per registered repository so task create always had a `project_id` that matched `repository_id` (avoiding `project_repo_mismatch`). Operators found per-repo Defaults confusing: registering a repo created a project they did not ask for, and the Projects list labeled them `Default · {repo}`.

The intended model is: **one** built-in Default for the installation, tasks from any repository may use it, and users create additional **repo-bound** projects when they want organization.

Async worktree provision ([ADR-0083](./ADR-0083-async-task-worktree-provision.md)) previously derived `repository_id` from `projects.repository_id` on reconcile. A global Default with null `repository_id` would break that path unless the task itself stores the repository.

## Decision

1. **One global Default** — A single `{ name: "Default", is_default: true, repository_id: null }` row, seeded by migrate/ensure (stable id `GlobalDefaultProjectID`). Not created on `POST /git/repositories`.
2. **Stop seed-on-register** — Repository registration only registers git inventory (repo + main worktree).
3. **Any-repo binding** — Validators treat `is_default` as compatible with any registered repository / worktree. User projects still require `repository_id` and equality checks.
4. **`tasks.repository_id`** — Persist the create-time repository id so WorktreeProvisioner enqueue and `ListPendingWorktree` reconcile without reading the project repo.
5. **List by repo includes Default** — `GET /git/repositories/{repoId}/projects` returns user projects for that repo **plus** the global Default.
6. **Delete repo** — Cascades only projects with that `repository_id`; the global Default is never deleted.
7. **Migration (rev 28)** — Backfill `tasks.repository_id`, consolidate per-repo defaults into one global Default, remap orphan compose payloads, drop `idx_projects_repo_default`, add `idx_projects_global_default`.

## Consequences

### Positive

- Registering a repository no longer invents a project
- One clear Default in the Projects list
- User projects remain an optional, repo-scoped overlay

### Negative / trade-offs

- Special-case in binding validators (Default is the only unbound project)
- Tasks table gains `repository_id` (also useful for UI before worktree allocate)

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Keep per-repo defaults (ADR-0042) | Product rejection of seed-on-register |
| Optional `project_id` (ADR-0037 Case A) | Still want a system Default bucket |
| Derive repo only from worktree after allocate | Breaks ADR-0083 reconcile while `worktree_id` is null |

## See also

- [docs/data-model.md](../data-model.md)
- [docs/api.md](../api.md)
- [ADR-0083](./ADR-0083-async-task-worktree-provision.md)
