# Structural patterns audit — ROI-ranked findings

> Read-only audit (2026-07-12). **Phase 5 complete** (PR1–PR10 merged). God-file splits across web production, Go handlers/store, and test suites landed in the structural PR train.

**Handoff goal:** A new engineer or agent finds **where files are too large**, **how to split them**, and **which PR to land first** — without re-scanning the repo.

## Summary

- Items found: **14** (High: 7, Medium: 6, Low: 1)
- Top 3 by ROI:
  1. `taskevents/handler_events.go` — largest production handler; blocks handler test readability
  2. `TaskListDataTableRow.tsx` — five exports in one red-zone file; list table changes are high-traffic
  3. `AttemptActivitySection.tsx` — cycle detail activity tab mixes three panel types in one container

## ROI legend

| Score | Effort | Risk | Typical action |
| --- | --- | --- | --- |
| 8–10 | ≤2 days | Low–medium | Do next sprint |
| 5–7 | 2–4 days | Medium | One PR per choke point |
| 1–4 | Docs-only or defer | Low | Batch with related PR |

**Formula:** `(clarity_gain × blast_radius) / (effort × risk)` — scores rounded 1–10.

---

## Verified clean (baseline for handoff)

| Concern | Choke point | Evidence |
| --- | --- | --- |
| Cycle detail page shell | `web/src/tasks/pages/TaskCycleDetailPage.tsx` | 36 lines — thin page; delegates to attempt components |
| Cycles handler DTOs | `pkgs/taskcycles/handler/handler_cycles_json.go` | 223 lines — green zone; JSON split already done |
| BC route registration | `pkgs/tasks/handler/handler_routes.go` + `*.Register(m, Deps)` | Each BC owns routes; composition shell stays thin |
| Task detail page shell | `web/src/tasks/pages/TaskDetailPage.tsx` | 101 lines — yellow but acceptable; panels already modular under `task-detail/` |
| Create modal shell | `web/src/tasks/components/task-create-modal/TaskCreateModal.tsx` | 145 lines — yellow; field sections already extracted under `fields/` |
| Handler split guide | `pkgs/tasks/handler/README.md` §When a file feels too large | BC `Register` map + split table; warn-only size gate in `check-code-standards.ps1` |

---

## Findings (ranked)

### 1. `taskevents` handler god-file — ROI 9/10 (High) — **Status: done** ([#207](https://github.com/AlexsanderHamir/Hamix/pull/207))

- **Location:** `pkgs/taskevents/handler/handler_events.go` (489 lines).
- **Issue:** Red zone (>500) production handler; mixes route wiring, query parsing, JSON DTOs, and local HTTP helpers. Highest line count among BC handlers.
- **Proposed split:** `handler_events.go` (routes) + `handler_events_json.go`; migrate duplicate HTTP helpers to `pkgs/tasks/handlerhttp`.
- **Effort / risk:** 1–2 days; medium — contract tests on `/tasks/{id}/events*` must stay green.
- **Evidence:** `wc -l pkgs/taskevents/handler/handler_events.go` → 489; CODE_STANDARDS handler red >500.
- **Success signal:** No file in `taskevents/handler` >300 lines; existing HTTP contract tests green.
- **PR:** #6 `refactor/go-handler-splits`

### 2. Task list table monolith — ROI 9/10 (High) — **Status: done** ([#209](https://github.com/AlexsanderHamir/Hamix/pull/209))

- **Location:** `web/src/tasks/components/task-list/table/TaskListDataTableRow.tsx` (401 lines).
- **Issue:** Five exports (`TaskListTableSortHeader`, `TaskListTableHeader`, `TaskListTableBody`, `TaskListDataTableRow`, `syncHeaderCheckboxIndeterminate`) in one red-zone file; list UI is high-traffic.
- **Proposed split:** One component per file under `task-list/table/`; selection helper to `taskListTableSelection.ts`.
- **Effort / risk:** 1 day; low — move-only; update barrel imports.
- **Evidence:** `wc -l` → 401; CODE_STANDARDS presentational red >250.
- **Success signal:** Each table component ≤150 lines; list section tests green.
- **PR:** #4 `refactor/web-task-detail-splits`

