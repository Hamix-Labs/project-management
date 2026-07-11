# Cleanup order (agent guide)

**When to read:** Product feature freeze; work is refactor, deletion, or structural hardening — not new product surface.

**Goal:** Pay down fast-development debt without big-bang rewrites. One vertical slice per PR; `.\scripts\check.ps1` green before done.

**Backlog indexes (pick items from these; this doc defines *order*, not the task list):**

| Index | Use for |
| --- | --- |
| [audit/README.md](./audit/README.md) | Cross-report top-10, quick wins, structural themes |
| [audit/boundaries-roi.md](./audit/boundaries-roi.md) | Phase 1 — layer boundaries & contract seams |
| [audit/dead-code-roi.md](./audit/dead-code-roi.md) | Phase 2 — delete inventory (stubs, legacy stacks, barrels) |
| [../HARNESS_IMPROVEMENTS.md](../HARNESS_IMPROVEMENTS.md) | Harness reliability and operator leverage |
| [adr/](./adr/) | Cross-package structural decisions (write ADR before large splits) |

> **Note:** A Phase 2 umbrella [simplify-delete-roi.md](./audit/simplify-delete-roi.md) (delete + god-file splits + ADR finish) was drafted in-session but may need recreation. Until then, use [dead-code-roi.md](./audit/dead-code-roi.md) for deletions and [audit/README.md](./audit/README.md) §Structural themes for splits.

---

## Priority stack (strict order)

Work top-to-bottom. Do **not** skip ahead because a lower step looks easier.

```text
0. Safety net
1. Boundaries & contracts
2. Simplify / delete
3. Centralize policy
4. Targeted deduplication
5. Structural patterns (surgical DDD)
6. Abstractions (last)
∞. Docs (every PR)
```

Each phase is a **bucket** — many plans/PRs per number. This doc is the order; audit reports are the task list.

---

## Current progress (2026-07-08)

