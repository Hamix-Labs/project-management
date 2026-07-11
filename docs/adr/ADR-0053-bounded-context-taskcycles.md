# ADR-0053: Extract `pkgs/taskcycles` bounded context (store PR-3a, handler PR-3b)

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Execution cycle persistence — cycles, phases, stream events, criteria/verify reports, command runs, and indexed commits — lived in `pkgs/tasks/store/internal/{cycles,reports,commits}`. Prior extractions (taskcompose, taskchecklist, settings, gitinventory) established the embed-and-delegate pattern with dedicated contract packages and CI import gates. Cycle HTTP handlers were tightly coupled to the tasks mux; splitting store first (PR-3a) reduced review risk, then handler extraction (PR-3b) completed the bounded context.

## Decision

1. **New bounded context** — `pkgs/taskcycles/{contract,store,handler}` owns the `CycleStore` interface, cycle input types, GORM persistence under `store/internal/{cycles,reports,commits}`, and cycle/commit/cycle-failures HTTP routes under `handler/`.

2. **Composition root unchanged** — `pkgs/tasks/store.Store` embeds `pkgs/taskcycles/store.Store` as `cycles` and delegates through existing `facade_cycles.go`, `facade_commits.go`, and `facade_reports.go`. Public method signatures and harness call sites stay stable.

3. **Contract kernel** — `pkgs/tasks/contract` keeps `CycleStore`, `StartCycleInput`, and `CompletePhaseInput` aliased from `pkgs/taskcycles/contract` so bootstrap, harness fakes, and handler tests compile unchanged.

4. **Cross-domain helpers** — `FailureSurfaceMessage` is exported from `pkgs/taskcycles/store` for `pkgs/tasks/store/internal/stats` (recent_failures projection).

5. **Shared models** — GORM models for cycle tables remain in `pkgs/tasks/store/model`; domain types remain in `pkgs/tasks/domain`.

6. **Import gate** — CI rejects `pkgs/taskcycles` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (`scripts/check-go.sh` → `step_taskcycles_boundary`).

7. **Handler wiring (PR-3b)** — `pkgs/taskcycles/handler.Register` mounts cycle, commit, and `/tasks/cycle-failures` routes from `pkgs/tasks/handler/handler_routes.go`. Enriched `cycle_changed` SSE payloads are built in the taskcycles handler; the tasks mux supplies the hub publish callback. Full-mux contract tests remain in `pkgs/tasks/handler`.

## Consequences

### Positive

- Cycle persistence and HTTP ownership are explicit; tasks handler internals shrink.

### Negative / trade-offs

- Cycle store still depends on `pkgs/tasks/domain`, `pkgs/tasks/store/model`, and `pkgs/storekernel`.
- Full-mux contract tests remain in `pkgs/tasks/handler` by design.

## See also

- [pkgs/taskcycles/README.md](../../pkgs/taskcycles/README.md)
- [ADR-0051](./ADR-0051-bounded-context-taskchecklist.md)
- [ADR-0048](./ADR-0048-bounded-context-taskcompose.md)
