# ADR-0082: Human approval gate

**Date:** 2026-07-22  
**Status:** Accepted  
**Deciders:** Engineering (task lifecycle)

## Context

Successful execute+verify previously finalized tasks as `status=done` with no human review. Operators need an explicit agent-complete → human-approve gate before dependents unblock and work is considered finished.

## Decision

1. **Agent finalize → `review`** — After successful execute+verify and checklist completions, the harness sets `status=review` (not `done`). Cycle status remains `succeeded`.
2. **`POST /tasks/{id}/approve` is the only path to `done`** — Requires `X-Actor: user`, current status `review`, and checklist complete (`ValidateCanMarkDoneInTx`). Emits `status_changed`, `approval_granted`, and `on_task_done` (commits from the latest succeeded cycle). Notifies dependents.
3. **Reject free-form `PATCH`/`Update` to `done`** — Clients and agents must use approve. Create-time `status=done` with empty/complete checklist remains allowed for seeds/tests via the create path.
4. **Heal stuck `running` after succeeded cycle → `review`** — Same as finalize target.
5. **Do not reuse `TaskGate` / `manual_approval`** for post-work review (gates are pre-dequeue holds).

Polish (rework from `review`) is a follow-on (queued-run intent + `POST /polish`).

## Consequences

### Positive

- Clear separation: agent-complete (`review`) vs human completion (`done`).
- Dependencies continue to require predecessor `done`, so human approval gates the graph.
- Audit trail for approval via existing event types.

### Negative / Trade-offs

- Mid-rollout without Approve leaves tasks stuck in `review` (Approve ships in the same change).
- Operators who previously `PATCH`ed to `done` must use Approve.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| New status `awaiting_review` | `review` already exists and is unused on the agent happy path |
| Allow user `PATCH done` from any status | Leaky gate; Approve becomes optional theater |
| Reuse release gate for post-work approval | Wrong semantics (pre-dequeue vs post-verify) |
