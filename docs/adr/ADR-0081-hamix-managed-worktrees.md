# ADR-0081: Hamix-managed worktrees

**Date:** 2026-07-19  
**Status:** Accepted  
**Deciders:** Engineering (git/worktrees vertical)

## Context

Operators today register and create linked worktrees by hand, then pick a worktree when creating a task. That exports internal checkout bookkeeping into the product and fights the goal: **Hamix self-manages git checkouts; users pick a repository and inspect results.**

## Decision

1. **Allocate, don't register for tasks** — Creating a task with `repository_id` causes the server to allocate a linked worktree + branch. Operators do not register or create worktrees as a happy-path workflow.
2. **Never bind agents to default/main** — New tasks must not use the repository default branch or the `is_main` worktree row. Branch naming: `hamix/task-<first 8 hex of task id>` (no dashes).
3. **Path layout** — Managed worktrees live under the Hamix app data root (not beside the registered repo / Documents):
   `{ManagedWorktreeRoot}/worktrees/{repoID}/{branchSlug}`  
   where `ManagedWorktreeRoot` defaults to `{os.UserConfigDir()}/hamix` (Windows `%AppData%\hamix`, macOS `~/Library/Application Support/hamix`, Linux `~/.config/hamix`) and may be overridden with `HAMIX_MANAGED_WORKTREE_ROOT`. `branchSlug` is a filesystem-safe form of the branch name.
4. **Fetch before allocate** — Allocation always `git fetch`es `origin` first, then starts the new branch from `origin/<defaultBranch>`. Fail closed if fetch/auth fails; do not silently use a stale local main.
5. **Delete with the owning task** — When a task is deleted and its `worktree_id` still points at a Hamix-managed checkout (`!is_main`, branch name `hamix/task-<8 hex of that task id>`, no remaining tasks share the worktree), Hamix best-effort removes the checkout from disk (force) and deletes the matching branch row/ref. HTTP task delete still succeeds if disk cleanup fails. Idle/stale managed worktrees are **not** auto-deleted without that ownership path — operators may still remove via `DELETE /git/worktrees/{id}?remove_from_disk=true`. Conflict resolution (merge/rebase agents) stays out of scope; allocate/sync fail with a clear error.

## Consequences

### Positive

- Task create only needs a repository (+ project); Hamix owns path, branch, and DB row.
- Checkouts stay out of the user's Documents / repo-parent directory; Open In still uses the stored absolute path.
- Task delete does not leave orphaned managed worktrees behind for the common allocate path.
- Clear refusal path for default/main checkouts.

### Negative / Trade-offs

- Requires a reachable `origin` remote for allocate (local-only repos need a configured remote).
- Existing rows allocated under the old `{dir(repo)}/.hamix/...` layout keep their absolute paths until deleted; only new allocates use the app-data root.
- Operator worktree CRUD is removed in a follow-up cutover; until then APIs may still exist but must not be the product path for new tasks.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Keep operator register/create worktrees | Fights self-managed checkout goal |
| Store managed trees beside the repo under `.hamix/` | Pollutes Documents/repo parent; atypical for app data |
| Auto-delete after terminal idle alone | Risk of losing inspectable work; task-owned delete + user confirm for leftovers preferred |
| Start from local `main` without fetch | Silent stale base; fail-closed preferred |

## See also

- [ADR-0033](./ADR-0033-git-worktrees-and-branches.md)
- [ADR-0037](./ADR-0037-global-repos-project-tree.md)
- [ADR-0040](./ADR-0040-git-reconcile-v2.md)
- [docs/domain/worktrees-and-branches.md](../domain/worktrees-and-branches.md)
- [docs/configuration.md](../configuration.md) (`HAMIX_MANAGED_WORKTREE_ROOT`)
