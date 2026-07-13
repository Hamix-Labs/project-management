# Policy centralization audit — ROI-ranked findings

> Read-only audit (2026-07-12). **Phase 3 complete** (2026-07-13): implementation landed [#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203).

**Handoff goal:** One choke point per invariant so a new engineer or agent finds policy in catalog modules and CI gates — not scattered hooks and magic numbers.

## Summary

- Items found: **7** (High: 3, Medium: 3, Low: 1)
- Top 3 by ROI:
  1. Task invalidation catalog missing (no `decideTaskInvalidationKeys`; `tasks/` has no CI gate)
  2. Read limits scattered outside `readpolicy` bootstrap trio
  3. CI policy gates incomplete (tasks vertical + read-limit parity)

## ROI legend

| Score | Effort | Risk | Typical action |
| --- | --- | --- | --- |
| 8–10 | ≤2 days | Low–medium | Do next sprint |
| 5–7 | 2–4 days | Medium | One PR per choke point |
| 1–4 | Docs-only or defer | Low | Batch with related PR |

**Formula:** `(clarity_gain × blast_radius) / (effort × risk)` — scores rounded 1–10.

---

## Verified clean (baseline for handoff)

| Concern | Choke point | ADR |
| --- | --- | --- |
| Query staleTime tiers | `web/src/lib/queryPolicy.ts` | ADR-0025 |
| Bootstrap read limits | `pkgs/tasks/handler/readpolicy/readpolicy.go` + `web/src/lib/readLimits.ts` (`TASK_LIST_PAGE_SIZE === BootstrapListLimit`) | ADR-0026 |
| Project/git invalidation | `web/src/lib/queryInvalidation/` + `projects/mutations/`, `worktrees/mutations/` | ADR-0044 |
| SSE sync orchestration | `web/src/tasks/sync/` (`decideSyncFrame`, `taskSyncCoordinator`) | ADR-0022 |
| Mutation guard | `web/src/tasks/sync/mutationGuard.ts` + `web/src/tasks/mutations/guardedTaskWrite.ts` | ADR-0025 M1–M2 |
| writepolicy package purity | `scripts/check-code-standards.ps1` (readpolicy/writepolicy import gate) | ADR-0026 |
| SSE publish runtime choke | `pkgs/tasks/handler/sse_notify.go` (`publishPolicyEvent` + `writepolicy`) | [#202](https://github.com/AlexsanderHamir/Hamix/pull/202) |
| Task invalidation catalog | `web/src/lib/queryInvalidation/decideTaskInvalidationKeys.ts` + `invalidateTaskCache` | ADR-0080 |
| Sync flush invalidation | `decideFlushBatch.ts` / `applySyncEffects.ts` share catalog `listStats` scope | [#203](https://github.com/AlexsanderHamir/Hamix/pull/203) |
| invalidateQueries CI (partial) | same script — **projects/**, **worktrees/**, **tasks/** (`mutations/`, `sync/` allowed) | ADR-0044, ADR-0080 |

---

## Findings (ranked)

### 1. Task invalidation catalog missing — ROI 9/10 (High) — **Status: done (2026-07-12)**

- **PR:** ADR-0080, `decideTaskInvalidationKeys`, `invalidateTaskCache`, CI gate for `tasks/`.

### 2. Read limits scattered — ROI 8/10 (High) — **Status: done** (PR #201)

- **Location:** `readpolicy` has bootstrap limits only. Duplicates: `pkgs/taskcore/handler/handler_task_crud.go` (`limit=50`, max 200), `pkgs/taskcycles/handler/handler_cycles_query.go` (50/200, stream 100/500), `pkgs/taskevents/handler/handler_events.go` (50/200), `pkgs/taskcore/store/internal/stats/list_cycle_failures.go`. Web: `web/src/api/cycles.ts`, `web/src/constants/tasks.ts`, `useTaskCycleDetailPageState.ts`.
- **Issue:** Backend/frontend caps drift without a single module; only bootstrap list limit is documented in ADR-0026.
- **Proposed change:** Extend `readpolicy` with named constants; BC handlers import them; mirror in `web/src/lib/readLimits.ts` or `task-paging`; Go ↔ TS parity tests.
- **Effort / risk:** 2 days; low–medium; touches 4 handler packages + web api layer.
- **Evidence:** `rg 'limit = 50|defaultCycleListLimit|maxCycleListLimit' pkgs/`; `readpolicy.go` has 3 bootstrap constants only.
- **Success signal:** Migrated handlers have zero magic limit literals; parity test locks bootstrap + events + cycles.

### 3. SSE publish policy not enforced at runtime — ROI 7/10 (Medium) — **Status: done** ([#202](https://github.com/AlexsanderHamir/Hamix/pull/202))

- **Location:** `pkgs/tasks/handler/writepolicy/publish_policy.go` (classification); `pkgs/tasks/handler/sse_notify.go` (actual `hub.Publish`); BC handlers inject `Notify` callbacks (`projects/handler`, `settings/handler`, `taskcore/handler`).
- **Issue:** `IsHintOnly` / `EnrichedTaskChangeEvent` used in tests only (`handler_writepolicy_test.go`, `publish_policy_test.go`). Publish path does not consult writepolicy — handoff cannot trust one table.
- **Proposed change:** Route `sse_notify` helpers through writepolicy choke (assert hint-only vs enriched payload shape). No SSE wire JSON change.
- **Effort / risk:** 1–2 days; medium (SSE contract tests must stay green).
- **Evidence:** `rg writepolicy\. pkgs/` → handlers/tests only, not `sse_notify.go`.
- **Success signal:** Choke-path tests; `docs/domain/sse-hub.md` points at writepolicy table.

### 4. Mutation post-success invalidation inconsistent — ROI 7/10 (Medium) — **Status: done (2026-07-12)**

- **Resolution:** Hooks use `listStats` scope via `invalidateTaskCache` (ADR-0080).

### 5. Sync invalidation logic overlap — ROI 6/10 (Medium) — **Status: done** ([#203](https://github.com/AlexsanderHamir/Hamix/pull/203))

- **Location:** `web/src/tasks/sync/decideFlushBatch.ts`, `applySyncEffects.ts` (duplicate list/stats keys); `decideSyncFrame.ts` inline keys for `task_event`, `resync`, `settings`.
- **Issue:** Three modules encode overlapping invalidation key lists; flush vs mutation paths can diverge.
- **Proposed change:** After task catalog (#1), share flush key selection; document sync-owned vs catalog-owned frames (ADR-0022).
- **Effort / risk:** 1–2 days; low; depends on #1.
- **Evidence:** `decideFlushBatch` and `applySyncEffects` both push `taskQueryKeys.listRoot()` / `stats()`.
- **Success signal:** Single source for flush keys; `decideSyncFrame.test.ts` green.

### 6. CI policy gates incomplete — ROI 8/10 (High) — **Status: done (2026-07-12)**

- **Resolution:** Tasks vertical gate (ADR-0080) and Go ↔ TS read-limit parity test (`testdata/readlimits.json`) shipped in policy PR train #2 and #201.

### 7. Handler default vs SPA explicit limit drift — ROI 5/10 (Low) — **Status: done** (PR #201)

- **Location:** `parseListParams` default `limit=50` in `handler_task_crud.go`; SPA always sends `limit=20` (`TASK_LIST_PAGE_SIZE`); `readpolicy.BootstrapListLimit = 20`.
- **Issue:** Not a runtime bug (SPA passes explicit limit) but confusing for API consumers and handoff docs.
- **Proposed change:** Document in `docs/api.md`; optional comment in readpolicy linking default vs bootstrap.
- **Effort / risk:** ≤2 hours; none.
- **Evidence:** `handler_task_crud.go` L296 vs `paging.ts` L2.
- **Success signal:** `api.md` states default 50 vs SPA 20; audit row closed.

---

## Suggested implementation order

| PR | Slice | Findings | Status |
| --- | --- | --- | --- |
| 1 | This audit doc | — | merged [#198](https://github.com/AlexsanderHamir/Hamix/pull/198) |
| 2 | Task invalidation catalog + CI gate | #1, #4, #6 (tasks) | merged [#199](https://github.com/AlexsanderHamir/Hamix/pull/199) |
| 3 | Read limits catalog + parity | #2, #7, #6 (read) | merged [#201](https://github.com/AlexsanderHamir/Hamix/pull/201) |
| 4 | SSE publish choke | #3 | merged [#202](https://github.com/AlexsanderHamir/Hamix/pull/202) |
| 5 | Sync invalidation dedup | #5 | merged [#203](https://github.com/AlexsanderHamir/Hamix/pull/203) |

Phase 3 closed — see [cleanup-order.md](../cleanup-order.md).

---

## Not flagged (verified live)

| Item | Reason |
| --- | --- |
| `queryPolicy.ts` staleTime tiers | Centralized; ADR-0025 |
| Project/git `decide*InvalidationKeys` | ADR-0044 + CI |
| `guardedTaskWrite` + mutation guard | ADR-0025 M1–M2 on task mutations |
| `decideSyncFrame` architecture | ADR-0022; owns SSE frame orchestration |
| Harness / operator UX | Out of scope per cleanup-order |
