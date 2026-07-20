# ADR-0040: Git reconcile v2 (path repair with stable IDs)

**Date:** 2026-06-27  
**Status:** Accepted (superseded in part by [ADR-0081](./ADR-0081-hamix-managed-worktrees.md) — Sync replaces discover/register operator flows; path repair + relocate remain)  
**Deciders:** Engineering (git/worktrees vertical)

## Context

Hamix stores absolute filesystem paths for `git_repositories` and `git_worktrees`. When operators rename or move checkouts, DB paths drift from git porcelain. The original reconcile only diffed paths and could drop/recreate worktree rows, breaking `tasks.worktree_id` and project bindings.

Downstream subsystems (agent worker `WorkingDir`, `/repo/*`, `@`-mention validation) resolve paths through worktree IDs — they recover automatically once DB paths are repaired, but the repair mechanism must preserve IDs.

## Decision

Replace path-only reconcile with a **ReconcileEngine** (`pkgs/tasks/store/reconcile_git.go`) that:

1. **Opens or bootstraps** the repository — try stored main path; when all paths are stale, require `bootstrap_path` and verify same repo via branch HEAD SHA / common dir (`bootstrap_mismatch` on mismatch).
2. **Updates paths in place** for registered linked worktrees matched by branch name (preserves `worktree_id`).
3. **Removes** vanished registered rows when safe (`has_running_task` → skip + report). Does **not** bulk-import unregistered git worktrees — new task worktrees come from allocate (ADR-0081); Sync refreshes metadata after fetch.
4. **Refreshes** branch `head_sha` from git.
5. Exposes structured **`{ status, report }`** on `POST …/reconcile` and thin **`POST …/relocate`** aliases for operator flows. Prefer **`POST …/sync`** for fetch + metadata refresh.

Shared primitives:

- **`gitwork.PathKey`** — normalize paths for compare (slash clean, Windows case-folding). Used by reconcile, inventory, and path checks.
- **`openRepoForReconcile` / `tryOpenRepoPath`** (store) — bootstrap + verify pattern reused by relocate; not a generic non-git path service.

Manual UI: Sync + Relocate when the main path is missing. Stale managed worktrees show a hint after 24h terminal idle; remove is user-confirmed.

Optional startup: `HAMIX_GIT_RECONCILE_ON_STARTUP=repair-only` runs conservative reconcile (no bootstrap, no remove, no `git worktree repair`) when stored main path still stat's OK.

## Consequences

### Positive

- Folder moves no longer orphan tasks when branch names are unchanged.
- One chokepoint fixes worker, repo HTTP, and mentions without duplicate reconcile logic.
- Operators get explicit bootstrap flow instead of silent wrong-path guesses.

### Negative / Trade-offs

- Absolute paths remain in schema; relative-path identity is a future migration (see LLD cross-cutting section).
- Reconcile does not fix Cursor binary paths or worker scratch artifacts — those stay on Settings / env.
- Detached linked worktrees without branch keys need manual `POST …/worktrees/{id}/relocate`.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Delete + re-insert worktrees on path change | Breaks `worktree_id` FK semantics |
| Auto-bootstrap first existing path on startup | Unsafe without operator confirmation |
| Generic PathReconcileService for all Hamix paths | Non-git resources use different identity keys |
| PATCH repository/worktree paths | Two mutation paths; relocate calls same engine instead |

## Downstream consumers (path repair)

| Consumer | Depends on |
|----------|------------|
| Agent worker / harness | `git_worktrees.path` via `worktree_id` |
| `GET /repo/*`, mention validation | `OpenWorktreeRoot` from DB path |
| SPA `/repositories` | Sync + register-repository path discovery (legacy `/worktrees` redirects here) |

Not covered: `app_settings.CursorBin`, `task_cycle_command_runs` artifact paths, historical verify stdout paths.

## See also

- [docs/domain/worktrees-and-branches.md](../domain/worktrees-and-branches.md)
- [docs/api.md](../api.md) — reconcile / relocate routes
- [ADR-0033](./ADR-0033-git-worktrees-and-branches.md), [ADR-0037](./ADR-0037-global-repos-project-tree.md)
