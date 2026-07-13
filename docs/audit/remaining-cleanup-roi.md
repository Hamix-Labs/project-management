# Remaining cleanup ROI — Phase 4 + Phase 6

> Read-only audit (2026-07-13). Covers the last ~20% of [cleanup-order.md](../cleanup-order.md): **Phase 4** (third-occurrence DRY, 10%) and **Phase 6** (abstractions with ≥2 real impls, 10%).

**Handoff goal:** A new engineer or agent finds **what to DRY**, **what to abstract**, **what not to touch**, and **which PR comes next** — without re-scanning the repo.

**Lifecycle:** This audit (and related ROI / cleanup-order docs) is **deleted in PR10** when the train completes — not permanent archival.

## Summary

- Items found: **11** actionable (Phase 4: 5, Phase 6: 6) + anti-findings in **Not flagged**
- High: 5 · Medium: 5 · Low: 1
- Top 3 by ROI:
  1. Path-ID parsers duplicated across ≥7 BC handler packages
  2. `debugHTTPRequest` / `truncateRunes` / `debugHTTPOut` copied across BC `httplog*` helpers
  3. Repo-scoped `useQuery` boilerplate (~6 hooks share the same shape)

## ROI legend

| Score | Effort | Risk | Typical action |
| --- | --- | --- | --- |
| 8–10 | ≤2 days | Low–medium | Do next sprint |
| 5–7 | 2–4 days | Medium | One PR per extract |
| 1–4 | Docs-only or defer | Low | Batch with related PR |

**Formula:** `(clarity_gain × blast_radius) / (effort × risk)` — scores rounded 1–10.

**Phase rules (locked):**

| Phase | Do | Don't |
| --- | --- | --- |
| **4** | Extract on **3rd same-shape** occurrence | DRY two call sites |
| **6** | Interface/helper when **≥2 real production implementations** exist | Scaffold runners, hypothetical plugins |

---

## Verified clean (baseline for handoff)

