# Cycle commit tracking

How the worker indexes git commits per task from the MCP commit register, and feeds verify, resume, and the commits API.

| | |
| --- | --- |
| **Applies to** | Agent harness execute/verify phases, `task_cycle_commits`, commits API |
| **Audience** | Contributors touching `pkgs/agents/harness`, cycle store, or cycle detail UI |
| **Prerequisite** | [execute-agent.md](./execute-agent.md) — execute prompt and criteria self-report |
| **Decision record** | [ADR-0093](../adr/ADR-0093-mcp-commit-register.md) (supersedes claim ingest in [ADR-0032](../adr/ADR-0032-agent-claimed-commit-index.md)); [ADR-0014](../adr/ADR-0014-cycle-commit-tracking.md) |

## In this article

- [Overview](#overview)
- [Wire contract](#wire-contract)
- [See also](#see-also)

## Overview

When `app_settings.repo_root` points at a git worktree, the execute agent stages with Shell `git add` and creates commits **only** via MCP `hamix.commit`. Each successful call appends the full HEAD SHA to `commit-register.json` under the cycle report dir. After a successful runner exit, the worker requires exact set equality between the register and `cycle_base_sha..HEAD`, then upserts register SHAs into `task_cycle_commits`.

| Failure | Reason |
| --- | --- |
| Empty / missing register | `execute_missing_commits` |
| HEAD advanced outside register | `execute_unregistered_commits` |
| Register SHA missing from HEAD range / unresolvable | `execute_invalid_commit` |

Verify reads **all commits indexed for the task** via `ListCommitsForTask(task_id)`.

> **Note** — Non-git working directories skip snapshot, ingest, and commit indexing entirely (`git.skipped` in phase details).

## Wire contract

### hamix.commit (execute)

- Input: `message` (required). Commits the **current index only** (no staging, no `--amend`, no `-a`).
- On success: appends `{ sha, message, branch?, written_at }` to `commit-register.json` and returns the full SHA.

### criteria-report.json (execute)

Criteria claims only — **no** `commits[]` ingest. Legacy `commits[]` fields are ignored if present.

### Ingest (worker)

1. Read register via `ParseCommitRegister`.
2. Build `H = rev-list --reverse cycle_base_sha..HEAD` and `R =` normalized register SHAs.
3. Require `set(R) == set(H)`.
4. Upsert on `(cycle_id, sha)` — append-only; never delete or supersede rows.

### Verify prompt

`ListCommitsForTask(task_id)` → `FormatGitContextForPrompt` — full task ledger across cycles/attempts.

### HTTP

- `GET /tasks/{id}/commits` — task-wide deduped by SHA (earliest `committed_at` wins); response ordered newest-first.
- `GET /tasks/{id}/cycles/{cycleId}/verdicts` — per-cycle commit rows (no `status` / `gate_reason`).

## See also

- [execute-agent.md](./execute-agent.md)
- [verify-agent.md](./verify-agent.md)
- [agent-mcp.md](./agent-mcp.md)
- [data-model.md](../data-model.md) — `task_cycle_commits` schema
