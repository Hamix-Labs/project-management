# ADR-0066: Bootstrap on BC contracts

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/tasks/service/bootstrap.go` typed cold-start reads against `store.AppSettings`, `store.ListFilter`, `store.TaskStats`, and `store.DraftSummary` aliases.

## Decision

`BootstrapStore` and `BootstrapData` use bounded-context contract/domain types directly (`settingsdomain`, `taskcorecontract`, `composecontract`, `projectsdomain`). `*store.Store` still satisfies the interface at the wiring edge.

## Consequences

### Positive

- Service layer no longer imports `pkgs/tasks/store` for types.
- Pattern for retiring facade type aliases.

### Negative / Trade-offs

- Handler wire mapping unchanged; only service boundary moves.
