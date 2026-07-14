# ADR-0080: SSE hub ownership in `realtime`

**Date:** 2026-07-13  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

[ADR-0020](ADR-0020-realtime-sse-layout.md) split SSE wire/coalesce into `pkgs/tasks/realtime` but kept the ring-buffer hub in `pkgs/tasks/handler`. [ADR-0070](ADR-0070-taskapi-shell-ownership.md) deferred moving the hub until a second consumer existed. `agentworker` already depends on `realtime.Publisher`, and handler whitebox + handlertest themes now exercise the hub without needing unexported `subscribe`.

## Decision

1. Move **`SSEHub`** (construction, ring buffer, `Publish`, `Subscribe`, `SubscribeSince`, eviction, coalescing) into **`pkgs/tasks/realtime`**.
2. Keep **HTTP framing** (`streamEvents`, frame writers) and **writepolicy notify glue** in `pkgs/tasks/handler`.
3. Export **`SubscribeSince`** + **`BufferedEvent`** / **`Subscriber`** so the stream handler no longer needs unexported hub APIs.
4. `cmd/taskapi` and `internal/taskapi` construct **`realtime.SSEHub`**. Handler retains type aliases (`type SSEHub = realtime.SSEHub`, `NewSSEHub`, …) for same-package tests.
5. Amend ADR-0020 dependency rule: `realtime` may import `middleware` + `calltrace` for SSE metrics (not stdlib-only).

```
realtime     → stdlib, middleware (SSE metrics), calltrace
handler      → realtime (hub + wire), middleware, calltrace
agentworker  → realtime.Publisher (not concrete hub required)
cmd/taskapi  → realtime (concrete hub), handler (HTTP)
```

## Consequences

### Positive

- Composition shell no longer digs for hub implementation beside route handlers.
- Hub whitebox tests live next to the hub (`realtime/*_test.go`).
- Second consumer (agentworker/`Publisher`) matches the deferred ADR-0070 extract bar.

### Negative / Trade-offs

- `realtime` is no longer stdlib-only (metrics + slog path names).
- Handler aliases preserve `handler.NewSSEHub` for test churn reduction; prefer `realtime.*` at the binary edge.

## Related

- [ADR-0020](ADR-0020-realtime-sse-layout.md) — initial split  
- [ADR-0070](ADR-0070-taskapi-shell-ownership.md) — deferred hub note (superseded for hub placement)  
- [docs/domain/sse-hub.md](../domain/sse-hub.md)  
- Phase 7 PR5 (tasks shell shape)
