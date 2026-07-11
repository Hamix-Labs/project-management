# ADR-0062: Retire pkgs/tasks/contract hub

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/tasks/contract` re-exported BC store slices (`CycleStore = cyclescontract.CycleStore`, etc.) for `HandlerStore` composition. It was a second alias hub after Tier 3 contract colocation ([ADR-0055](./ADR-0055-contract-colocation.md)).

## Decision

1. **Handler composition** — `pkgs/tasks/handler/handler_store.go` composes `HandlerStore` from BC `contract` packages directly.
2. **Store assert** — `pkgs/tasks/store/handler_api.go` duplicates the same BC composition as `HandlerAPI` (store must not import handler).
3. **Delete** `pkgs/tasks/contract/` including alias files and hub tests.
4. **Callers** — `taskcycles/handler`, `runners/handler`, `handler/storefake` import BC contracts or `handler.HandlerStore`.

## Consequences

### Positive

- One less indirection layer between HTTP handlers and BC contracts.
- `HandlerStore` ownership is explicit at the composition root.

### Negative / Trade-offs

- `HandlerAPI` and `HandlerStore` are duplicate interface compositions (acceptable: dependency direction).

## See also

- [pkgs/tasks/handler/README.md](../../pkgs/tasks/handler/README.md)
