# ADR-0054: Extract `pkgs/taskevents` bounded context

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Task audit event reads, keyset paging, approval-pending probes, and response-thread append lived in `pkgs/tasks/store/internal/events` and `pkgs/tasks/handler/handler_task_events.go`. Event **append** during CRUD, cycles, checklist, and devmirror mutations already flows through `pkgs/storekernel` (`NextEventSeq`, `AppendEvent`) inside those subpackages — not through the events internal package. Prior extractions (taskchecklist, taskcycles) established the embed-and-delegate pattern with a dedicated handler `Register` and CI import gates.

## Decision

1. **New bounded context** — `pkgs/taskevents/{contract,store,handler}` owns event wire contracts, GORM read/thread paths, and HTTP routes (`GET/PATCH` under `/tasks/{id}/events*`).

2. **Composition root unchanged** — `pkgs/tasks/store.Store` embeds `pkgs/taskevents/store.Store` and delegates through `facade_events.go`. `ThreadEntriesForDisplay` remains re-exported at `pkgs/tasks/store` for devsim and legacy callers.

3. **Route registration** — `pkgs/taskevents/handler.Register` mounts routes from `pkgs/tasks/handler/handler_routes.go`. No path or JSON shape changes.

4. **Contract kernel** — `pkgs/tasks/contract` keeps `TaskEventStore` and `TaskEventsPage` aliased from `pkgs/taskevents/contract` so bootstrap, harness fakes, and handler tests compile unchanged.

5. **Append stays split** — Cross-subpackage dual-writes continue via `storekernel` in task CRUD, cycles, checklist, and devmirror. `taskevents/store` exposes `AppendTaskEvent` for standalone append and tests; it does not replace in-transaction kernel append paths.

6. **Shared models** — GORM models for `task_events` remain in `pkgs/tasks/store/model`; domain types remain in `pkgs/tasks/domain`.

7. **Import gate** — CI rejects `pkgs/taskevents` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (`scripts/check-go.sh` → `step_taskevents_boundary`).

8. **SSE after PATCH** — Events handler receives `NotifyTaskEventChanged func(taskID, eventSeq)` from tasks handler (`notifyTaskEventChanged`).

## Consequences

### Positive

- Event read ownership is explicit; tasks store internals shrink.
- Append invariants stay centralized in `storekernel` for dual-write paths.

### Negative / trade-offs

- Taskevents store still depends on `pkgs/tasks/domain`, `pkgs/tasks/store/model`, and `pkgs/storekernel`.
- Cross-route HTTP contract tests remain in `pkgs/tasks/handler` because they exercise the full taskapi mux.

## See also

- [pkgs/taskevents/README.md](../../pkgs/taskevents/README.md)
- [ADR-0051](./ADR-0051-bounded-context-taskchecklist.md)
