# Audit index

Phase 1 (boundaries) and Phase 2 (dead-code) ROI audits are **complete** — reports retired 2026-07-12. Active work: [cleanup-order.md](../cleanup-order.md).

## Phase 3 — policy centralization (complete)

Audit: [policy-roi.md](./policy-roi.md). Shipped [#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203): task invalidation catalog, read limits, SSE publish choke, sync invalidation dedup, CI gates.

## Phase 5 — structural patterns (in progress)

Audit: [structural-patterns-roi.md](./structural-patterns-roi.md). PR train #1–#10: web god-file splits → web tests → Go handlers/store → handler/agents tests → CI warn gate.

## Structural themes (in progress)

| Theme | Status | Next |
| --- | --- | --- |
| **Policy choke points** | done | [policy-roi.md](./policy-roi.md) — ADR-0080, readpolicy, writepolicy SSE choke, sync catalog |
| **Web god-file splits** | in progress | [structural-patterns-roi.md](./structural-patterns-roi.md) PR2–4 — cycle attempt → create modal → task detail |
| **Go handler / store splits** | not started | [structural-patterns-roi.md](./structural-patterns-roi.md) PR6–7 — BC handlers, storefake, store internal |
| **Handler test consolidation** | partial | [structural-patterns-roi.md](./structural-patterns-roi.md) PR5, PR8–9 — web + handler + agents test splits |

Completed themes (no active backlog): git vertical cleanup, web vertical boundaries, frontend mutation policy, store/handler DIP, select/widget unification.
