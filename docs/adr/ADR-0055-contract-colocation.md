# ADR-0055: Colocate BC contracts under `pkgs/{bc}/contract`

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Tier 2 extracted handler and store packages for checklist, cycles, events, and runners. Mature BCs (projects, settings, gitinventory) and taskcompose already had local `domain/` and `store/`, but **store interfaces and wire DTOs** for projects, settings, and git still lived in `pkgs/tasks/contract/`. That split made the four-layer blueprint (`domain` / `contract` / `store` / `handler`) hard to discover — contributors had to read two packages to understand one BC.

taskcompose already colocates `contract/` next to `domain/` and is the reference layout.

## Decision

1. **Local contract packages** — Move `ProjectStore`, `SettingsStore`, `GitReadStore`, and `GitWriteStore` (plus wire input types) into:
   - `pkgs/projects/contract/`
   - `pkgs/settings/contract/`
   - `pkgs/gitinventory/contract/`

2. **Alias hub unchanged** — `pkgs/tasks/contract` re-exports each interface and type as a one-line alias so `HandlerStore` composition, harness, and existing handler tests compile without import churn.

3. **BC internals** — Each BC's `handler/` and `store/` import their local `contract/` package (not `pkgs/tasks/contract`).

4. **No API change** — REST paths and JSON shapes unchanged.

5. **CI** — `pkgs/{bc}/contract` must not import `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (added in gitinventory PR tail commit).

## Consequences

### Positive

- Every mature BC matches the taskcompose navigation model.
- Contract ownership is obvious from folder layout.

### Negative / trade-offs

- Temporary duplication of type names at `tasks/contract` alias layer until a future hub slim-down.
- Three small PRs instead of one (reviewability).

## See also

- [tier_3_bc_blueprint plan](../../.cursor/plans/tier_3_bc_blueprint_5ff0fc21.plan.md)
- [ADR-0048](./ADR-0048-bounded-context-taskcompose.md)
