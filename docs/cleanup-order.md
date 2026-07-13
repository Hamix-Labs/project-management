# Cleanup order (agent guide)

**When:** Refactor / deletion / hardening — not new product surface.  
**Scope:** Architectural, organizational, and testing improvements only (boundaries, file layout, policy choke points, contract tests, god-file splits). Operator UX, harness reliability features, and new product APIs are **out of scope** until this pass ends.  
**Bar:** One slice per PR; `.\scripts\check.ps1` green.

**Backlog (pick rows; this doc is order only):**

| Index | Use for |
| --- | --- |
| [audit/README.md](./audit/README.md) | Active structural themes |
| [audit/policy-roi.md](./audit/policy-roi.md) | Phase 3 policy centralization (complete) |
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
| **4** | Targeted deduplication (3rd occurrence) | **10%** | not started |
| **5** | Structural patterns (god-files, handler/store splits) | **35%** | in progress |
| **6** | Abstractions (≥2 real impls) | **10%** | not started |
| **∞** | Docs (focused doc + ADR per PR) | — | ongoing |

Tier 5 + post–Tier 5 handoff ([#191](https://github.com/AlexsanderHamir/Hamix/pull/191)–[#192](https://github.com/AlexsanderHamir/Hamix/pull/192), handoff train merged): **done**.

---

## Next queue (remaining ~55%)

Phase 3 policy centralization (**20%**) is complete. Remaining slices:

| # | Slice | Weight | Source |
| --- | --- | --- | --- |
| **1** | Web god-file splits | **40%** | [audit/README.md](./audit/README.md) |
| **2** | Go handler / store file splits (CODE_STANDARDS red zone) | **25%** | [CODE_STANDARDS](../.cursor/rules/CODE_STANDARDS.mdc), [handler README](../pkgs/tasks/handler/README.md) |
| **3** | Handler test consolidation + fakes | **10%** | [audit/README.md](./audit/README.md) |
| **4** | Third-occurrence DRY + late abstractions | **5%** | Phases 4 + 6 |

---

## Phase rules (one line each)

| # | Do | Don't |
| --- | --- | --- |
| **0** | Extend existing tests on every move | Rename without test touch |
| **1** | Fix dependency direction (`.cursor/rules/go-layout.mdc`, `web-layout.mdc`) | New repository/UoW layers |
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

[guide.md](./guide.md) · [AGENTS.md](../AGENTS.md) · [CONTRIBUTING.md](../CONTRIBUTING.md)
