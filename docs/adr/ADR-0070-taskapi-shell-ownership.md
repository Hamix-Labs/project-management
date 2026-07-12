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
4. Realtime (SSE hub) stays in `pkgs/tasks/handler` for Tier 5; optional `pkgs/tasks/realtime` extract deferred until a second consumer exists.

## Consequences

### Positive

- Clear ownership map for contributors landing full-stack features.
- Shell docs match post–Tier 4 file layout.

### Negative / Trade-offs

- SSE hub still colocated with HTTP handlers until realtime extract PR.

## See also

- [docs/architecture.md](../architecture.md)
