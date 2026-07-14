# ADR-0020: Realtime SSE Layout

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-19
**Status:** Accepted
**Deciders:** Engineering

## Context

The SSE hub lived in a single red-zone file [`pkgs/tasks/handler/sse.go`](../../pkgs/tasks/handler/sse.go) (~660 lines) mixing wire types, hub concurrency, HTTP streaming, and handler notify glue. The agent worker supervisor imported `*handler.SSEHub` and handler wire types, coupling process assembly to the HTTP package.

[`docs/domain/sse-hub.md`](../domain/sse-hub.md) already documents logical domains; code layout did not match. Coalesce policy was embedded in the hub with no pure tests outside integration suites.

## Decision

Introduce **`pkgs/tasks/realtime`** as the stable import surface for wire types, coalesce policy, and the `Publisher` port. Split handler transport across focused files; **hub implementation initially stayed in `handler`** — later moved to `realtime` per [ADR-0080](ADR-0080-sse-hub-realtime-ownership.md).

| Package / file | Responsibility |
|----------------|----------------|
| `pkgs/tasks/realtime` | `ChangeType`, `Event`, `RunProgressPayload`, `CoalesceKey`, `Publisher`, **`SSEHub`** (ADR-0080) |
| `handler/sse_types.go` | Type aliases to `realtime` for backward-compatible handler API |
| `handler/sse_hub_alias.go` | Thin aliases to `realtime.SSEHub` constructors |
| `handler/sse_stream.go` | `streamEvents`, frame writers, reconnect replay |
| `handler/sse_notify.go` | Handler `notify*` helpers + store hydration |

**`internal/taskapi/agentworker`** accepts `realtime.Publisher` instead of a concrete hub. Production wiring constructs `realtime.NewSSEHubWith(...)` in `cmd/taskapi/run_helpers.go`.

### Dependency rules

```
realtime     → stdlib + middleware (SSE metrics) + calltrace   (amended ADR-0080)
handler      → realtime, middleware, calltrace
agentworker  → realtime (Publisher), not handler for publish paths
cmd/taskapi  → realtime (concrete hub), handler (HTTP)
```

**Forbidden:** `realtime` importing `handler`. **Deferred:** store-origin change notifier (publish after commit from store facade).

### Non-goals (this ADR)

- No microservices, Kafka, or external pub/sub
- No Postgres outbox (future ADR when multi-replica)
- No moving all handler `notify*` call sites to store in this change
- No frontend `useTaskEventStream` split (separate track)

## Consequences

### Positive

- Hub files under file-size green/yellow bar
- Coalesce policy table-tested without hub mutex
- Supervisor decoupled from HTTP handler package for publish
- Clear seam for future store-backed `ChangeNotifier`

### Negative / Trade-offs

- Type aliases in handler preserve API but duplicate names (`TaskChangeEvent` vs `realtime.Event`)
- `funclogmeasure` path updates when symbols move between files

## Related

- [ADR-0019](ADR-0019-agentworker-internal-layout.md) — supervisor layout; deferred SSE hub ports
- [docs/domain/sse-hub.md](../domain/sse-hub.md) — behavioral reference
