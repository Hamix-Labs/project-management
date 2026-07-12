# ADR-0075: Non-blocking harness notifier contract

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Harness notifiers (`CycleChangeNotifier`, `TaskUpdatedNotifier`, `ProgressNotifier`) run on the agent worker goroutine. SSE adapters previously called `PublishEnrichedTaskUpdated` synchronously with a 5s store load timeout, so a slow SSE hub or DB read could stall the run loop ([HARNESS_IMPROVEMENTS.md](../HARNESS_IMPROVEMENTS.md) P0 #4).

## Decision

1. **Task updated** — enqueue task IDs on a bounded channel (depth 32); a dedicated worker goroutine loads and publishes. Enqueue is non-blocking; overflow increments `hamix_agent_notifier_dropped_total{kind="task_updated"}`.
2. **Cycle change and run progress** — fire-and-forget publish with a 50ms wait cap; slow `Publisher.Publish` does not block the harness callback and records `kind="cycle_change"` or `kind="run_progress"` drops.
3. **Metrics** — `hamix_agent_notifier_dropped_total` registered from `internal/taskapi/agentworker`.

## Consequences

### Positive

- Harness run duration is decoupled from SSE subscriber speed.
- Operators can alert on notifier drops.

### Negative / Trade-offs

- Under extreme load, clients may miss task_updated events (same class of loss as hub slow-subscriber eviction).

## See also

- [docs/domain/sse-hub.md](./domain/sse-hub.md)
- [HARNESS_IMPROVEMENTS.md](../HARNESS_IMPROVEMENTS.md) P0 #4
