# ADR-0058: Task cycles domain and store model ownership

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

[ADR-0053](./ADR-0053-bounded-context-taskcycles.md) extracted cycle HTTP and store facades but left cycle **domain types** in `pkgs/tasks/domain` and **GORM models** in `pkgs/tasks/store/model`. That split contradicted the four-layer BC blueprint established by `pkgs/taskchecklist/` and `pkgs/taskevents/` (local `domain/` + `store/model/`).

Harness and the tasks facade still import cycle shapes through `pkgs/tasks/domain` during the Tier 3 migration window.

## Decision

1. **Domain ownership** — `TaskCycle`, phases, stream events, verdict reports, command runs, commits, cycle enums, and state-machine helpers live in `pkgs/taskcycles/domain/`. `TaskCycle.TriggeredBy` is a string holding `taskevents/domain.Actor` wire values (no import cycle).

2. **Model ownership** — GORM structs and domain↔model mappers for cycle tables live in `pkgs/taskcycles/store/model/`.

3. **Compatibility aliases** — `pkgs/tasks/domain/cycle_aliases.go` re-exports cycle types, enums, and helper functions so harness and existing imports compile unchanged.

4. **Migrate hub** — `pkgs/tasks/store/model/migrate_models.go` registers `taskcycles/store/model` types in FK-safe order alongside other BC models.

5. **Parity registry** — Cycle `ParityPairs` entries in `pkgs/tasks/store/model/parity.go` reference `taskcycles/domain` + `taskcycles/store/model` (schema guards stay centralized until a dedicated cycles parity package is warranted).

6. **Import gates** — `scripts/check-go.sh` → `step_taskcycles_boundary` forbids GORM and `pkgs/tasks/domain` imports in `taskcycles/domain`, plus handler/store/internal imports from `pkgs/tasks/handler` and `pkgs/tasks/store/internal`.

7. **Unchanged** — Cycle HTTP routes/JSON, `FailureSurfaceMessage` export from `taskcycles/store`, `TaskContextSnapshot` domain (stays in tasks until taskcore #186). Store/handler may import `pkgs/tasks/domain` for `Actor`, task rows, and shared errors.

## Consequences

### Positive

- Taskcycles BC matches taskchecklist/taskevents navigation and dependency rules.
- Domain layer is stdlib + `taskchecklist/domain` (VerifierKind only); persistence tags stay in `store/model`.

### Negative / trade-offs

- Temporary duplicate type names at `tasks/domain` alias layer until taskcore (#186).
- Cycle models FK-reference `tasks/store/model.Task` and `taskchecklist/store/model.TaskChecklistItem`.

## See also

- [pkgs/taskcycles/README.md](../../pkgs/taskcycles/README.md)
- [ADR-0053](./ADR-0053-bounded-context-taskcycles.md)
- [ADR-0039](./ADR-0039-domain-persistence-separation.md)