### 3. Cycle attempt activity section — ROI 9/10 (High) — **Status: done** ([#206](https://github.com/AlexsanderHamir/Hamix/pull/206))

- **Location:** `web/src/tasks/components/task-detail/attempt/AttemptActivitySection.tsx` (373 lines).
- **Issue:** Red container; mixes Cursor activity, audit timeline, and stream event row rendering in one file.
- **Proposed split:** Container ≤120 lines + `CursorActivityPanel.tsx`, `AuditActivityPanel.tsx`, `StreamEventRow.tsx` co-located under `attempt/`.
- **Effort / risk:** 1 day; low — `TaskCycleDetailPage.test.tsx` covers regressions.
- **Evidence:** `wc -l` → 373; container limit red >200.
- **Success signal:** `AttemptActivitySection.tsx` ≤150 lines; children ≤250 each.
- **PR:** #2 `refactor/web-cycle-attempt-splits`

### 4. Web test god-files (top 4) — ROI 8/10 (High) — **Status: done** ([#211](https://github.com/AlexsanderHamir/Hamix/pull/211))

- **Location:** `TaskListSection.test.tsx` (935), `TaskCyclesPanel.test.tsx` (705), `useTasksApp.test.tsx` (569), `useTaskEventStream.test.tsx` (513).
- **Issue:** All exceed web-testing-bar red threshold (>500 lines); slow to navigate and review.
- **Proposed split:** One file per concern (filters/sort, bulk selection, SSE frames, modal state); share MSW setup via `web/src/test/`.
- **Effort / risk:** 2–3 days; low — test-only moves after production splits stable.
- **Evidence:** `wc -l` on each file; CODE_STANDARDS `*.test.tsx` red >500.
- **Success signal:** All four files ≤500 lines; full web test suite green.
- **PR:** #5 `refactor/web-test-splits`

### 5. Agents test god-files — ROI 8/10 (High) — **Status: done** ([#214](https://github.com/AlexsanderHamir/Hamix/pull/214))

- **Location:** `pkgs/agents/runner/cursor/cursor_test.go` (1058), `pkgs/agents/worker/worker_test.go` (944), `pkgs/agents/runner/runner_test.go` (526).
- **Issue:** Cursor and worker tests far past Go test red zone (>600); runner test in yellow/red border.
- **Proposed split:** `cursor_run_*_test.go` by scenario; `worker_*_test.go` by happy-path/failure/shutdown/queue; `runner_*_test.go` by concern.
- **Effort / risk:** 2 days; low–medium — agents CI group is coverage-gated.
- **Evidence:** `wc -l` → 1058 / 944 / 526.
- **Success signal:** Each file ≤600 lines; `go-tests (agents)` green.
- **PR:** #9 `refactor/agents-test-splits`

### 6. Cycle attempt phases section — ROI 7/10 (Medium) — **Status: done** ([#206](https://github.com/AlexsanderHamir/Hamix/pull/206))

- **Location:** `web/src/tasks/components/task-detail/attempt/AttemptPhasesSection.tsx` (283 lines).
- **Issue:** Red container; phase list presentation mixed with container state.
- **Proposed split:** `AttemptPhasesSection.tsx` (container ≤120) + `AttemptPhaseList.tsx` (presentational).
- **Effort / risk:** ≤1 day; low.
- **Evidence:** `wc -l` → 283.
- **Success signal:** Container ≤150 lines; cycle detail tests green.
- **PR:** #2 `refactor/web-cycle-attempt-splits`

### 7. `storefake` unimplemented stubs — ROI 7/10 (High) — **Status: done** ([#212](https://github.com/AlexsanderHamir/Hamix/pull/212))

- **Location:** `pkgs/tasks/handler/storefake/handler_store.go` (371 lines).
- **Issue:** Red zone; all BC unimplemented store stubs in one file — hard to see which BC a test fake covers.
- **Proposed split:** `handler_store.go` + `storefake_unimplemented_<bc>.go` per bounded context.
- **Effort / risk:** 1 day; low — move-only; handler tests import paths update only.
- **Evidence:** `wc -l` → 371; CODE_STANDARDS store red >500 approaching.
- **Success signal:** `handler_store.go` ≤200 lines; handler contract tests green.
- **PR:** #7 `refactor/go-store-splits`

