# ADR-0057: Task events domain and store model ownership

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

[ADR-0054](./ADR-0054-bounded-context-taskevents.md) extracted taskevents HTTP and store facades but left **domain types** in `pkgs/tasks/domain` and **GORM models** in `pkgs/tasks/store/model`. That split contradicted the four-layer BC blueprint established by `pkgs/taskchecklist/` ([ADR-0056](./ADR-0056-taskchecklist-domain-model.md)).

Harness, storekernel, and tasks facade still import event shapes through `pkgs/tasks/domain` during the Tier 3 migration window.

## Decision

1. **Domain ownership** — `TaskEvent`, `EventType`, `ResponseThreadEntry`, `EventTypeAcceptsUserResponse`, and `Actor` (event/cycle wire enum) live in `pkgs/taskevents/domain/`. The domain package imports stdlib only.

2. **Model ownership** — GORM structs and domain↔model mappers for `task_events` live in `pkgs/taskevents/store/model/`.

3. **Compatibility aliases** — `pkgs/tasks/domain` re-exports event and actor symbols as type/const aliases so harness and existing imports compile unchanged.

4. **Migrate hub** — `pkgs/tasks/store/model/migrate_models.go` registers `taskeventsmodel.TaskEvent` in FK-safe order alongside other BC models.

5. **Parity registry** — `TaskEvent` `ParityPairs` entry in `pkgs/tasks/store/model/parity.go` references `taskeventsdomain` + `taskeventsmodel`.

6. **Import gates** — `scripts/check-go.sh` → `step_taskevents_boundary` forbids GORM and `pkgs/tasks/domain` imports in `taskevents/domain`.

7. **Unchanged** — Event HTTP routes/JSON, append dual-write via `storekernel.AppendEvent` from CRUD/cycles/checklist, SSE after PATCH.

## Consequences

### Positive

- Taskevents BC matches taskchecklist four-layer navigation and dependency rules.
- Domain layer is stdlib-only; persistence tags stay in `store/model`.

### Negative / trade-offs

- Temporary duplicate type names at `tasks/domain` alias layer until taskcore (#186).
- `storekernel` and tasks store stats paths still reference `taskeventsmodel.TaskEvent` for raw GORM queries.

## See also

- [pkgs/taskevents/README.md](../../pkgs/taskevents/README.md)
- [ADR-0054](./ADR-0054-bounded-context-taskevents.md)
- [ADR-0039](./ADR-0039-domain-persistence-separation.md)
