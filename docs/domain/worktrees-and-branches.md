# Worktrees, branches, and @-mentions

How Hamix-managed git worktrees scope agent runs, `/repo/*` autocomplete, and `@`-mention validation in task prompts.

| | |
| --- | --- |
| **Applies to** | `pkgs/gitwork/`, `pkgs/repo/`, git store/handlers, `web/src/worktrees/`, task `worktree_id` / `repository_id` |
| **Audience** | Contributors touching git binding, worker `WorkingDir`, or prompt mention validation |
| **Prerequisite** | [ADR-0081](../adr/ADR-0081-hamix-managed-worktrees.md), [ADR-0033](../adr/ADR-0033-git-worktrees-and-branches.md), [ADR-0039](../adr/ADR-0039-fixed-worktree-branch.md), [data-model.md](../data-model.md) (git tables) |
| **Companion articles** | [execute-agent.md](./execute-agent.md), [agent-supervisor.md](./agent-supervisor.md), [cycle-commits.md](./cycle-commits.md) |

## Overview

Hamix scopes workspace access through **managed git worktrees** (`git_worktrees` rows). Operators register a **repository** by local path. Creating a task with `repository_id` (+ `project_id`) causes the server to **allocate** a linked worktree and branch (`hamix/task-<8 hex>`), persist `worktree_id`, and start from `origin/<defaultBranch>` after `git fetch`. Agents never bind to the main/`is_main` checkout or the repository default branch.

When no git repository is registered:

- The agent worker supervisor stays **idle** (`idle_reason=no_repository_registered`).
- `GET /repo/*` returns **400** without `worktree_id` query param, or **404** for unknown worktree.
- Prompts with `@`-mentions require `worktree_id` (set by allocate on create).

Operators manage repositories in the SPA:

- **`/worktrees`** — repository list; register a main checkout here.
- **`/worktrees/:repositoryId`** — inspect managed worktrees, **Sync** (fetch + metadata), relocate when the path moved, remove stale checkouts.
- **`/worktrees?register=1`** — deep link that opens the register-repository modal.

Happy path:

1. **Register repository** on `/worktrees` — path to the main git checkout.
2. **Create task** with `repository_id` (+ project) — Hamix allocates `{dir(repo)}/.hamix/{repoID}/worktrees/{branchSlug}`.
3. **Inspect** the worktree path on the repository detail page; UI diffs and `/repo/*` stay keyed by `worktree_id`.

**Stale hint:** a non-main worktree with no non-terminal tasks whose latest terminal task `updated_at` is older than **24h** shows a non-blocking stale hint. The operator confirms **Remove from disk** (`DELETE /git/worktrees/{id}?remove_from_disk=true`). Hamix does **not** auto-delete worktrees.

**Unregister vs delete from disk:** **Unregister** drops only the Hamix row. **Delete from disk** runs `git worktree remove` and deletes the row. The main worktree cannot be deleted from disk via the API.

**Runtime:** tasks on the same worktree run sequentially (per-worktree gate). Tasks on different worktrees may run in parallel when `HAMIX_AGENT_WORKER_CONCURRENCY` > 1. The worker refuses main/default-branch bindings and verifies HEAD matches the bound branch (no `git checkout` at pickup).

## Sync and path repair

Hamix stores **absolute paths** for repositories and worktrees. Use **Sync** on the repository detail page (`POST /git/repositories/{repoId}/sync`) to `git fetch origin` and refresh registered metadata **without** discovering/registering operator worktrees.

**Operator playbook when folders move:**

1. Prefer `git worktree repair` (or `git worktree move`) so git metadata stays consistent.
2. Click **Sync**. When the stored main path is missing, Hamix returns `needs_bootstrap_path` and opens **Relocate repository**.
3. For a single linked worktree with a known new path, use `POST /git/worktrees/{worktreeId}/relocate`.

Conflict resolution (merge/rebase agents) is out of scope — allocate/sync fail with a clear error.

See [ADR-0081](../adr/ADR-0081-hamix-managed-worktrees.md), [ADR-0040](../adr/ADR-0040-git-reconcile-v2.md) (superseded in part), and [git-checkout-resolution.md](./git-checkout-resolution.md).

> **Important** — Workspace trees are **read-only over HTTP**. Mutations happen when the execute agent (or the operator outside Hamix) changes files on disk.

## Key concepts

| Term | Definition |
| --- | --- |
| **Git repository** | A registered main checkout (`git_repositories.path`) |
| **Managed worktree** | Hamix-allocated linked checkout under `.hamix/{repoID}/worktrees/…` |
| **Branch** | A repo-level ref (`git_branches`); bound to at most one worktree |
| **`WorkingDir`** | `runner.Request.WorkingDir` — task worktree path at dequeue |
| **`@`-mention** | Token in `initial_prompt`: `@path` or `@path(start-end)` |

## Worker and supervisor

- Idle reasons: `no_repository_registered`, `all_worktrees_invalid`, `paused_by_operator`.
- Pre-run: per-worktree gate (`WorktreeGate`); refuse `is_main` / default branch; optional HEAD verify — no checkout at pickup.
- Pool: N queue consumers share one `MemoryQueue` (`HAMIX_AGENT_WORKER_CONCURRENCY`, default 4). Busy worktrees defer pickup via `TryLock` (~5s).
- Delete guard: **409** `has_running_task` when a **running** task targets the worktree or branch.
