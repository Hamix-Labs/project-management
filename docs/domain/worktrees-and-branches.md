# Worktrees, branches, and @-mentions

How Hamix-managed git worktrees scope agent runs, `/repo/*` autocomplete, and `@`-mention validation in task prompts.

| | |
| --- | --- |
| **Applies to** | `pkgs/gitwork/`, `pkgs/repo/`, git store/handlers, `web/src/worktrees/`, task `worktree_id` / `repository_id` |
| **Audience** | Contributors touching git binding, worker `WorkingDir`, or prompt mention validation |
| **Prerequisite** | [ADR-0081](../adr/ADR-0081-hamix-managed-worktrees.md), [ADR-0033](../adr/ADR-0033-git-worktrees-and-branches.md), [ADR-0039](../adr/ADR-0039-fixed-worktree-branch.md), [ADR-0097](../adr/ADR-0097-worktree-stack-layers.md), [data-model.md](../data-model.md) (git tables) |
| **Companion articles** | [execute-agent.md](./execute-agent.md), [agent-supervisor.md](./agent-supervisor.md), [cycle-commits.md](./cycle-commits.md) |

## Overview

Hamix scopes workspace access through **managed git worktrees** (`git_worktrees` rows). Operators register a **repository** by local path. Creating a task with `repository_id` (+ `project_id`) persists the task immediately; the server then **eagerly allocates** a linked worktree and branch (`hamix/task-<8 hex>`) in the background, persists `worktree_id`, and starts from `origin/<defaultBranch>` after `git fetch` ([ADR-0083](../adr/ADR-0083-async-task-worktree-provision.md)). Optionally, `POST /tasks` may include an existing non-main **`worktree_id`** to bind the new task to that workspace without allocating (enqueue). Agents never bind to the main/`is_main` checkout or the repository default branch, and do not pick up a task until `worktree_id` is set.

Tasks that share a `worktree_id` form a **worktree family**. The **root** is the task whose id named the allocate-time managed branch (`hamix/task-<8 hex>`, stored as `git_worktrees.name`); other binders are siblings on the same checkout, each with its own **stack layer** branch of the same naming form ([ADR-0097](../adr/ADR-0097-worktree-stack-layers.md)). Read APIs expose computed `worktree_root_task_id` (not a FK). `GET /tasks?worktree_id=` lists the family. This is workspace sharing, not subtasks ([ADR-0010](../adr/ADR-0010-remove-subtasks.md)).

Root allocate runs `gh stack init` so every managed worktree is a local GitHub stack (possibly of one layer). When an enqueued task first runs, Hamix runs `gh stack add` for that task’s layer from the current tip (parent may still be in progress). `git_worktrees.branch_id` tracks the **active** layer under the worktree gate.

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

**Task delete:** when no tasks remain bound to a non-main `hamix/task-*` worktree, delete best-effort removes that checkout from disk and the matching branch — including when the last remaining binder was a sharer (not the original allocator). Disk cleanup failures do not fail `DELETE /tasks/{id}`.

**Stale managed worktrees:** a non-main worktree with no non-terminal tasks whose latest terminal task `updated_at` is older than **24h** may be marked `stale` in the API. Staleness alone does **not** auto-delete; leftovers after non-owned bindings (or failed best-effort cleanup) may be removed via `DELETE /git/worktrees/{id}?remove_from_disk=true` (no SPA worktree browser).

**Unregister vs delete from disk (API):** **Unregister** drops only the Hamix row. **Delete from disk** runs `git worktree remove` and deletes the row. The main worktree cannot be deleted from disk via the API.

**Runtime:** tasks on the same worktree run sequentially (per-worktree gate). Tasks on different worktrees may run in parallel up to `app_settings.agent_task_parallelism` (Settings → **Max parallel tasks**). The worker refuses main/default-branch bindings. At pickup it ensures the task’s stack layer (checkout + rebind), then verifies HEAD matches the active bound branch.

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

## File listing for `@`-mentions

`GET /repo/files` returns every referenceable path under a worktree in one sorted list, so the prompt editor caches it and ranks matches in the browser instead of searching the filesystem per keystroke.

The list comes from `git ls-files --cached --others --exclude-standard`, which makes it **gitignore-aware without a hand-maintained exclusion list**: git applies nested `.gitignore` files, `.git/info/exclude`, and `core.excludesFile` itself. That matters beyond noise, because a mentioned file is inlined into the prompt sent to an agent — an ignored `.env` should not be offerable. `--cached` keeps tracked-but-ignored files listed, since a file already under version control is one the operator can legitimately reference.

A worktree that is not a git work tree (a plain registered directory, or a bare repository) falls back to a directory walk with the fixed skip list `/repo/search` uses, reported as `source: "walk"`. Ignore rules do not apply there.

The listing is capped at 50 000 paths (`repo.MaxFileListPaths`), sized for the browser holding it in memory rather than for the filesystem. Past that the response sets `truncated` and the editor says the list is partial.

`/repo/search` is unchanged and still serves the template repo-scope picker and mention validation.

## Worker and supervisor

- Idle reasons: `no_repository_registered`, `all_worktrees_invalid`, `paused_by_operator`.
- Pre-run: per-worktree gate (`WorktreeGate`); refuse `is_main` / default branch; optional HEAD verify — no checkout at pickup.
- Pool: N queue consumers share one `MemoryQueue` (`app_settings.agent_task_parallelism`, default 150). Busy worktrees defer pickup via `TryLock` (~5s).

### Known limits (not full process isolation)

Worktree binding isolates the **file workspace** (execute/verify cwd, MCP merge path, `/repo/*`). Concurrent tasks still share:

| Limit | Effect |
| --- | --- |
| Shared Cursor home | Parallel Cursor CLIs inherit one OS user profile (`HOME` / `APPDATA`) — account/cache cross-talk, not repo files |
| Process-wide cancel / reload | `POST /settings/cancel-current-run` and worker respawn on material settings change abort **all** in-flight slots |
| Shared `.git` object store | Linked worktrees of the same repository share objects and locks |

See also [agent-queue.md](./agent-queue.md).
- Delete guard: **409** `has_running_task` when a **running** task targets the worktree or branch.
