# Policy centralization audit — ROI-ranked findings

> Read-only audit (2026-07-12). No code changed in the audit PR. Implementation: [policy-roi PR train](policy-roi.md#suggested-implementation-order) via `docs/cleanup-order.md` Phase 3.

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
| Bootstrap read limits | `pkgs/tasks/handler/readpolicy/readpolicy.go` + `web/src/tasks/task-paging/paging.ts` (`TASK_LIST_PAGE_SIZE === 20`) | ADR-0026 |
| Project/git invalidation | `web/src/lib/queryInvalidation/` + `projects/mutations/`, `worktrees/mutations/` | ADR-0044 |
| SSE sync orchestration | `web/src/tasks/sync/` (`decideSyncFrame`, `taskSyncCoordinator`) | ADR-0022 |
| Mutation guard | `web/src/tasks/sync/mutationGuard.ts` + `web/src/tasks/mutations/guardedTaskWrite.ts` | ADR-0025 M1–M2 |
| writepolicy package purity | `scripts/check-code-standards.ps1` (readpolicy/writepolicy import gate) | ADR-0026 |
| invalidateQueries CI (partial) | same script — **projects/** and **worktrees/** only | ADR-0044 |

---

## Findings (ranked)

### 1. Task invalidation catalog missing — ROI 9/10 (High) — **Status: open**

- **Location:** Inline `invalidateQueries` in `web/src/tasks/hooks/useTaskPatchFlow.ts`, `useTaskDetailMutations.ts`, `useTaskDeleteFlow.ts`; `web/src/tasks/checklist/checklistOptimistic.ts`; `web/src/tasks/mutations/patchTaskEventUserResponseMutation.ts`. Partial helpers `invalidateTaskListCoherence.ts`, `invalidateTaskDetailCoherence.ts` — not a full scope catalog.
- **Issue:** Project/git invalidation has `decideProjectInvalidationKeys` / `decideGitInvalidationKeys` (ADR-0044) with CI enforcement; task vertical does not. Handoff engineer must grep hooks to learn cache effects.
- **Proposed change:** Add `decideTaskInvalidationKeys` in `web/src/lib/queryInvalidation/`; migrate callers; extend `check-code-standards.ps1` for `tasks/` (allow `tasks/mutations/`, `tasks/sync/`, `lib/queryInvalidation/`). ADR-0080.
- **Effort / risk:** 2–3 days; medium (mutation coherence); ~12 production call sites.
- **Evidence:** `rg invalidateQueries web/src/tasks` (production, non-test); CI gate in `scripts/check-code-standards.ps1` L165–186 lists only projects/worktrees.
- **Success signal:** CI fails on new inline `invalidateQueries` in `tasks/hooks/`; table-driven tests mirror project catalog pattern.

### 2. Read limits scattered — ROI 8/10 (High) — **Status: open**

- **Location:** `readpolicy` has bootstrap limits only. Duplicates: `pkgs/taskcore/handler/handler_task_crud.go` (`limit=50`, max 200), `pkgs/taskcycles/handler/handler_cycles_query.go` (50/200, stream 100/500), `pkgs/taskevents/handler/handler_events.go` (50/200), `pkgs/taskcore/store/internal/stats/list_cycle_failures.go`. Web: `web/src/api/cycles.ts`, `web/src/constants/tasks.ts`, `useTaskCycleDetailPageState.ts`.
- **Issue:** Backend/frontend caps drift without a single module; only bootstrap list limit is documented in ADR-0026.
- **Proposed change:** Extend `readpolicy` with named constants; BC handlers import them; mirror in `web/src/lib/readLimits.ts` or `task-paging`; Go ↔ TS parity tests.
- **Effort / risk:** 2 days; low–medium; touches 4 handler packages + web api layer.
- **Evidence:** `rg 'limit = 50|defaultCycleListLimit|maxCycleListLimit' pkgs/`; `readpolicy.go` has 3 bootstrap constants only.
- **Success signal:** Migrated handlers have zero magic limit literals; parity test locks bootstrap + events + cycles.

### 3. SSE publish policy not enforced at runtime — ROI 7/10 (Medium) — **Status: open**

- **Location:** `pkgs/tasks/handler/writepolicy/publish_policy.go` (classification); `pkgs/tasks/handler/sse_notify.go` (actual `hub.Publish`); BC handlers inject `Notify` callbacks (`projects/handler`, `settings/handler`, `taskcore/handler`).
- **Issue:** `IsHintOnly` / `EnrichedTaskChangeEvent` used in tests only (`handler_writepolicy_test.go`, `publish_policy_test.go`). Publish path does not consult writepolicy — handoff cannot trust one table.
- **Proposed change:** Route `sse_notify` helpers through writepolicy choke (assert hint-only vs enriched payload shape). No SSE wire JSON change.
- **Effort / risk:** 1–2 days; medium (SSE contract tests must stay green).
- **Evidence:** `rg writepolicy\. pkgs/` → handlers/tests only, not `sse_notify.go`.
- **Success signal:** Choke-path tests; `docs/domain/sse-hub.md` points at writepolicy table.

### 4. Mutation post-success invalidation inconsistent — ROI 7/10 (Medium) — **Status: open**

- **Location:** `useTaskPatchFlow.ts` / `useTaskDetailMutations.ts` `onSuccess` → `taskQueryKeys.all` + `stats()`; `invalidateTaskListCoherence.ts` → `listRoot()` + `stats()` only.
- **Issue:** Over-invalidation on patch hides intended narrow contract; inconsistent with ADR-0025 list coherence goals.
- **Proposed change:** Resolve in PR #2 via `decideTaskInvalidationKeys` scopes (`listStats`, `detail`, etc.).
- **Effort / risk:** Bundled with #1; low incremental.
- **Evidence:** Diff hook `onSuccess` blocks vs `invalidateTaskListAndStats`.
- **Success signal:** All task mutations use catalog scopes; tests lock per mutation kind.

### 5. Sync invalidation logic overlap — ROI 6/10 (Medium) — **Status: open**

- **Location:** `web/src/tasks/sync/decideFlushBatch.ts`, `applySyncEffects.ts` (duplicate list/stats keys); `decideSyncFrame.ts` inline keys for `task_event`, `resync`, `settings`.
- **Issue:** Three modules encode overlapping invalidation key lists; flush vs mutation paths can diverge.
- **Proposed change:** After task catalog (#1), share flush key selection; document sync-owned vs catalog-owned frames (ADR-0022).
- **Effort / risk:** 1–2 days; low; depends on #1.
- **Evidence:** `decideFlushBatch` and `applySyncEffects` both push `taskQueryKeys.listRoot()` / `stats()`.
- **Success signal:** Single source for flush keys; `decideSyncFrame.test.ts` green.

### 6. CI policy gates incomplete — ROI 8/10 (High) — **Status: open**

- **Location:** `scripts/check-code-standards.ps1` — tasks vertical not gated; no Go ↔ TS read-limit parity test.
- **Issue:** Regressions reintroduce inline invalidation or limit drift without CI failure.
- **Proposed change:** Ship gates in same PRs as migrations (#1 tasks gate, #2 parity test) — not a standalone “enable gate later” PR.
- **Effort / risk:** ≤1 day each; low.
- **Evidence:** CI script verticals array = projects, worktrees only.
- **Success signal:** `check-code-standards.ps1` fails on violation; parity test in readpolicy PR.

### 7. Handler default vs SPA explicit limit drift — ROI 5/10 (Low) — **Status: open**

- **Location:** `parseListParams` default `limit=50` in `handler_task_crud.go`; SPA always sends `limit=20` (`TASK_LIST_PAGE_SIZE`); `readpolicy.BootstrapListLimit = 20`.
- **Issue:** Not a runtime bug (SPA passes explicit limit) but confusing for API consumers and handoff docs.
- **Proposed change:** Document in `docs/api.md`; optional comment in readpolicy linking default vs bootstrap.
- **Effort / risk:** ≤2 hours; none.
- **Evidence:** `handler_task_crud.go` L296 vs `paging.ts` L2.
- **Success signal:** `api.md` states default 50 vs SPA 20; audit row closed.

---

## Suggested implementation order

| PR | Slice | Findings |
| --- | --- | --- |
| 1 | This audit doc | — |
| 2 | Task invalidation catalog + CI gate | #1, #4, #6 (tasks) |
| 3 | Read limits catalog + parity | #2, #7, #6 (read) |
| 4 | SSE publish choke | #3 |
| 5 | Sync invalidation dedup | #5 (after #2 merged) |

See [cleanup-order.md](../cleanup-order.md) Phase 3.

---

## Not flagged (verified live)

| Item | Reason |
| --- | --- |
| `queryPolicy.ts` staleTime tiers | Centralized; ADR-0025 |
| Project/git `decide*InvalidationKeys` | ADR-0044 + CI |
| `guardedTaskWrite` + mutation guard | ADR-0025 M1–M2 on task mutations |
| `decideSyncFrame` architecture | ADR-0022; owns SSE frame orchestration |
| Harness / operator UX | Out of scope per cleanup-order |
