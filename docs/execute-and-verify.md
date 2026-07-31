# Execute and claim acceptance

How Hamix runs a task: one execute agent implements the work and self-claims done criteria; the harness accepts those claims. There is no separate verify agent or PhaseVerify Cursor pass on new cycles ([ADR-0092](./adr/ADR-0092-execute-owns-verify-commands.md)).

| | |
| --- | --- |
| **Applies to** | Creating tasks, writing done criteria (checklist items), reviewing failed cycles |
| **Audience** | Operators and anyone defining work for the agent worker |
| **Related articles** | [domain/done-criteria.md](./domain/done-criteria.md), [domain/execute-agent.md](./domain/execute-agent.md), [domain/verify-agent.md](./domain/verify-agent.md), [ADR-0092](./adr/ADR-0092-execute-owns-verify-commands.md) |

## In this article

- [Overview](#overview)
- [One task at a time](#one-task-at-a-time)
- [Execute and claim acceptance](#execute-and-claim-acceptance)
- [Report files (behind the scenes)](#report-files-behind-the-scenes)
- [Creating tasks and criteria](#creating-tasks-and-criteria)
- [What happens when a task runs](#what-happens-when-a-task-runs)
- [Managed worktrees (isolation boundary)](#managed-worktrees-isolation-boundary)
- [Writing good criteria](#writing-good-criteria)
- [Failures and operator retry](#failures-and-operator-retry)
- [What you see in the UI](#what-you-see-in-the-ui)
- [See also](#see-also)

## Overview

Every task with done criteria goes through **execute**, then **claim acceptance**:

1. **Execute phase.** Implements the task, runs any operator-listed verify commands itself, and submits what it believes it finished (`claimed_done` + evidence).
2. **Claim acceptance.** The harness accepts every active criterion with `claimed_done: true` as `verified_by=execute_claim`. There is no worker shell re-run and no PhaseVerify Cursor pass.

A task reaches **done** only when every active criterion is accepted that way — not when execute merely finishes without a valid criteria report.

You define the contract when you create the task: the task description and checklist items. The system handles the rest.

> **Important:** New tasks require **at least one** checklist item (done criterion). Each item needs clear, checkable text.

> **Historical note:** Older cycles may still show a `verify` phase, `verified_by=execute_agent`, and `verify-report.json` / command-run rows. New cycles do not.

## One task at a time

As of today, you can create as many tasks as you need. There is no cap on how many you can add to the board.

The agent worker still runs **one task at a time** per worktree (parallelism across different worktrees is configured separately). If many tasks are ready on the same worktree, only one is picked up at once. The rest wait until the current run finishes.

Operator retry via `POST /tasks/{id}/retry` (`fresh` or `resume`) follows the same rule. It only applies when a task has already **failed**. The request queues the task as ready with the retry choice saved; it does not start work immediately if another task owns the worktree.

Tasks that are blocked (for example, waiting on dependencies or a deferred pickup time) stay out of the queue until they become ready.

## Execute and claim acceptance

| Step | Role | Trusted for final acceptance? |
| --- | --- | --- |
| **Execute** | Reads your task prompt and criteria, changes the repo, commits when required, runs any listed verify commands, and reports what it claims to have done. | **No.** Self claim only. |
| **Claim acceptance (harness)** | Gates on `claimed_done` for every active criterion and writes completions with `verified_by=execute_claim` on full pass. | **Yes.** Sole acceptance authority on new cycles. |

Optional **verify commands** on a criterion are instructions for the **execute** agent to self-check before claiming. The harness does **not** re-run them.

## Report files (behind the scenes)

While a cycle runs, the execute agent writes a short JSON report to a **worker scratch folder** outside your git repo (`HAMIX_WORKER_REPORT_DIR`). You do not create, edit, or open this file.

| File | Written by | Purpose |
| --- | --- | --- |
| `criteria-report.json` | Execute agent | Per criterion **self claim**: `claimed_done` + `evidence`. Commits are recorded via `hamix.commit`, not this file. |
| `commit-register.json` | MCP `hamix.commit` | Full SHAs for execute ingest; harness requires set equality with `cycle_base_sha..HEAD`. |

The file is **temporary**. The worker parses it once, stores durable results in the database, and deletes the scratch folder when the cycle ends. For support and history, use the task UI (checklist, cycle events, verdicts), not the JSON path.

> **Note:** If execute sets `claimed_done: false` for any active criterion, claim acceptance fails that gate immediately and the cycle terminates.

## Creating tasks and criteria

When you create a task, you supply:

| Field | What it drives |
| --- | --- |
| **Task description** (`initial_prompt`) | What the execute agent implements. |
| **Done criteria** (checklist items) | Acceptance requirements. Each has a stable `id` and readable `text`. |
| **Verify commands** (optional, per criterion) | Read-only shell checks the execute agent must run and match before claiming done. Up to five per item. |

**Edit locks:** After the agent picks up the task (`running`), you cannot add or change criterion definitions. Plan acceptance requirements **before** pickup.

**Not the same as release gates:** `task.gate` is a separate operator release mechanism. Done criteria control whether work is accepted as complete.

## What happens when a task runs

```text
1. Execute agent runs in the task worktree
   → implements the task
   → runs any listed verify commands
   → writes criteria-report.json (self claims per criterion)

2. Claim acceptance (harness)
   → any claimed_done: false → fail
   → all claimed_done: true → accept as execute_claim

3. Decision (one-shot)
   → all pass → task marked done; checklist completions recorded
   → any fail → cycle fails; queue a new attempt via POST /tasks/{id}/retry
```

## Managed worktrees (isolation boundary)

Each task binds a Hamix-managed git worktree (`worktree_id`). Execute and Cursor `--workspace` use that worktree path — not a global `repo_root`. Creating a task with `repository_id` allocates a linked worktree under `{ManagedWorktreeRoot}/worktrees/...` (see [domain/worktrees-and-branches.md](./domain/worktrees-and-branches.md)).

Tasks on **different** worktrees may run in parallel up to `app_settings.agent_task_parallelism` (Settings → **Max parallel tasks**). Tasks on the **same** worktree stay sequential (`WorktreeGate`).

Register a repository on `/repositories`, then create tasks against it. Do not point the agent at your day-to-day main checkout unless you intend to share that directory.

## Writing good criteria

Write criteria the execute agent can evaluate without guesswork.

**Do**

- One clear outcome per item.
- Observable behavior: endpoints, tests, files, commands, status codes.
- Short, specific text tied to the task goal.
- When adding verify commands, state a clear `expected_outcome` the agent can compare.

**Examples**

- `GET /health returns 200 with {"status":"ok"}.`
- `go test ./pkgs/tasks/handler/... passes.`
- `New handler returns 404 when the task id is missing.`

**Avoid**

- Vague goals: “code is clean”, “feature works”, “looks good”.
- Criteria that require subjective judgment only.
- Destructive verify commands (mutate the tree, install globally, etc.). Use read-only checks.

> **Warning:** A verify command exiting 0 does **not** automatically mark a criterion done. The agent must still claim `claimed_done: true` with evidence; the harness accepts that claim.

## Failures and operator retry

Each cycle is **one-shot** ([ADR-0092](./adr/ADR-0092-execute-owns-verify-commands.md)): one execute, then claim acceptance. Any gate or claim failure **terminates the cycle** — there is no in-cycle execute↔verify retry budget.

When a cycle fails, queue a **new** attempt with `POST /tasks/{id}/retry` (`mode: fresh` or `mode: resume`). Resume mode can carry forward criteria that already passed on the parent cycle so the agent does not redo settled work. The task detail SPA no longer offers Start over / Resume from failure buttons.

**When does the checklist update?** Only when the full run succeeds and the task reaches **done**. Until then, partial progress inside a failed cycle does not create permanent checklist completions.

## What you see in the UI

| You want to know… | Where to look |
| --- | --- |
| What must be done | Task detail → Done criteria (checklist) |
| Progress during a run | Live ticker; activity stream; phase agent reply |
| Why acceptance failed | Cycle events, verification details, verdicts API |
| Whether the task is truly done | Task status `done` after a **succeeded** cycle |

You do not need access to `criteria-report.json` for normal operation.

After execute succeeds, activity may show **Accepting criteria claims…** while the harness finishes claim acceptance (not a separate verify agent).

## See also

| Doc | Why read it |
| --- | --- |
| [data-model.md](./data-model.md) (Checklist) | Schema, edit locks, verdict tables |
| [domain/done-criteria.md](./domain/done-criteria.md) | Criteria lifecycle and completion ledger |
| [domain/execute-agent.md](./execute-agent.md) | Execute prompt and report format |
| [domain/verify-agent.md](./domain/verify-agent.md) | Claim acceptance (post-execute) |
| [configuration.md](./configuration.md) | Execute runner, `HAMIX_WORKER_REPORT_DIR` |
| [api.md](./api.md) | Create task body (`checklist_items`), checklist routes |
