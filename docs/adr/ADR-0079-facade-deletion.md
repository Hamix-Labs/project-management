# ADR-0079: Delete pkgs/tasks/store monolithic facade

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Tier 5 split task persistence into bounded contexts. After [ADR-0078](./ADR-0078-facade-alias-retirement.md) removed facade type aliases, `pkgs/tasks/store` still delegated every BC method through fourteen `facade_*.go` files plus cross-cutting ready-task notify and pickup-wake hooks. New teams had to read the facade to understand taskapi wiring.

## Decision

1. **Delete** `pkgs/tasks/store` (facade implementation, `store/model` migrate hub, `internal/notify`).
2. **Add** `internal/taskapi/composition` — constructs BC `*store.Store` values, implements `pkgs/tasks/wire.HandlerAPI` and `pkgs/agents/worker.Store`, owns task CRUD notify/pickup-wake side effects.
3. **Add** `internal/taskapi/storehooks` — `ReadyTaskNotifier`, `PickupWake`, task git-context resolution, and `ApplyNotifyDecision` helpers.
4. **Wire** `cmd/taskapi` through `composition.NewAPI(db)`; agent worker supervisor and HTTP handler take `*composition.API`.
5. **Move** GORM AutoMigrate model list to `pkgs/tasks/postgres/migrate_models.go` (BC-local `store/model/migrate_models.go` remain authoritative per BC).
6. **Relocate** facade integration tests to `internal/taskapi/composition` or BC store packages before deletion.

## Consequences

### Positive

- Single composition root for taskapi; BC stores are the only persistence implementations.
- Cross-cutting hooks live beside the binary, not in a importable `pkgs/` facade.
- `pkgs/tasks/store/model` migrate hub removed — postgres orchestration calls BC models directly.

### Negative / Trade-offs

- `internal/taskapi/composition` still delegates many methods (moved from facade files) until further slimming.
- Tests that used `store.NewStore` now use `composition.NewAPI`.

## Alternatives

- **Keep facade as tombstone package** — rejected; prolongs dual navigation with BC stores.
- **Embed BC stores on one struct** — rejected; Go embedding conflicts on `Store` type name and `DB()` across BCs.

## See also

- [ADR-0078](./ADR-0078-facade-alias-retirement.md) — alias retirement prerequisite
- [ADR-0070](./ADR-0070-taskapi-shell-ownership.md) — taskapi shell ownership
- `internal/taskapi/composition` — new wiring root
