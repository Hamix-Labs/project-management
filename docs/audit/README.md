# Audit index

Phase 1 (boundaries) and Phase 2 (dead-code) ROI audits are **complete** — reports retired 2026-07-12 ([#198](https://github.com/AlexsanderHamir/Hamix/pull/198)). Active work: [cleanup-order.md](../cleanup-order.md).

## Phase 3 — policy centralization (complete)

Audit: [policy-roi.md](./policy-roi.md). Shipped [#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203): task invalidation catalog, read limits, SSE publish choke, sync invalidation dedup, CI gates.

## Phase 5 — structural patterns (complete)

Audit: [structural-patterns-roi.md](./structural-patterns-roi.md). Shipped [#205](https://github.com/AlexsanderHamir/Hamix/pull/205)–[#215](https://github.com/AlexsanderHamir/Hamix/pull/215): web god-file splits → web tests → Go handlers/store → handler/agents tests → CI warn gate.

## Structural themes

| Theme | Status | Reference |
| --- | --- | --- |
| **Policy choke points** | done | [policy-roi.md](./policy-roi.md) — ADR-0080, readpolicy, writepolicy SSE choke, sync catalog |
| **Web god-file splits** | done | [structural-patterns-roi.md](./structural-patterns-roi.md) — cycle attempt, create modal, task detail |
| **Go handler / store splits** | done | [structural-patterns-roi.md](./structural-patterns-roi.md) — BC handlers, storefake, store internal |
| **Handler test consolidation** | done | [structural-patterns-roi.md](./structural-patterns-roi.md) — web + handler + agents test splits |

Completed themes (no active backlog): git vertical cleanup, web vertical boundaries, frontend mutation policy, store/handler DIP, select/widget unification.
