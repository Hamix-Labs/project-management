# ADR-0068: Harness BC contract alignment

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/agents/harness/internal/contract` duplicated cycle and checklist harness store shapes that belonged in bounded-context contracts.

## Decision

1. **`pkgs/taskcycles/contract`** owns `CycleHarnessStore`, `CycleWorkerStore`, and harness entry types (criteria/verify reports, command runs, commits, stream events).
2. **`pkgs/taskchecklist/contract`** owns `ChecklistHarnessStore` and `ChecklistVerifyItem`.
3. Harness internal contract aliases those slices; `*cyclesstore.Store` and `*checkliststore.Store` satisfy them via type aliases and converters.

## Consequences

### Positive

- Harness contract is composition of BC contracts plus harness-only slices (task, events, settings).
- No parallel type tree in harness internal packages.

### Negative / Trade-offs

- Facade `tasks/store` still delegates to BC stores until Tier 6 facade retirement.

## See also

- [ADR-0061](./ADR-0061-harness-contract-slices.md)