### 8. Cycles store `phases` internal — ROI 7/10 (High) — **Status: done** ([#212](https://github.com/AlexsanderHamir/Hamix/pull/212))

- **Location:** `pkgs/taskcycles/store/internal/cycles/phases.go` (448 lines).
- **Issue:** Red zone store internal; read and write paths colocated.
- **Proposed split:** `phases_read.go` + `phases_write.go` (or concern-based names per CODE_STANDARDS).
- **Effort / risk:** 1–2 days; medium — extend store tests on touch (`tasktestdb`).
- **Evidence:** `wc -l` → 448.
- **Success signal:** Each file ≤300 lines; `go-tests (tasks)` green.
- **PR:** #7 `refactor/go-store-splits`

### 9. Checklist store internal — ROI 7/10 (Medium) — **Status: done** ([#212](https://github.com/AlexsanderHamir/Hamix/pull/212))

- **Location:** `pkgs/taskchecklist/store/internal/checklist/checklist.go` (415 lines).
- **Issue:** Red zone; same read/write colocation pattern as phases.
- **Proposed split:** `checklist_read.go` + `checklist_write.go`.
- **Effort / risk:** 1 day; medium.
- **Evidence:** `wc -l` → 415.
- **Success signal:** Each file ≤300 lines.
- **PR:** #7 `refactor/go-store-splits`

### 10. Create modal cluster — ROI 7/10 (Medium) — **Status: done** ([#210](https://github.com/AlexsanderHamir/Hamix/pull/210))

- **Location:** `TaskCreateModalFormBody.tsx` (220), `fields/TaskCreateModalAgentSection.tsx` (236), `create/hooks/useTaskCreateMutations.ts` (226), `useTaskCreateEntryActions.ts` (207).
- **Issue:** Form body and agent section in red/yellow border; hooks mix mutation vs entry-routing concerns.
- **Proposed split:** Section components per field group; agent section → runner/model fields; narrow hooks to single responsibility.
- **Effort / risk:** 1–2 days; low — `TaskCreateModal.test.tsx` guards behavior.
- **Evidence:** `wc -l` on cluster files; shell `TaskCreateModal.tsx` already 145 (yellow).
- **Success signal:** No file in create-modal cluster in red zone.
- **PR:** #3 `refactor/web-create-modal-splits`

### 11. Task detail loaded view layout — ROI 7/10 (Medium) — **Status: done** ([#209](https://github.com/AlexsanderHamir/Hamix/pull/209))

- **Location:** `web/src/tasks/pages/TaskDetailLoadedView.tsx` (224 lines).
- **Issue:** Red container; composes existing panels but layout wiring dominates the file.
- **Proposed split:** Thin `TaskDetailLoadedView.tsx` + layout children (`TaskDetailLayoutColumns.tsx` or per-region wrappers) that only compose existing `task-detail/` panels.
- **Effort / risk:** ≤1 day; low — compose, don't duplicate panel logic.
- **Evidence:** `wc -l` → 224; `TaskDetailPage.tsx` already 101 (yellow).
- **Success signal:** `TaskDetailLoadedView.tsx` ≤120 lines.
- **PR:** #4 `refactor/web-task-detail-splits`

### 12. Handler / SSE contract test god-files — ROI 6/10 (Medium) — **Status: done** ([#213](https://github.com/AlexsanderHamir/Hamix/pull/213))

- **Location:** `handler_http_checklist_contract_test.go` (642), `handler_http_cycles_contract_test.go` (579), `sse_trigger_surface_test.go` (521), `sse_lossless_test.go` (447).
- **Issue:** Contract harness files past Go test red (>600) or yellow; shared server setup duplicated.
- **Proposed split:** By HTTP verb / route group / SSE write surface; extract `handler_http_test_helpers_test.go`.
- **Effort / risk:** 1–2 days; low — test-only after PR6 handler splits.
- **Evidence:** `wc -l` on each file.
- **Success signal:** Priority four files ≤600 lines; `go-tests (tasks)` green.
- **PR:** #8 `refactor/go-handler-test-splits`

