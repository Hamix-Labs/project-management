# ADR-0061: Harness internal contract BC slices

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/agents/harness/internal/contract/store.go` was a single 72-line interface over `tasks/store` facade input types and monolith naming. Harness subpackages (`verify`, `resume`, `git`) shared one god-interface.

## Decision

1. **Slice files** — `internal/contract/{task,cycle,checklist,snapshots,settings,events,projects}.go` each own one persistence concern.
2. **BC input types** — harness contracts import `taskcore/contract`, `taskcycles/contract` + `taskcycles/store`, `taskchecklist/contract` + `taskchecklist/store`, `projects/store`, `settings/contract` directly — not `pkgs/tasks/store` aliases.
3. **Composed Store** — `store.go` embeds slices only; `harness.Store` remains the public alias.
4. **storefake** — continues to wrap `*tasks/store.Store`; compile-time `contract.Store` assert unchanged.

## Consequences

### Positive

- Harness contract reads like BC composition; easier to narrow fakes per subpackage.
- Input types trace to BC owners in code review.

### Negative / Trade-offs

- Harness-only methods (e.g. `ListCyclesForTask`, `ListChecklistForVerify`) stay on `CycleStore` / `ChecklistStore` until BC contracts absorb them.

## See also

- [docs/domain/harness.md](../domain/harness.md)
