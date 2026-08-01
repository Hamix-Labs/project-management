# ADR-0082: Human approval gate

**Date:** 2026-07-22  
**Status:** Accepted (amended 2026-08-01 — open-PR path)  
**Deciders:** Engineering (task lifecycle)

## Context

Successful execute+verify previously finalized tasks as `status=done` with no human review. Operators need an explicit agent-complete → human-approve gate before dependents unblock and work is considered finished. Shipping code also needs a pull request as the delivery artifact before `done`.

## Decision

1. **Agent finalize → `review`** — After successful execute+verify and checklist completions, the harness sets `status=review` (not `done`). Cycle status remains `succeeded`.
2. **`POST /tasks/{id}/open-pr` queues PR creation** — Requires `X-Actor: user` and `status=review`. Sets `pending_retry` kind `open_pr` (mode `resume`), status `ready`, emits `approval_granted` and `open_pr_requested`. The worker resumes the same Cursor conversation to create the PR via MCP.
3. **Harness open-pr success → `pr_ready`** — After the MCP `hamix.create_pull_request` receipt is accepted, the harness sets `status=pr_ready` and emits `pr_opened`. Dependents are **not** unblocked.
4. **`POST /tasks/{id}/approve` is the only path to `done`** — Requires `X-Actor: user`, current status `pr_ready`, and checklist complete (`ValidateCanMarkDoneInTx`). Emits `status_changed` and `on_task_done` (commits from the latest succeeded cycle). Notifies dependents.
5. **Reject free-form `PATCH`/`Update` to `done` or `pr_ready`** — Clients and agents must use open-pr / approve / harness finalize. Create-time `status=done` with empty/complete checklist remains allowed for seeds/tests via the create path.
6. **Heal stuck `running` after succeeded cycle → `review`** — Same as finalize target.
7. **Do not reuse `TaskGate` / `manual_approval`** for post-work review (gates are pre-dequeue holds).

Polish (rework from `review`) remains `POST /polish` and is independent of open-pr.

**Execute visit policy** — `open_pr` (and instructions-only polish) resolve to `CommitIngestAllowEmptyWhenNoHeadDelta` + a post-execute path that skips claim acceptance; see [ADR-0093](./ADR-0093-mcp-commit-register.md) and `ResolveExecuteVisitPolicy`. Cycle meta stamps `skip_claim_acceptance` (and `polish_skip_verify` only for polish).

## Consequences

### Positive

- Clear separation: agent-complete (`review`) → human-approved delivery (`pr_ready`) → graph completion (`done`).
- Dependencies continue to require predecessor `done`, so mark-done after PR still gates the graph.
- Audit trail for approval and PR open via existing and new event types.

### Negative / Trade-offs

- Tasks can sit in `pr_ready` until an operator marks done (merge automation is a follow-on).
- Host must have `gh` + git push credentials for the open-pr agent run.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| New status `awaiting_review` | `review` already exists and is unused on the agent happy path |
| Allow user `PATCH done` from any status | Leaky gate; Approve becomes optional theater |
| Reuse release gate for post-work approval | Wrong semantics (pre-dequeue vs post-verify) |
| `pr_ready` satisfies dependency edges | Premature unblock before human mark-done / merge |
