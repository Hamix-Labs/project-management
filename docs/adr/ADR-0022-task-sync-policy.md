# ADR-0022: Task Sync Policy (Frontend Decide vs Apply)

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-19
**Status:** Accepted
**Deciders:** Engineering

## Context

[ADR-0020](ADR-0020-realtime-sse-layout.md) split backend SSE transport; the frontend invalidation path remained imperative across [`sseCacheBridge.ts`](../../web/src/tasks/hooks/sseCacheBridge.ts), [`optimisticVersion.ts`](../../web/src/tasks/hooks/optimisticVersion.ts), [`useTaskEventStream.ts`](../../web/src/tasks/hooks/useTaskEventStream.ts), and [`queryClient.ts`](../../web/src/lib/queryClient.ts). A prior hook decomposition improved file size but not policy ownership.

Contributors answering "what happens on `task_cycle_changed` while PATCH is in flight?" must trace multiple modules. Flush rules (enrichment skip, detail-prefix invalidation) live only in integration tests.

## Decision

Introduce **`web/src/tasks/sync/`** as the frontend cache-coherence layer using **Decide → Apply** (mirroring [ADR-0021](ADR-0021-harness-execute-orchestration.md)).

| Module | Role |
|--------|------|
| `decideSyncFrame` | Pure: parsed frame + mutation guard → schedule, pending delta, effects |
| `decideFlushBatch` | Pure: pending snapshot → invalidate query keys |
| `applySyncEffects` | Impure: `setQueryData`, `invalidateQueries`, RUM, progress push |
| `taskSyncCoordinator` | Holds pending state, debounce timers, wires Decide + Apply |
| `mutationGuard` | Per-task optimistic version counter (SSE echo suppression) |
| `connectionPolicy` | SSE live flag for `refetchOnWindowFocus` |

**Boundary rules:**

- [`parseTaskChangeFrame`](../../web/src/tasks/task-query/sseInvalidate.ts) stays wire-decode only; no QueryClient.
- `sync/` must not import React hooks or components.
- [`useTaskEventStream`](../../web/src/tasks/hooks/useTaskEventStream.ts) remains the sole app-wide EventSource mount; delegates to coordinator.
- Mutation hooks use public API `beginTaskMutation` / `endTaskMutation` from `sync/index.ts`.

Timing constants unchanged: **900ms / 2500ms** task flush; **5000ms / 10000ms** progress stream flush.

### Non-goals (this ADR)

- SSE wire protocol or backend publish changes
- Extending mutation guard to checklist/bulk/create flows
- React Context for sync state
- Unified mutation optimistic helper for all flows

## Consequences

### Positive

- Frame and flush policy are table-tested without EventSource or React.
- One module answers cache-coherence questions for task realtime UX.
- Docs align with backend realtime layout (ADR-0020).

### Enrichment coverage (flush rules)

- **E1 — Task-row only:** `patch_task_detail` / `enrichedTasks` means the task *row* cache is current. It must not suppress invalidation of sibling detail keys (cycles list/detail, checklist, events) that were not patched in the same apply step. Checklist is owned by `pending.tasks` (`task_updated`): `decideFlushBatch` always invalidates `checklist(taskId)` for each pending task id, including when enrichment skips `detailRoot`.
- **E2 — Cycle hints are independent:** `decideFlushBatch` invalidates `cycles(taskId)` from pending cycle entries regardless of whether the same task id is also in `pending.tasks` / enriched. Coupling “task pending ⇒ skip cycles” assumed `detailRoot` would always run; enrichment made that false. Cycle pending does **not** invalidate checklist (phase ledger ≠ checklist).
- Harness verified checklist completions publish `task_updated` (aligned with HTTP checklist / ADR-0026) so checklist refresh is driven by task frames, not cycle side effects. Cycle frames queue cycle pending only (no `addTaskId`) so verify-end `task_cycle_changed` cannot refetch checklist before completions commit.

### Negative / Trade-offs

- Temporary shims during migration (`optimisticVersion` re-exports)
- Coordinator still owns timer wiring (transport concern adjacent to policy)
- Task-row status during an agent run relies on harness `task_updated` publishes (not cycle frames alone)

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| React Context for guard state | Forces mutation hooks to subscribe; module map is smaller |
| Merge parser into sync | Wire decode belongs with task-query keys |
| Big-bang rewrite without shims | Higher regression risk vs incremental delegate |

## Related

- [ADR-0020](ADR-0020-realtime-sse-layout.md) — backend SSE layout; deferred frontend track
- [docs/domain/sse-hub.md](../domain/sse-hub.md) — event catalog and SPA behavior
- [docs/web.md](../web.md) — sync module read order
