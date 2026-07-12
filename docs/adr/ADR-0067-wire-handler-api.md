# ADR-0067: Canonical wire.HandlerAPI

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/tasks/handler/handler_store.go` and `pkgs/tasks/store/handler_api.go` each duplicated a large composed persistence interface.

## Decision

1. **`pkgs/tasks/wire.HandlerAPI`** is the single composed contract for taskapi wiring.
2. Handler and store packages expose type aliases only (`HandlerStore`, `HandlerAPI`).
3. CI step `tasks wire handler API` greps for duplicate interface blocks outside `pkgs/tasks/wire`.

## Consequences

### Positive

- One edit surface when adding a BC contract slice to the HTTP facade.
- BC handlers never import `pkgs/tasks/handler` for store types.

### Negative / Trade-offs

- `wire` is a wiring-edge package — not a domain hub; no business logic.

## See also

- Tier 5A PR3 — [cleanup-order.md](../cleanup-order.md)
