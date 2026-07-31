# Claim acceptance (post-execute)

How the harness accepts done criteria after execute: claim-only (`execute_claim`) for all criteria, including those with operator `verify_commands` the execute agent was instructed to self-check ([ADR-0092](../adr/ADR-0092-execute-owns-verify-commands.md)).

> **Product default (ADR-0092)** — Every active criterion with `claimed_done: true` is accepted as `verified_by=execute_claim`. There is no worker shell re-run and no PhaseVerify Cursor pass. Criteria with `verify_commands` are listed in the **execute** prompt so the agent can run them before claiming. `claimed_done: false` fails the cycle (`agent_self`). Operators recover via `POST /tasks/{id}/retry` (new cycle).

| | |
| --- | --- |
| **Applies to** | Agent worker harness, cycle verdict API |
| **Audience** | Contributors touching `pkgs/agents/harness` or verdict UI |
| **Prerequisite** | [done-criteria.md](./done-criteria.md) — criteria lifecycle and completion ledger |
| **Companion article** | [execute-agent.md](./execute-agent.md) — execute prompt composition and criteria self-report; [harness.md](./harness.md) — cycle loop |

## Overview

After execute succeeds, the harness loads `criteria-report.json`, accepts claimed criteria, mirrors verdict rows, and on full pass writes `task_checklist_completions` with `verified_by=execute_claim`.

| Role | Responsibility |
| --- | --- |
| **Execute agent** | Implements work; runs any listed verify commands; calls `hamix.submit_criteria_report` with `claimed_done` + evidence. |
| **Worker (harness)** | Gates on `claimed_done`, persists mirrors, applies completions on success. Does **not** re-run commands or open PhaseVerify. |

> **Note** — Tasks with **zero criteria** (legacy) skip claim acceptance; successful execute alone marks the task `done`.

Historical cycles may still show `phase=verify`, `verified_by=execute_agent`, and `command_runs` from the pre-ADR-0092 path.

## Workflow

1. Execute finishes and submits `criteria-report.json`.
2. Harness requires every active (non-locked) criterion id with `claimed_done`.
3. All claims true → completions + task proceeds to review/`done` path as today.
4. Any claim false → terminate with `verification_failed…` (one-shot).

## See also

- [ADR-0092](../adr/ADR-0092-execute-owns-verify-commands.md)
- [execute-agent.md](./execute-agent.md)
- [done-criteria.md](./done-criteria.md)