### 13. `taskcore` / `taskcycles` handler yellow zone — ROI 6/10 (Medium) — **Status: done** ([#207](https://github.com/AlexsanderHamir/Hamix/pull/207))

- **Location:** `pkgs/taskcore/handler/handler_task_crud.go` (320), `pkgs/taskcycles/handler/handler_cycles.go` (318).
- **Issue:** Yellow zone (301–500); patch logic and phases/stream routes add review burden.
- **Proposed split:** `handler_task_patch.go` from crud; `handler_cycles_phases.go` + `handler_cycles_stream.go` from cycles (mirror existing `handler_cycles_json.go` pattern).
- **Effort / risk:** 1–2 days; medium — HTTP contract tests on touched routes.
- **Evidence:** `wc -l` → 320 / 318; `handler_cycles_json.go` already 223 (green).
- **Success signal:** No scoped handler file >300 lines.
- **PR:** #6 `refactor/go-handler-splits`

### 14. Handler README split guide + CI warn gate — ROI 5/10 (Low) — **Status: done** ([#215](https://github.com/AlexsanderHamir/Hamix/pull/215))

- **Location:** `pkgs/tasks/handler/README.md` §When a file feels too large; `scripts/check-code-standards.ps1` warn-only size scan.
- **Issue:** Handoff doc listed BC routes but not split patterns; no automated regression signal when files re-enter red zone.
- **Resolution:** README split table + test placement; warn-only line-count check in `check-code-standards.ps1` (exit 0); Phase 5 marked done in [cleanup-order.md](../cleanup-order.md).

---

## Suggested implementation order

| PR | Branch | Slice | Findings | CI gate |
| --- | --- | --- | --- | --- |
| 1 | `audit/structural-patterns-roi` | This audit doc | — | Full `.\scripts\check.ps1` + all Actions jobs green |
| 2 | `refactor/web-cycle-attempt-splits` | Cycle attempt sections | #3, #6 | Full check + all jobs green before PR3 |
| 3 | `refactor/web-create-modal-splits` | Create modal cluster | #10 | Full check + all jobs green (parallel with PR2/4 after PR1) |
| 4 | `refactor/web-task-detail-splits` | Task detail + list table | #2, #11 | Full check + all jobs green |
| 5 | `refactor/web-test-splits` | Web test god-files | #4 | Full check + all jobs green after PR2–4 |
| 6 | `refactor/go-handler-splits` | BC handler splits | #1, #13 | Full check + `go-tests (tasks)` green |
| 7 | `refactor/go-store-splits` | storefake + store internal | #7, #8, #9 | Full check + `go-tests (tasks)` green |
| 8 | `refactor/go-handler-test-splits` | Handler/SSE contract tests | #12 | Full check + `go-tests (tasks)` green |
| 9 | `refactor/agents-test-splits` | Agents test god-files | #5 | Full check + `go-tests (agents)` green |
| 10 | `chore/structural-ci-readme` | README + CI warn gate | #14 | Full check + all jobs green — **Phase 5 closed** |

**Train rule:** Do not branch the next PR until the current PR's CI is **fully green** (`go-lint`, four `go-tests` groups, `web`). Triage: `gh run view <RUN_ID> --repo AlexsanderHamir/Hamix --log-failed`.

Phase 5 is **complete** — see [cleanup-order.md](../cleanup-order.md).

---

## Not flagged (verified live)

| Item | Reason |
| --- | --- |
| `pkgs/tasks/handler` composition package | Intentional single-package route shell; BC logic lives in sibling `pkgs/*/handler` |
| `handler_cycles_json.go` | Already split; green zone |
| `TaskCycleDetailPage.tsx` | Thin page shell (36 lines) |
| `task-detail/` panel components | Already modular; PR4 only touches layout composer |
| Harness / operator UX | Out of scope per [cleanup-order.md](../cleanup-order.md) |
| `taskcycles/.../cycles.go` (331) | Yellow only; split in PR7 if still >300 after phases move |
