# Audit index

Phase 1 (boundaries) and Phase 2 (dead-code) ROI audits are **complete** — reports retired 2026-07-12. Active work: [cleanup-order.md](../cleanup-order.md).

## Phase 3 — policy centralization (in progress)

Backlog: [policy-roi.md](./policy-roi.md). Implementation PR train: task invalidation catalog → read limits → SSE publish choke → sync dedup.

## Structural themes (in progress)

| Theme | Status | Next |
| --- | --- | --- |
| **Policy choke points** | in progress | [policy-roi.md](./policy-roi.md) #1 task invalidation catalog |
| **Web god-file splits** | in progress | Cycle detail page → create modal → task detail page |
| **Go handler / store splits** | not started | Files past CODE_STANDARDS red limits — [handler README](../../pkgs/tasks/handler/README.md) |
| **Handler test consolidation** | partial | Contract harness done; extend fakes / `tasktestserver` patterns |

Completed themes (no active backlog): git vertical cleanup, web vertical boundaries, frontend mutation policy, store/handler DIP, select/widget unification.
