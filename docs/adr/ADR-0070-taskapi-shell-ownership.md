# ADR-0070: Taskapi shell ownership

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

After Tier 4 BC extraction, `cmd/taskapi` and `internal/taskapi` remain the permanent assembly shell — but package docs still referenced deleted monolithic handler files.

## Decision

1. **`cmd/taskapi`** wires config, Postgres, `*store.Store`, agent worker supervisor, and HTTP listen/shutdown only.
2. **`internal/taskapi`** owns `NewHTTPHandler` — middleware stack + `handler.NewHandler` options (repo provider, schema drift, agent control).
3. **`pkgs/tasks/handler`** is the REST/SSE mux; routes live in `handler_routes.go` and BC-delegating handlers — not a growing god package.
4. Realtime SSE hub lives in `pkgs/tasks/realtime` ([ADR-0080](ADR-0080-sse-hub-realtime-ownership.md)); handler keeps HTTP stream + notify glue.

## Consequences

### Positive

- Clear ownership map for contributors landing full-stack features.
- Shell docs match post–Tier 4 file layout.

### Negative / Trade-offs

- Composition shell ownership matches post–Tier 4 layout; SSE hub ownership is documented in ADR-0080.

## See also

- [docs/architecture.md](../architecture.md)
