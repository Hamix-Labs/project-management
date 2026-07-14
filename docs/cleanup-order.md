# Cleanup order (agent guide)

**When:** Refactor / deletion / hardening — not new product surface.  
**Scope:** Architectural, organizational, and testing improvements only (boundaries, file layout, policy choke points, contract tests, god-file splits). Operator UX, harness reliability features, and new product APIs are **out of scope** until this pass ends.  
**Bar:** One slice per PR; `.\scripts\check.ps1` green.

**Backlog (pick rows; this doc is order only):**

| Index | Use for |
| --- | --- |
| [audit/README.md](./audit/README.md) | Active structural themes |
| [audit/policy-roi.md](./audit/policy-roi.md) | Phase 3 policy centralization (complete) |
| [audit/structural-patterns-roi.md](./audit/structural-patterns-roi.md) | Phase 5 god-files, handler/store splits (complete) |
| [audit/remaining-cleanup-roi.md](./audit/remaining-cleanup-roi.md) | Phase 4 DRY + Phase 6 abstractions (in progress) |
| [adr/](./adr/) | Cross-package ADRs |

---

## Priority stack (strict order)

Work top-to-bottom. Weights = share of total cleanup effort (phase **0** and **∞** are cross-cutting, not in the 100%).

| # | Phase | Weight | Status |
| --- | --- | --- | --- |
| **0** | Safety net (`check.ps1`, contract tests on touch) | — | ongoing |
| **1** | Boundaries & contracts | **25%** | done |
| **2** | Simplify / delete | **25%** | done |
| **3** | Centralize policy | **20%** | done ([#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203); [policy-roi](./audit/policy-roi.md)) |
| **4** | Targeted deduplication (3rd occurrence) | **10%** | in progress ([remaining-cleanup-roi](./audit/remaining-cleanup-roi.md)) |
| **5** | Structural patterns (god-files, handler/store splits) | **35%** | done ([#205](https://github.com/AlexsanderHamir/Hamix/pull/205)–[#215](https://github.com/AlexsanderHamir/Hamix/pull/215); [structural-patterns-roi](./audit/structural-patterns-roi.md)) |
| **6** | Abstractions (≥2 real impls) | **10%** | in progress ([remaining-cleanup-roi](./audit/remaining-cleanup-roi.md)) |
| **∞** | Docs (focused doc + ADR per PR) | — | ongoing |

Tier 5 + post–Tier 5 handoff ([#191](https://github.com/AlexsanderHamir/Hamix/pull/191)–[#192](https://github.com/AlexsanderHamir/Hamix/pull/192), handoff train merged): **done**.

---

## Next queue (remaining ~20%)

Phases 3 (**20%**) and 5 (**35%**) are complete. Phase 4 + 6 **in progress** — ranked handoff: [remaining-cleanup-roi.md](./audit/remaining-cleanup-roi.md).

| # | Slice | Weight | Source |
| --- | --- | --- | --- |
| **1** | Third-occurrence DRY (PR2–PR5) | **10%** | Phase 4 — [remaining-cleanup-roi](./audit/remaining-cleanup-roi.md) |
| **2** | Late abstractions ≥2 impls (PR6–PR10) | **10%** | Phase 6 — [remaining-cleanup-roi](./audit/remaining-cleanup-roi.md) |

---

## Phase rules (one line each)

| # | Do | Don't |
| --- | --- | --- |
| **0** | Extend existing tests on every move | Rename without test touch |
| **1** | Fix dependency direction (`.cursor/rules/backend/go/layout.mdc`, `web-layout.mdc`) | New repository/UoW layers |
| **2** | Delete dead code; split red-zone files | Wrap code you should delete |
| **3** | One choke point per invariant (`readpolicy/`, `writepolicy/`, `tasks/sync/`, `queryInvalidation/`) | Generic policy frameworks |
| **4** | Extract on 3rd same-shape occurrence | DRY two call sites |
| **5** | BC splits where pain is measurable | Wholesale DDD; product/harness features |
| **6** | Interfaces when ≥2 implementations exist | Hypothetical plugins |
| **∞** | Update `api.md` / `web.md` / ADR with behavior | Doc-only sprints |

---

## Agent checklist

- [ ] Read this doc + one backlog row
- [ ] Confirm slice is architectural / organizational / testing — not product
- [ ] One invariant or one deletion per PR
- [ ] `.\scripts\check.ps1` before done

---

## See also

[guide.md](./guide.md) · [agent-map.md](./agent-map.md) · [CONTRIBUTING.md](../CONTRIBUTING.md)
