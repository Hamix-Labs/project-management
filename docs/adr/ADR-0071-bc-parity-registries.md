# ADR-0071: BC-local parity registries (ADR-0039 phase)

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/tasks/store/model/parity.go` centralized all domain↔model parity pairs — coupling every BC schema test to the tasks facade hub.

## Decision

1. Each BC **`store/model/parity.go`** owns its `ParityPairs` registry (pattern: `settings`, `gitinventory`, `taskcycles`).
2. **`pkgs/tasks/store/model`** retains the aggregate `ParityPairs` for unified migrate/schema tests until hub deletion; entries are removed as BC registries land.
3. BC packages copy `parity_helpers.go` + `field_parity_test.go` locally (same pattern as gitinventory).

## Consequences

### Positive

- ADR-0039 decentralization progresses without blocking on facade removal.
- BC teams run parity tests in their package.

### Negative / Trade-offs

- Temporary duplication of parity helper helpers until a shared `storekernel/parity` extract is justified.

## See also

- [ADR-0039](./ADR-0039-domain-persistence-separation.md)