| Phase | Status | Notes |
| --- | --- | --- |
| **1. Boundaries** | **Largely done** | 12 findings in [boundaries-roi.md](./audit/boundaries-roi.md); shipped via [#147](https://github.com/AlexsanderHamir/Hamix/pull/147)–[#153](https://github.com/AlexsanderHamir/Hamix/pull/153) |
| **2. Simplify / delete** | **Next** | Start [dead-code-roi.md](./audit/dead-code-roi.md) **#1** (legacy project-scoped git stack, ROI 10) |
| **3–6** | Not started | Defer until Phase 2 high-ROI deletes land |
| **3 structural (early)** | **In progress** | `pkgs/projects/` extracted — ADR-0045; `pkgs/gitinventory/` extracted — ADR-0046; template for settings split |
| **∞. Docs** | Continuous | ADR-0044, ADR-0039 acceptance, audit status rows — same PR as behavior |

### Phase 1 PR map (dependency order)

Execute boundaries in this sequence (one concern per PR; do not merge unrelated findings):

| PR | Finding | Slice |
| --- | --- | --- |
| [#147](https://github.com/AlexsanderHamir/Hamix/pull/147) | boundaries #1 | Web vertical decouple (`tasks/` → inner ring) |
| [#148](https://github.com/AlexsanderHamir/Hamix/pull/148) | boundaries #2 | Guarded create/bulk mutations |
| [#149](https://github.com/AlexsanderHamir/Hamix/pull/149) | boundaries #3 | Query invalidation catalog ([ADR-0044](./adr/ADR-0044-query-invalidation-catalog.md)) |
| [#150](https://github.com/AlexsanderHamir/Hamix/pull/150) | boundaries #4 | `task_event_changed` SSE (stop hint-only `task_updated`) |
| [#151](https://github.com/AlexsanderHamir/Hamix/pull/151) | boundaries #5 | Settings `notifyScopelessChange` helper |
| [#152](https://github.com/AlexsanderHamir/Hamix/pull/152) | boundaries #6 | Checklist verify-commands guarded mutation |
| [#153](https://github.com/AlexsanderHamir/Hamix/pull/153) | boundaries #7–#12 | Event response mutation, scheduling guard, postgres migrate models, ADR-0039, query policy lib, bootstrap seed |

**Pivot rule:** Do not start Phase 2 deletions that touch the legacy git HTTP surface until boundaries Phase 1 is merged — invalidation catalog and vertical boundaries assume global git is the live path.

---

## Suggested execution order (cross-track)

After reading this doc, pick **one row** per plan/PR from the linked audit file.

1. **Phase 1 tail** — Merge [#153](https://github.com/AlexsanderHamir/Hamix/pull/153) if still open; confirm [boundaries-roi.md](./audit/boundaries-roi.md) §Verified clean greps pass.
2. **Phase 2 — dead-code #1** — Legacy project-scoped git stack (~900–1,100 lines). Extract shared JSON helpers first; then delete web hooks, Go handlers, MSW handlers, and `docs/api.md` legacy routes. See [dead-code-roi.md §1](./audit/dead-code-roi.md#1-legacy-project-scoped-git-stack--roi-1010-high).
3. **Phase 2 — dead-code #2** — Subtask-era CSS (~120–150 lines, visual-only).
4. **Phase 2 — dead-code #3–#9** — Quick-delete bundle (stubs, deprecated aliases, orphan barrels).
5. **Phase 2 — splits** — God files past CODE_STANDARDS red limits (see [audit/README.md](./audit/README.md) §Structural themes → Web god-file splits).
6. **Phase 3+** — Policy gaps, third-occurrence DRY, harness P0 from [HARNESS_IMPROVEMENTS.md](../HARNESS_IMPROVEMENTS.md).

---

### 0. Safety net

**Do:** Keep `.\scripts\check.ps1` green. Prefer extending existing contract tests (`handler_http_*_contract_test.go`, `parseTaskApi*`, sync `decideSyncFrame` tests) over refactoring without coverage.

**Don't:** Rename/move code with no test touch in the same PR.

Phase 0 is **not** a standalone plan — bake it into every child plan's exit criteria.

---

### 1. Boundaries & contracts

**Do:** Fix dependency-direction violations before any DRY or new layers.

| Layer | Rule | Violation signal |
| --- | --- | --- |
| Go `domain/` | stdlib only; no GORM/HTTP | `gorm:` tags, `datatypes.JSON` in domain |
| Go `store/` | SQL/GORM + map to domain | SQL or GORM in handlers |
| Go `handler/` | HTTP + domain types | DB drivers in handlers |
| `web/src/api/` | sole `fetch` owner | `fetch` in components |
| `web/src/features/` | no cross-feature imports | feature A imports feature B |
| Sync | policy in `tasks/sync/` | ad-hoc cache invalidation in components |

**ROI-ranked backlog:** [audit/boundaries-roi.md](./audit/boundaries-roi.md) — ranked violations and verified-clean boundaries. Pick one finding per plan/PR. Violation signals above are **audit checks**; several are already clean (see audit §Verified clean).

**Hamix anchors:** [ADR-0039](./adr/ADR-0039-domain-persistence-separation.md) (domain vs persistence), [ADR-0026](./adr/ADR-0026-backend-data-coherence.md) (write publish), [ADR-0025](./adr/ADR-0025-frontend-data-coherence.md) (sync/mutations), [ADR-0044](./adr/ADR-0044-query-invalidation-catalog.md) (project/git invalidation). Standards: `.cursor/rules/CODE_STANDARDS.mdc`.

**Don't:** Add repository/UoW/service layers to "fix" boundaries — split and map at the existing seam.

---

### 2. Simplify / delete

**Do:** Remove dead code, flatten control flow, finish accepted ADRs, split files past CODE_STANDARDS red limits.

**Don't:** Wrap code you should delete in new helpers. Don't abstract before the behavior is obviously right.

**ROI-ranked backlog:** [audit/dead-code-roi.md](./audit/dead-code-roi.md) — ranked deletions and suggested deletion order. Pick one finding per plan/PR.

**Detailed evidence:** Same file — full grep paths, effort, and blast radius per item.

**Start here:** [dead-code-roi.md §Suggested deletion order](./audit/dead-code-roi.md#suggested-deletion-order) — **#1 legacy git** first.

---

### 3. Centralize policy

**Do:** One choke point per invariant; migrate callers to it; then delete scattered copies.

**Existing patterns (extend, don't reinvent):**

| Concern | Choke point |
| --- | --- |
| Backend read limits | `pkgs/tasks/handler/readpolicy/` |
| Backend SSE after write | `handler_writepolicy.go`, `writepolicy/` |
| Frontend sync | `web/src/tasks/sync/` (`decideSyncFrame`, `applySyncEffects`) |
| Frontend mutations | `web/src/tasks/mutations/`, `mutationGuard.ts` |
| Project/git invalidation | `web/src/lib/queryInvalidation/` ([ADR-0044](./adr/ADR-0044-query-invalidation-catalog.md)) |

**Don't:** Build generic policy frameworks. One module, one concern.

---

### 4. Targeted deduplication

**Do:** Extract only on the **third** real occurrence with the same shape and same owner.

**Don't:** DRY two call sites "because it looks similar." Premature extraction creates wrong abstractions.

---

### 5. Structural patterns (surgical)

**Do:** Apply bounded-context splits only where measurable pain exists — god files, mixed resources, cross-stack contract drift.

| Area | Typical action |
| --- | --- |
| `pkgs/tasks` | Handler/store file splits; ADR-0039 phasing (leaf tables → `Task`/`Project`) |
| `pkgs/agents/harness` | `HARNESS_IMPROVEMENTS.md` P0/P1 before new orchestration |
| `web/` | Sync/API coherence; token-based CSS; feature boundary fixes |

**Don't:** "Implement DDD" wholesale. Align code to [architecture.md](./architecture.md), not a textbook.

---

### 6. Abstractions (last)

**Do:** Introduce interfaces/registries only when **≥2 real implementations** exist or a test seam is required on a hot path.

**Don't:** Plugin systems, base classes, or indirection for hypothetical future runners/features.

---

### ∞. Docs (continuous)

**Do (same PR as behavior change):** Update the focused doc (`api.md`, `web.md`, `domain/*`). Add ADR when a decision crosses packages. Keep [AGENTS.md](../AGENTS.md) routing accurate.

**Don't:** Documentation-only sprints that duplicate existing docs. Don't add a second source of truth.

---

## PR discipline

1. **One vertical slice** — one behavior or one boundary, not "refactor week."
2. **No mixed intent** — don't combine deletion + new feature + abstraction in one PR.
3. **Verify:** `.\scripts\check.ps1` (or `-GoOnly` / `-WebOnly` when scoped).
4. **Structural cross-package work:** ADR first if not already decided; link ADR in PR.

---

## Anti-patterns (stop and reconsider)

| Anti-pattern | Why |
| --- | --- |
| DRY before boundaries | Locks in wrong module ownership |
| Abstractions before deletion | Wraps dead code |
| New docs instead of updating existing | Drift |
| Big-bang domain rewrite | Use ADR-0039 phasing (leaf → core entities) |
| Chasing 100% DRY | Small duplication is often clearer |
| Deleting legacy git before extracting shared JSON helpers | Breaks global handlers that reuse projection code |

---

## Agent checklist (start of session)

- [ ] Read this doc + the relevant backlog row ([audit](./audit/README.md) or [harness](../HARNESS_IMPROVEMENTS.md))
- [ ] Confirm phase: am I doing 1–6, not shipping product?
- [ ] Identify the **single invariant** or **single deletion** for this PR
- [ ] Run `.\scripts\check.ps1` before finishing

---

## See also

- [audit/README.md](./audit/README.md) — ranked cleanup items and cross-report top-10
- [audit/boundaries-roi.md](./audit/boundaries-roi.md) — Phase 1 findings
- [audit/dead-code-roi.md](./audit/dead-code-roi.md) — Phase 2 delete inventory
- [guide.md](./guide.md) — doc routing for humans
- [AGENTS.md](../AGENTS.md) — code paths and scoped reads
- [CONTRIBUTING.md](../CONTRIBUTING.md) — PR verification bar
