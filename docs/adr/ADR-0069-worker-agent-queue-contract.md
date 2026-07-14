# ADR-0069: Worker QueueStore on taskcore contract

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/agents/worker/store.go` imported `pkgs/tasks/store` for queue wire types (`ReadyTaskQueueCursor`, `AgentPickupResult`, `TaskGitContext`, etc.).

## Decision

1. **`pkgs/taskcore/contract.AgentQueueStore`** owns agent dequeue, pickup wake reads, `ReadyForAgentPickup`, and `ResolveTaskGitContext`.
2. Queue wire types and `FailedPredicate` constants live in `taskcore/contract/agent_queue.go`.
3. **`pkgs/agents/worker.QueueStore`** embeds `AgentQueueStore` + `taskcycles/contract.CycleWorkerStore`.
4. `tasks/scheduling.FailedPredicate` aliases the contract type; taskcore ready paths use contract-shaped queue rows.

## Consequences

### Positive

- Worker package no longer depends on facade types for queue contracts.
- Single source for failed-predicate string values in logs/metrics.

### Negative / Trade-offs

- `*store.Store` facade still implements the composed interface at wiring time.

## See also

- Tier 5A PR5 — [#191](https://github.com/AlexsanderHamir/Hamix/pull/191)
