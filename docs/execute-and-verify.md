# Execute and verify

How Hamix runs a task in two phases with one agent. Execute implements work; the verify phase judges your checklist. This article explains what your criteria mean in practice.

| | |
| --- | --- |
| **Applies to** | Creating tasks, writing done criteria (checklist items), reviewing failed cycles |
| **Audience** | Operators and anyone defining work for the agent worker |
| **Related articles** | [domain/done-criteria.md](./domain/done-criteria.md), [domain/execute-agent.md](./domain/execute-agent.md), [domain/verify-agent.md](./domain/verify-agent.md), [ADR-0090](./adr/ADR-0090-command-only-verify.md) |

## In this article

- [Overview](#overview)
- [One task at a time](#one-task-at-a-time)
- [Execute and verify phases](#execute-and-verify-phases)
- [Report files (behind the scenes)](#report-files-behind-the-scenes)
- [Creating tasks and criteria](#creating-tasks-and-criteria)
- [What happens when a task runs](#what-happens-when-a-task-runs)
- [Dedicated worktree (recommended)](#dedicated-worktree-recommended)
- [Do not edit the workspace during verify](#do-not-edit-the-workspace-during-verify)
- [Writing good criteria](#writing-good-criteria)
- [Failures and operator retry](#failures-and-operator-retry)
- [What you see in the UI](#what-you-see-in-the-ui)
- [See also](#see-also)

## Overview

Every task with done criteria goes through execute, then verification:

1. **Execute phase.** Implements the task and states what it believes it finished.
2. **Verify phase.** For criteria **with** shell verify commands, the same agent reviews worker-collected evidence and judges whether each command's `expected_outcome` matches captured output. Criteria **without** verify commands are accepted from the execute self-claim when `claimed_done: true` (`verified_by=execute_claim`) — no second Cursor pass.

A task reaches **done** only when every active criterion is accepted (`verified_by=execute_agent` or `execute_claim`), not when execute merely claims success.

You define the contract when you create the task: the task description and checklist items. The system handles the rest.

> **Important:** New tasks require **at least one** checklist item (done criterion). Each item needs clear, checkable text.

## One task at a time

As of today, you can create as many tasks as you need. There is no cap on how many you can add to the board.

The agent worker still runs **one task at a time**. If you create 100 tasks and they are all ready for the agent, only one is picked up and executed at once. The rest wait in line until the current run finishes (success or failure), then the next eligible task starts.

**Start over** and **Resume from failure** follow the same rule. They only appear when a task has already **failed**, so that task is not running anymore. The action does not start work immediately: it queues the task as ready with your retry choice saved. If another task is executing when you click retry, your task waits in line like any other ready task. Hamix does not block the button because another task is in flight; the single worker prevents two runs at once.

Tasks that are blocked (for example, waiting on dependencies or a deferred pickup time) stay out of the queue until they become ready.

## Execute and verify phases

| Phase | Role | Trusted for final acceptance? |
| --- | --- | --- |
| **Execute** | Reads your task prompt and criteria, changes the repo, commits when required, and reports what it claims to have done. | **No.** Self claim only. |
| **Verify** | Claim-only items: harness accepts execute self-claim. Command-backed items: same agent judges command output vs `expected_outcome` after worker shell checks. | **Yes.** Sole authority for marking criteria done on success. |

The worker runs shell verify commands **before** the verify LLM pass and feeds that output into the prompt. That is independent evidence — not a separate AI judge.

## Report files (behind the scenes)

While a cycle runs, agents write short JSON reports to a **worker scratch folder** outside your git repo (`HAMIX_WORKER_REPORT_DIR`). You do not create, edit, or open these files.

| File | Written by | Purpose |
| --- | --- | --- |
| `criteria-report.json` | Execute agent | Per criterion **self claim**: `claimed_done` + `evidence`. Optional git `commits[]`. |
| `verify-report.json` | Execute agent (verify phase) | Per criterion **verdict**: `verified` + `reasoning`. |

These files are **temporary**. The worker parses them once, stores durable results in the database, and deletes the scratch folder when the cycle ends. For support and history, use the task UI (checklist, cycle events, verdicts), not the JSON paths.

> **Note:** If execute sets `claimed_done: false` for a criterion, verify is **skipped** for that item and the cycle fails that gate immediately.

## Creating tasks and criteria

When you create a task, you supply:

| Field | What it drives |
| --- | --- |
| **Task description** (`initial_prompt`) | What the execute agent implements. |
| **Done criteria** (checklist items) | Acceptance requirements. Each has a stable `id` and readable `text`. |
| **Verify commands** (optional, per criterion) | Read only shell checks (e.g. `go test ./...`) whose output verify can inspect. Up to five per item. |

**Edit locks:** After the agent picks up the task (`running`), you cannot add or change criterion definitions. Plan acceptance requirements **before** pickup.

**Not the same as release gates:** `task.gate` is a separate operator release mechanism. Done criteria control whether work is accepted as complete.

## What happens when a task runs

```text
1. Execute agent runs in your repo
   → implements the task
   → writes criteria-report.json (self claims per criterion)

2. Gate
   → any claimed_done: false → fail (no verify for that item)

3. Verify
   → claim-only (no verify_commands): accept execute claim (execute_claim)
   → command-backed: worker runs shell checks, then verify LLM judges output

4. Decision (one-shot)
   → all pass → task marked done; checklist completions recorded
   → any fail → cycle fails; use Retry / Start over for a new attempt
```

## Managed worktrees (isolation boundary)

Each task binds a Hamix-managed git worktree (`worktree_id`). Execute, verify shell commands, and Cursor `--workspace` all use that worktree path — not a global `repo_root`. Creating a task with `repository_id` allocates a linked worktree under `{ManagedWorktreeRoot}/worktrees/...` (see [domain/worktrees-and-branches.md](./domain/worktrees-and-branches.md)).

Tasks on **different** worktrees may run in parallel up to `app_settings.agent_task_parallelism` (Settings → **Max parallel tasks**). Tasks on the **same** worktree stay sequential (`WorktreeGate`).

Register a repository on `/repositories`, then create tasks against it. Do not point the agent at your day-to-day main checkout unless you intend to share that directory.

## Do not edit the task worktree during verify

Execute and verify both run in the **task’s bound worktree**. There is no extra sandbox beyond that checkout.

While the **verify** phase is running, the worker snapshots git state before and after judgment. If you save files, commit, checkout, or otherwise change the working tree or `HEAD` during that window, the cycle terminates as **`verify_tampered`** — a hard failure with **no retry**, not a normal verify miss.

| When you edit | Typical outcome |
| --- | --- |
| During **verify** (commits, new edits, checkout) | `verify_tampered` — cycle fails terminally |
| Before verify starts (execute finished, verify not yet running) | Verify may still run, but judges the combined repo state (your edits + the agent's work) |
| During **execute** | No integrity check; verify later sees whatever the tree contains |

**Practical rule:** treat the task worktree as read-only from the moment verify starts until the cycle succeeds or fails. Edit other checkouts freely.

Mechanism and rationale: [domain/verify-agent.md](./domain/verify-agent.md#integrity-enforcement), [ADR-0003](./adr/ADR-0003-verify-component-upgrade.md).

## Writing good criteria

Write criteria the agent can evaluate in both phases without guesswork.

**Do**

- One clear outcome per item.
- Observable behavior: endpoints, tests, files, commands, status codes.
- Short, specific text tied to the task goal.

**Examples**

- `GET /health returns 200 with {"status":"ok"}.`
- `go test ./pkgs/tasks/handler/... passes.`
- `New handler returns 404 when the task id is missing.`

**Avoid**

- Vague goals: “code is clean”, “feature works”, “looks good”.
- Criteria that require subjective judgment only.
- Destructive verify commands (mutate the tree, install globally, etc.). Use read only checks.

> **Warning:** A verify command exiting 0 does **not** automatically mark a criterion done. The verify phase still makes the final call.

## Failures and operator retry

Each cycle is **one-shot** ([ADR-0090](./adr/ADR-0090-command-only-verify.md)): one execute, at most one command-verify pass. Any gate or verify failure **terminates the cycle** — there is no in-cycle execute↔verify retry budget.

When a cycle fails, use **Start over** or **Resume from failure** on the task detail page to queue a **new** attempt. **Resume from failure** can carry forward criteria that already passed verify on the parent cycle so the agent does not redo settled work.

**When does the checklist update?** Only when the full run succeeds and the task reaches **done**. Until then, partial progress inside a failed cycle does not create permanent checklist completions.

If the run ends in failure, checklist items stay unsatisfied and the task does not move to **done**, even if some criteria would have passed on that attempt.

## What you see in the UI

| You want to know… | Where to look |
| --- | --- |
| What must be done | Task detail → Done criteria (checklist) |
| Progress during a run | Checklist satisfied counts; cycle / phase events |
| Why verify failed | Cycle events, verification details, verdicts API |
| Whether the task is truly done | Task status `done` after a **succeeded** cycle |

You do not need access to `criteria-report.json` or `verify-report.json` for normal operation.

## See also

| Doc | Why read it |
| --- | --- |
| [data-model.md](./data-model.md) (Checklist) | Schema, edit locks, verdict tables |
| [domain/done-criteria.md](./domain/done-criteria.md) | Full verification loop and wire contracts |
| [domain/execute-agent.md](./domain/execute-agent.md) | Execute prompt and report format |
| [domain/verify-agent.md](./domain/verify-agent.md) | Verify prompt, shell checks, integrity rules |
| [configuration.md](./configuration.md) | Execute runner, `HAMIX_WORKER_REPORT_DIR` |
| [api.md](./api.md) | Create task body (`checklist_items`), checklist routes |
