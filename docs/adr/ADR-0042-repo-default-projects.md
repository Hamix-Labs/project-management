# ADR-0042: Per-repository default projects

**Date:** 2026-07-05  
**Status:** Accepted  
**Deciders:** Hamix maintainers  
**Supersedes:** [ADR-0037](./ADR-0037-global-repos-project-tree.md) §5 Case A (optional `project_id`) and §6 global default project

## Context

Hamix seeded a single global default project (`DEFAULT_PROJECT_ID`) with no `repository_id`. Task create often sent that UUID together with a `worktree_id`, triggering `project_repo_mismatch` because the worktree belongs to a registered repo.

[ADR-0037](./ADR-0037-global-repos-project-tree.md) already models **one repo → many projects** via `projects.repository_id`. The missing piece is operational: a **system default per repo** so engineers need not create a project for basic work, while still allowing optional named projects for organization.

## Decision

1. **System default per repo** — On `POST /git/repositories`, the store creates `{ name: "Default", is_default: true, repository_id: repo.id }` in the same transaction as the repository row. Users never create this via `POST /projects`.

2. **Optional user projects** — `POST /projects` requires `repository_id` and creates **additional** projects only. `is_default` is not settable on the public API.

3. **Mandatory task binding** — `POST /tasks` requires `project_id` and `worktree_id`. The repo default satisfies `project_id`. Server validates `project.repository_id == worktree.repository_id`.

4. **Remove global default** — Delete the global default project row, stop seeding it in `postgres.Migrate`, and remove `DefaultProjectID` / `DEFAULT_PROJECT_ID` constants. Repo-level defaults only.

5. **`projects.is_default`** — Boolean flag; at most one `is_default=true` row per `repository_id`. Default projects are non-deletable and have fixed name/status (same protection as the former global default).

## Consequences

### Positive

- Same ergonomics as today's global default, scoped correctly per repo
- Fixes `project_repo_mismatch` for the common "just create a task" path
- One repo can host many projects; default is the zero-config entry point

### Negative / trade-offs

- Data migration for installations referencing the global default UUID
- Task create no longer supports `project_id` null (default covers that case)
- Slightly more rows in `projects` (one default per registered repo)

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Optional project at create (ADR-0037 Case A) | User wants explicit repo → project → worktree; default covers "no custom project" |
| User creates default on first visit | Extra friction; contradicts product intent |
| Archive global default row | User requires full removal; repo defaults only |
| `git_repositories.default_project_id` FK | `is_default` on project keeps metadata colocated |

## See also

- [docs/data-model.md](../data-model.md) — projects and task create rules
- [docs/api.md](../api.md) — `POST /git/repositories`, `GET /git/repositories/{repoId}/projects`, `POST /tasks`
