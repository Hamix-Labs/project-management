# ADR-0083: Async task worktree provision

**Date:** 2026-07-23  
**Status:** Accepted  
**Deciders:** Engineering (git/worktrees + task create)

## Context

[ADR-0081](./ADR-0081-hamix-managed-worktrees.md) requires `git fetch origin` and a linked worktree allocate on task create. That work ran **synchronously on `POST /tasks`**, so create latency tracked disk/network git I/O (~7–10s observed). The product still needs Hamix-managed checkouts before agents run, and the SPA still shows deterministic branch/worktree names derived from the task id.

## Decision

1. **Synchronous command, asynchronous fulfillment** — `POST /tasks` persists the task (with `repository_id`/`project_id` validated) and returns `201` **before** allocate. `worktree_id` may be null on the create response.
2. **Eager provision** — Immediately after create, an in-process **WorktreeProvisioner** wakes (plus startup/ticker reconcile) to run the same allocate path as ADR-0081 (`fetch` + `worktree add` + persist rows), then patches `worktree_id`.
3. **Agent gate** — Tasks without `worktree_id` are not worker-ready (`FailedPredicateWorktree`) and are excluded from the ready queue candidates. Ready-queue notify is deferred until provision succeeds (or pickup_not_before still applies after bind).
4. **Fail closed** — Allocate failure marks the task `failed` (status-changed audit) and publishes SSE; do not leave forever-pending ready rows.
5. **UI labels** — Branch/worktree **display names** remain deterministic from task id (`hamix/task-<8 hex>` / filesystem slug). Copy path / Open in stay disabled until a real path exists.
6. **@-mentions on create** — Validate against the repository **main** checkout while the task worktree does not yet exist.

Amends ADR-0081 decision 4: fetch+allocate remain mandatory before **agent run**, not before **HTTP create**.

## Consequences

### Positive

- Create feels like a normal CRUD write; git I/O no longer blocks the SPA.
- Same allocate semantics and fail-closed freshness as ADR-0081 once the agent starts.
- Crash-safe enough for single-process taskapi: DB rows with null `worktree_id` are reconciled on restart (same durability model as the in-memory agent queue).

### Negative / Trade-offs

- Brief window where the task exists without a checkout; UI must show “preparing” and predicted names.
- Multi-replica / durable outbox still deferred (same as agent MemoryQueue).

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Keep sync allocate; only skip fetch | Still blocks on `worktree add`; incomplete |
| Lazy allocate on agent pickup | Delays first run; product wants eager workspace |
| Fire-and-forget goroutine without reconcile | Lost on crash; no durable pending set |
| New task status `provisioning` | Extra state machine surface; null `worktree_id` + ready gate is enough |

## See also

- [ADR-0081](./ADR-0081-hamix-managed-worktrees.md)
- [docs/domain/worktrees-and-branches.md](../domain/worktrees-and-branches.md)
- [docs/api.md](../api.md)