| Concern | Choke point | Evidence |
| --- | --- | --- |
| Phase 3 policy catalogs | `readpolicy/`, `writepolicy/`, `queryInvalidation/`, `tasks/sync/` | [policy-roi.md](./policy-roi.md) complete ([#198](https://github.com/AlexsanderHamir/Hamix/pull/198)–[#203](https://github.com/AlexsanderHamir/Hamix/pull/203)) |
| Phase 5 god-file splits | Web/Go/test splits landed | [structural-patterns-roi.md](./structural-patterns-roi.md) complete ([#205](https://github.com/AlexsanderHamir/Hamix/pull/205)–[#215](https://github.com/AlexsanderHamir/Hamix/pull/215)) |
| Shared HTTP JSON errors | `pkgs/tasks/handlerhttp` (`WriteJSONError`, `invalidInputDetail`) | Package exists; BCs already import for JSON writes |
| Runner registry | `pkgs/agents/runner/registry/` | One production Cursor adapter; Claude Code remains scaffold |
| PlaceProvider | deleted / not revived | Out of scope for this train |
| Query staleTime tiers | `web/src/lib/queryPolicy.ts` | ADR-0025 |

---

## Findings (ranked) — Phase 4 first

### 1. Path-ID parsers copied across BCs — ROI 9/10 (High)

- **Location:** Same-shape `parseBoundedPathID` / `parseTaskPathID` (and cousins) in:
  - `pkgs/taskcore/handler/handler_path_ids.go`
  - `pkgs/tasks/handler/handler_path_ids.go` (**orphan** — only unit tests call it; production routes moved to BCs)
  - `pkgs/taskcycles/handler/http_helpers.go`
  - `pkgs/taskchecklist/handler/http_helpers.go`
  - `pkgs/taskevents/handler/http_helpers.go`
  - `pkgs/taskcompose/handler/http_helpers.go`
  - `pkgs/projects/handler/http_helpers.go` (`parsePathID`)
- **Issue:** Cap (128 bytes), trim, and `ErrInvalidInput` wrapping drift independently; composition-shell twin is dead in production.
- **Proposed change:** Shared `ParsePathID` / `ParseBoundedPathID` (and thin wrappers) in `pkgs/tasks/handlerhttp`; migrate BC call sites; delete or rehome orphan `tasks/handler/handler_path_ids.go` (+ test) with the extract.
- **Effort / risk:** 1–2 days; low–medium — contract tests for path 400s must stay green.
- **Evidence:** `rg -l 'parseBoundedPathID|parseTaskPathID|parsePathID' pkgs --glob '*.go'` → ≥7 packages; orphan callers only in `handler_path_ids_test.go` / comments.
- **Success signal:** One handlerhttp implementation; BC files import it; no duplicate `parseBoundedPathID` bodies under `pkgs/*/handler`.
- **PR slot:** PR3 `refactor/handlerhttp-path-ids`

### 2. HTTP I/O debug helpers duplicated — ROI 9/10 (High)

- **Location:** `debugHTTPRequest` ×**8**, `truncateRunes` ×**7**, `debugHTTPOut` ×**4** across:
  - `taskcore` / `tasks` `httplog_io.go`
  - `taskcycles` / `taskchecklist` / `taskevents` / `settings` / `repo` / `runners` `http_helpers.go` / `httplog.go`
- **Issue:** Identical `http.io` slog shape and rune truncation; copies diverge when observability fields change.
- **Proposed change:** Shared `DebugHTTPRequest`, `DebugHTTPOut`, `TruncateRunes` in `handlerhttp` (or thin `httplog` helper package colocated under handlerhttp).
- **Effort / risk:** 1–2 days; low — move-only; log field parity tests optional.
- **Evidence:** `rg -l 'func debugHTTPRequest' pkgs` → 8; `func truncateRunes` → 7; `func debugHTTPOut` → 4.
- **Success signal:** Single exported helpers; BC packages call them; local copies deleted.
- **PR slot:** PR4 `refactor/handlerhttp-httplog`

### 3. `parseBoundedLimit` + `firstQueryValue` (3 copies) — ROI 8/10 (High)

- **Location:**
  - `pkgs/projects/handler/handler_params.go` (live)
  - `pkgs/taskcompose/handler/http_helpers.go` (live)
  - `pkgs/tasks/handler/handler_query_params.go` (**dead** — no callers outside the file)
- **Issue:** Third copy is literal dead code in the composition shell; live pair will drift without a shared parse.
- **Proposed change:** Extract to `handlerhttp` (`ParseBoundedLimit` / `FirstQueryValue`); migrate projects + taskcompose; **delete** `handler_query_params.go`.
- **Effort / risk:** ≤1 day; low.
- **Evidence:** `rg 'func parseBoundedLimit|func firstQueryValue' pkgs` → exactly 3 definition sites; tasks/handler has no non-self references.
- **Success signal:** Dead file gone; two live BCs import shared helpers.
- **PR slot:** PR2 `refactor/handlerhttp-bounded-limit`

### 4. `invalidInputDetail` variants — ROI 7/10 (Medium)

- **Location:** Four private helpers:
  - `pkgs/tasks/handlerhttp/json.go` (canonical for shared JSON errors)
  - `pkgs/gitinventory/handler/http_helpers.go`
  - `pkgs/repo/handler/http_helpers.go`
  - `pkgs/settings/handler/http_helpers.go`
- **Issue:** Same “unwrap / stringify for 400 detail” pattern; drift risks inconsistent API error bodies.
- **Proposed change:** Export / share `InvalidInputDetail` from `handlerhttp`. **Do not** force `gitinventory` to import handlerhttp if that creates an import cycle — leave a documented local alias or thin duplicate for that package only.
- **Effort / risk:** 1 day; medium (cycle sensitivity).
- **Evidence:** `rg 'func invalidInputDetail|func InvalidInputDetail' pkgs` → 4.
- **Success signal:** repo + settings (+ gitinventory if cycle-safe) call one shared helper; API 400 detail strings unchanged in contract tests.
- **PR slot:** PR5 `refactor/handlerhttp-invalid-input` (after PR2–PR4 land)

### 5. Orphan composition-shell helpers — ROI 6/10 (Medium)

- **Location:** `pkgs/tasks/handler/handler_path_ids.go` (+ `_test.go`), `handler_query_params.go`.
- **Issue:** Left behind when handlers moved to BCs; confuse handoff (“is composition the source of truth?”).
- **Proposed change:** Delete with PR2 (query params) and PR3 (path IDs) — not a standalone PR.
- **Effort / risk:** bundled; none if done with extracts.
- **Evidence:** See findings #1 and #3.
- **Success signal:** Neither orphan file remains after Wave B/C extracts.
- **PR slot:** PR2 + PR3

---

## Findings (ranked) — Phase 6

### 6. Repo-scoped `useQuery` factory — ROI 8/10 (High)

- **Location:** Same shape (`repositoryId` + `enabled !== false && repositoryId.trim() !== ""` + queryKey/queryFn):
  - `web/src/hooks/useGlobalWorktrees.ts` (and worktrees twin re-exports where present)
  - `web/src/hooks/useGlobalBranches.ts`
  - `web/src/worktrees/hooks/useGlobalLiveWorktrees.ts`
  - `web/src/worktrees/hooks/useGlobalRepository.ts`
  - `web/src/worktrees/hooks/useWorktreeCheckoutStatus.ts`
  - `web/src/hooks/useProjectsByRepository.ts` (and `web/src/projects/` twin)
- **Issue:** Six+ near-identical hooks; enabling/trim rules easy to fork.
- **Proposed change:** Extract `useRepoScopedQuery` under `web/src/hooks/`; rewrite the listed hooks as thin wrappers.
- **Effort / risk:** 1 day; low — existing MSW/hook tests cover call sites.
- **Evidence:** Live `useGlobalWorktrees` body is 14 lines of shared pattern; `rg` on `repositoryId.trim()` under `web/src/**/use*.ts`.
- **Success signal:** Factory owns enable/trim; wrappers ≤15 lines each; web tests green.
- **PR slot:** PR6 `refactor/web-repo-scoped-query`

### 7. `Get(ctx, id) (*Task, error)` type aliases — ROI 7/10 (Medium)

- **Location:**
  - `pkgs/taskevents/contract.TaskGetter`
  - `pkgs/taskcycles/handler.TaskReader`
  - `internal/taskapi/agentworker.taskGetter`
  - Method also on `pkgs/taskcore/contract` CRUD + harness `internal/contract`
- **Issue:** Three+ identical one-method interfaces for the same store capability.
- **Proposed change:** Promote a single `TaskGetter` in `pkgs/taskcore/contract`; alias or replace the duplicates.
- **Effort / risk:** 1 day; medium — import graphs across BCs and agentworker.
- **Evidence:** `rg 'type TaskGetter|type TaskReader|type taskGetter' pkgs internal`.
- **Success signal:** One exported type; dependents alias or import it; no behavior change.
- **PR slot:** PR7 `refactor/taskcore-task-getter`

### 8. Duplicate `AgentWorkerControl` — ROI 7/10 (Medium)

- **Location:** Identical interfaces in `pkgs/tasks/handler/handler.go` and `pkgs/settings/handler/handler.go` (`CancelCurrentRun`, `Reload`, `ProbeRunner`).
- **Issue:** Two homes for the same worker-control surface; settings and composition can diverge.
- **Proposed change:** Single definition — prefer `settings/contract` or a small shared package both handlers import.
- **Effort / risk:** ≤1 day; low–medium.
- **Evidence:** `rg -A4 'type AgentWorkerControl interface' pkgs`.
- **Success signal:** One definition; both handlers import it; compile green.
- **PR slot:** PR8 `refactor/agent-worker-control`

### 9. Notify func aliases — ROI 6/10 (Medium)

- **Location:**
  - `pkgs/taskcore/handler.NotifyChangeFunc` — `func(typ ChangeType, id string)`
  - `pkgs/projects/handler.NotifyFunc` — same arity
  - `pkgs/settings/handler.NotifyFunc` — scopeless `func(typ ChangeType)`
- **Issue:** Near-duplicate SSE notify type names spread across BCs.
- **Proposed change:** Thin aliases or shared type names in `pkgs/tasks/realtime` (do not change SSE wire).
- **Effort / risk:** ≤1 day; low.
- **Evidence:** `rg 'type NotifyChangeFunc|type NotifyFunc' pkgs`.
- **Success signal:** BCs use realtime aliases; wire tests unchanged.
- **PR slot:** PR9 `refactor/realtime-notify-pathmap`

### 10. `HostPathMapper` ↔ `PathMap` packaging — ROI 5/10 (Low)

- **Location:** `gitinventory/handler.HostPathMapper` interface; concrete `tasks/handler.PathMap` with `DisplayHostPath`.
- **Issue:** Interface and env-backed mapper live in different packages without a compile-time assert that `*PathMap` implements the interface.
- **Proposed change:** `var _ HostPathMapper = (*PathMap)(nil)` (and optional colocation if layout allows without cycles).
- **Effort / risk:** ≤0.5 day; low.
- **Evidence:** `DisplayHostPath` on both; gitinventory deps inject mapper from composition.
- **Success signal:** Compile-time assert present; handlers green.
- **PR slot:** PR9 (batched with notify aliases)

### 11. Duplicate `PickupWake` interface — ROI 6/10 (Medium)

- **Location:**
  - `pkgs/taskcore/store/pickup_wake.go`
  - `internal/taskapi/storehooks/pickup_wake.go`
- **Issue:** Identical three-method interface in store BC and composition storehooks; dual ownership.
- **Proposed change:** One owner; alias the other (prefer storehooks as registry home or taskcore/store as domain hook — pick one in implementation PR).
- **Effort / risk:** ≤1 day; medium — composition + agents scheduler touch points.
- **Evidence:** `rg 'type PickupWake interface' pkgs internal`.
- **Success signal:** Single definition; the other file type-aliases; agents `PickupWakeScheduler` still satisfies it.
- **PR slot:** PR10 `chore/remaining-cleanup-close` (with docs retirement)

---

## Suggested implementation order (PR train)

| PR | Branch | Slice | Phase | Findings |
| --- | --- | --- | --- | --- |
| **1** | `audit/remaining-cleanup-roi` | This audit + wire README / cleanup-order | docs | — |
| **2** | `refactor/handlerhttp-bounded-limit` | Shared `parseBoundedLimit` / `firstQueryValue`; delete dead query params | 4 | #3, #5 (partial) |
| **3** | `refactor/handlerhttp-path-ids` | Shared path-ID parsers; migrate BCs; delete orphan path_ids | 4 | #1, #5 (partial) |
| **4** | `refactor/handlerhttp-httplog` | Shared DebugHTTPRequest/Out + TruncateRunes | 4 | #2 |
| **5** | `refactor/handlerhttp-invalid-input` | Shared InvalidInputDetail (respect gitinventory cycle) | 4 | #4 |
| **6** | `refactor/web-repo-scoped-query` | `useRepoScopedQuery` + rewrite ~6 hooks | 6 | #6 |
| **7** | `refactor/taskcore-task-getter` | Promote shared TaskGetter | 6 | #7 |
| **8** | `refactor/agent-worker-control` | Single AgentWorkerControl | 6 | #8 |
| **9** | `refactor/realtime-notify-pathmap` | Notify aliases + PathMap assert | 6 | #9, #10 |
| **10** | `chore/remaining-cleanup-close` | PickupWake collapse + **delete** cleanup-order + all `docs/audit/*-roi.md` (+ audit README); fix ADR broken links | 4/6 close | #11 + docs |

**Waves:** A = PR1 (sequential) → B = PR2–PR4 (parallel after PR1 merge) → C = PR5 → D = PR6→PR9 sequential → E = PR10.

---

## Not flagged (anti-DRY / anti-abstraction)

| Item | Reason |
| --- | --- |
| Phase 3 policy catalogs / readpolicy limits | Already centralized; re-DRY would fight ADR-0026 / ADR-0080 |
| gitinventory full handlerhttp mirror | Import-cycle risk; share only cycle-safe symbols (see #4) |
| 2-only `writeStoreError` mappers | Phase 4 requires **3rd** occurrence — do not extract |
| Mechanical `storefake` stubs | Pattern already split in Phase 5; codegen out of scope |
| MSW factories | Already choke-pointed under `web/src/test/` |
| Claude Code scaffold runner | Not a second **production** runner — fails Phase 6 ≥2-impl rule |
| Harness registry / adapterkit helpers | Justified only by scaffold; no abstract registry expansion |
| MemoryQueue-only `ReadyTaskQueue` | Single real impl — do not invent plug-in queue interface |
| PlaceProvider revival | Explicitly out of scope |
| Product UX / harness reliability features | Outside cleanup-order pass |

---

## See also

[cleanup-order.md](../cleanup-order.md) · [audit/README.md](./README.md) · [policy-roi.md](./policy-roi.md) · [structural-patterns-roi.md](./structural-patterns-roi.md)
