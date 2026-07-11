# ADR-0051: Extract `pkgs/taskchecklist` bounded context

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Task checklist persistence and HTTP routes lived in `pkgs/tasks/store/internal/checklist` and `pkgs/tasks/handler/handler_checklist.go`. Checklist completion can unblock task dependents — that orchestration belongs in the tasks composition root, not inside the checklist store. Prior extractions (taskcompose, settings, gitinventory) established the embed-and-delegate pattern with a dedicated handler `Register` and CI import gates.

## Decision

1. **New bounded context** — `pkgs/taskchecklist/{contract,store,handler}` owns checklist wire types, GORM persistence, and HTTP routes (`GET/POST/PATCH/DELETE` under `/tasks/{id}/checklist*`).

2. **Composition root unchanged** — `pkgs/tasks/store.Store` embeds `pkgs/taskchecklist/store.Store` and delegates through `facade_checklist.go`. `notifyUnblockedDependents` runs in the tasks facade after completion mutations that set `criteria_satisfied_at`.

3. **Route registration** — `pkgs/taskchecklist/handler.Register` mounts routes from `pkgs/tasks/handler/handler_routes.go`. No path or JSON shape changes.

4. **Contract kernel** — `pkgs/tasks/contract` keeps `ChecklistStore` and wire types aliased from `pkgs/taskchecklist/contract` so bootstrap, harness fakes, and handler tests compile unchanged.

5. **Cross-domain InTx helpers** — `ValidateCanMarkDoneInTx`, `ValidateCanAddCriterionInTx`, and `DeleteOwnedItemsInTx` are exported from `pkgs/taskchecklist/store` for `pkgs/tasks/store/internal/{tasks,devmirror}`.

6. **Shared models** — GORM models for checklist tables remain in `pkgs/tasks/store/model`; domain types remain in `pkgs/tasks/domain`.

7. **Import gate** — CI rejects `pkgs/taskchecklist` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (`scripts/check-go.sh` → `step_taskchecklist_boundary`).

8. **SSE after mutations** — Checklist handler receives `NotifyTaskUpdated func(ctx, taskID) error` from tasks handler (`notifyTaskUpdatedEnriched`).

## Consequences

### Positive

- Checklist ownership is explicit; tasks store internals shrink.
- Dependent-unblock side effects stay documented at the tasks facade boundary.

### Negative / trade-offs

- Checklist store still depends on `pkgs/tasks/domain`, `pkgs/tasks/store/model`, and `pkgs/storekernel` for task rows, events, and shared persistence helpers.
- Cross-route HTTP contract tests remain in `pkgs/tasks/handler` because they exercise the full taskapi mux.

## See also

- [pkgs/taskchecklist/README.md](../../pkgs/taskchecklist/README.md)
- [ADR-0048](./ADR-0048-bounded-context-taskcompose.md)
