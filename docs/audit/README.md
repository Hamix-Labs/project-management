# Audit index

Phase 1 (boundaries) and Phase 2 (dead-code) ROI audits are **complete** — reports retired 2026-07-12 ([#198](https://github.com/AlexsanderHamir/Hamix/pull/198)). Active work: [cleanup-order.md](../cleanup-order.md) + [remaining-cleanup-roi.md](./remaining-cleanup-roi.md).

## Phase 3 — policy centralization (complete)

Audit: [policy-roi.md](./policy-roi.md). Shipped [#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203): task invalidation catalog, read limits, SSE publish choke, sync invalidation dedup, CI gates.

## Phase 4 — targeted deduplication (in progress)

Audit: [remaining-cleanup-roi.md](./remaining-cleanup-roi.md) (Phase 4 section). Third-occurrence DRY into `handlerhttp` (bounded limit, path IDs, httplog, invalid-input) — PR train after this audit merges.

## Phase 5 — structural patterns (complete)

Audit: [structural-patterns-roi.md](./structural-patterns-roi.md). Shipped [#205](https://github.com/AlexsanderHamir/Hamix/pull/205)–[#215](https://github.com/AlexsanderHamir/Hamix/pull/215): web god-file splits → web tests → Go handlers/store → handler/agents tests → CI warn gate.

## Phase 6 — abstractions ≥2 real impls (in progress)

Audit: [remaining-cleanup-roi.md](./remaining-cleanup-roi.md) (Phase 6 section). Repo-scoped query factory, TaskGetter unify, AgentWorkerControl, notify/PathMap, PickupWake — same PR train as Phase 4.

## Structural themes

| Theme | Status | Reference |
| --- | --- | --- |
| **Policy choke points** | done | [policy-roi.md](./policy-roi.md) — ADR-0080, readpolicy, writepolicy SSE choke, sync catalog |
| **Web god-file splits** | done | [structural-patterns-roi.md](./structural-patterns-roi.md) — cycle attempt, create modal, task detail |
| **Go handler / store splits** | done | [structural-patterns-roi.md](./structural-patterns-roi.md) — BC handlers, storefake, store internal |
| **Handler test consolidation** | done | [structural-patterns-roi.md](./structural-patterns-roi.md) — web + handler + agents test splits |
| **Third-occurrence DRY** | in progress | [remaining-cleanup-roi.md](./remaining-cleanup-roi.md) — Phase 4 |
| **Late abstractions** | in progress | [remaining-cleanup-roi.md](./remaining-cleanup-roi.md) — Phase 6 |

Completed themes (no active backlog): git vertical cleanup, web vertical boundaries, frontend mutation policy, store/handler DIP, select/widget unification.
