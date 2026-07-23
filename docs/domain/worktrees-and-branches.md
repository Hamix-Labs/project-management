# Worktrees, branches, and @-mentions

How Hamix-managed git worktrees scope agent runs, `/repo/*` autocomplete, and `@`-mention validation in task prompts.

| | |
| --- | --- |
| **Applies to** | `pkgs/gitwork/`, `pkgs/repo/`, git store/handlers, `web/src/worktrees/`, task `worktree_id` / `repository_id` |
| **Audience** | Contributors touching git binding, worker `WorkingDir`, or prompt mention validation |
| **Prerequisite** | [ADR-0081](../adr/ADR-0081-hamix-managed-worktrees.md), [ADR-0033](../adr/ADR-0033-git-worktrees-and-branches.md), [ADR-0039](../adr/ADR-0039-fixed-worktree-branch.md), [data-model.md](../data-model.md) (git tables) |
| **Companion articles** | [execute-agent.md](./execute-agent.md), [agent-supervisor.md](./agent-supervisor.md), [cycle-commits.md](./cycle-commits.md) |

## Overview

Hamix scopes workspace access through **managed git worktrees** (`git_worktrees` rows). Operators register a **repository** by local path. Creating a task with `repository_id` (+ `project_id`) persists the task immediately; the server then **eagerly allocates** a linked worktree and branch (`hamix/task-<8 hex>`) in the background, persists `worktree_id`, and starts from `origin/<defaultBranch>` after `git fetch` ([ADR-0083](../adr/ADR-0083-async-task-worktree-provision.md)). Agents never bind to the main/`is_main` checkout or the repository default branch, and do not pick up a task until `worktree_id` is set.

When no git repository is registered:

- The agent worker supervisor stays **idle** (`idle_reason=no_repository_registered`).
- `GET /repo/*` returns **400** without `worktree_id` query param, or **404** for unknown worktree.
- Prompts with `@`-mentions on create are validated against the repository **main** checkout until the task worktree exists; after allocate, `/repo/*` and mentions use the task `worktree_id`.

Operators manage repositories in the SPA:

- **`/repositories`** — repository list; register and **Delete**.
- **`/repositories?register=1`** — deep link that opens the register-repository modal.
- Legacy **`/worktrees`** and **`/worktrees/:repositoryId`** redirect to `/repositories`.

Happy path:

1. **Register repository** on `/repositories` — path to the main git checkout.
2. **Create task** with `repository_id` (+ project) — `POST /tasks` returns quickly; Hamix then allocates `{ManagedWorktreeRoot}/worktrees/{repoID}/{branchSlug}` (default `{UserConfigDir}/hamix`, override `HAMIX_MANAGED_WORKTREE_ROOT`) asynchronously.
3. Managed worktrees stay internal to Hamix; operators manage repositories from the list, not a worktree detail page. The SPA shows predicted branch/worktree names from the task id while provisioning.

**Task delete:** deleting a task that owns a Hamix-managed worktree (matching `hamix/task-*` branch, not main, no other tasks still bound) best-effort removes that checkout from disk and the matching branch. Disk cleanup failures do not fail `DELETE /tasks/{id}`.

**Stale managed worktrees:** a non-main worktree with no non-terminal tasks whose latest terminal task `updated_at` is older than **24h** may be marked `stale` in the API. Staleness alone does **not** auto-delete; leftovers after non-owned bindings (or failed best-effort cleanup) may be removed via `DELETE /git/worktrees/{id}?remove_from_disk=true` (no SPA worktree browser).

**Unregister vs delete from disk (API):** **Unregister** drops only the Hamix row. **Delete from disk** runs `git worktree remove` and deletes the row. The main worktree cannot be deleted from disk via the API.

**Runtime:** tasks on the same worktree run sequentially (per-worktree gate). Tasks on different worktrees may run in parallel when `HAMIX_AGENT_WORKER_CONCURRENCY` > 1. The worker refuses main/default-branch bindings and verifies HEAD matches the bound branch (no `git checkout` at pickup).

## Sync and path repair

Hamix stores **absolute paths** for repositories and worktrees. Sync / relocate remain available as **HTTP APIs** (`POST /git/repositories/{repoId}/sync`, `POST /git/repositories/{repoId}/relocate`, `POST /git/worktrees/{worktreeId}/relocate`) for automation and ops tooling — they are **not** exposed in the SPA. Task create allocates and refreshes managed worktrees as needed.

**Operator playbook when folders move:**

1. Prefer `git worktree repair` (or `git worktree move`) so git metadata stays consistent.
2. Call `POST /git/repositories/{repoId}/sync`. When the stored main path is missing, the API returns `needs_bootstrap_path`; follow with `POST /git/repositories/{repoId}/relocate` and body `{ path }`.
3. For a single linked worktree with a known new path, use `POST /git/worktrees/{worktreeId}/relocate`.

Conflict resolution (merge/rebase agents) is out of scope — allocate/sync fail with a clear error.

See [ADR-0081](../adr/ADR-0081-hamix-managed-worktrees.md), [ADR-0040](../adr/ADR-0040-git-reconcile-v2.md) (superseded in part), and [git-checkout-resolution.md](./git-checkout-resolution.md).

> **Important** — Workspace trees are **read-only over HTTP**. Mutations happen when the execute agent (or the operator outside Hamix) changes files on disk.

## Key concepts

| Term | Definition |
| --- | --- |
| **Git repository** | A registered main checkout (`git_repositories.path`) |
| **Managed worktree** | Hamix-allocated linked checkout under `{ManagedWorktreeRoot}/worktrees/{repoID}/…` |
| **Branch** | A repo-level ref (`git_branches`); bound to at most one worktree |
| **`WorkingDir`** | `runner.Request.WorkingDir` — task worktree path at dequeue |
| **`@`-mention** | Token in `initial_prompt`: `@path` or `@path(start-end)` |

## Worker and supervisor

- Idle reasons: `no_repository_registered`, `all_worktrees_invalid`, `paused_by_operator`.
- Pre-run: per-worktree gate (`WorktreeGate`); refuse `is_main` / default branch; optional HEAD verify — no checkout at pickup.
- Pool: N queue consumers share one `MemoryQueue` (`HAMIX_AGENT_WORKER_CONCURRENCY`, default 4). Busy worktrees defer pickup via `TryLock` (~5s).
- Delete guard: **409** `has_running_task` when a **running** task targets the worktree or branch.
