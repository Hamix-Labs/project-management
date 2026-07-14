# `pkgs/tasks/` shell audit — extractions vs folder organization

> Read-only audit (2026-07-13). Compares `pkgs/tasks/` to the finished BC pattern (`taskchecklist`, `taskcompose`, `taskcore`, …) and ranks what still earns an extract vs what needs **folder hygiene** inside the shell.

**Handoff goal:** Know when **not** to invent another bounded context, what leftover shell concerns still deserve a package move, and how to shrink the ~100-file flat `handler/` directory without fighting Go’s one-package-per-directory rule.

## Summary

- Items found: **9** actionable + explicit anti-findings
- High: 4 · Medium: 3 · Low: 2
- Top 3 by ROI:
  1. Relocate BC HTTP contract tests out of flat `pkgs/tasks/handler/` (~77 `*_test.go`)
  2. Finish SSE hub/stream move into `pkgs/tasks/realtime` (ADR-0020 / ADR-0070 deferred work)
  3. Repatriate middleware stack tests still living under `handler/`

**Headline verdict:** Domain BC extraction is **done**. Further “new `pkgs/<bc>/`” packages for leftover shell surfaces (bootstrap, health, RUM, git-binding glue) would **fight** the composition model. The real debt is **navigation**: one Go package directory holding production mux + ~5 BC wire-test themes + middleware leftovers.

## ROI legend

| Score | Effort | Risk | Typical action |
| --- | --- | --- | --- |
| 8–10 | ≤2 days | Low–medium | Do next sprint |
| 5–7 | 2–4 days | Medium | One PR per cluster |
| 1–4 | Docs-only or defer | Low | Batch with related PR |

**Formula:** `(clarity_gain × blast_radius) / (effort × risk)` — scores rounded 1–10.

**BC extract bar (locked by prior ADRs):** A sibling of `taskchecklist` / `taskcompose` needs (1) stable domain types, (2) owned persistence or a clear contract, (3) a cohesive HTTP resource group, (4) CI import gates. Aggregation across BCs stays in the shell.

---

## Baseline — what already matches the good pattern

| Concern | Home | Evidence |
| --- | --- | --- |
| Checklist / compose / cycles / events / task CRUD | `pkgs/taskchecklist`, `taskcompose`, `taskcycles`, `taskevents`, `taskcore` | Four-layer `{domain,contract,store,handler}` + `Register` from composition |
| Projects / git / settings / repo / runners | Sibling `pkgs/*` | ADR-0045–0049, ADR-0052 |
| Store facade retired | `internal/taskapi/composition` | ADR-0079 — no `pkgs/tasks/store` |
| Domain / contract hubs retired | BC packages own types | ADR-0060, ADR-0062 |
| HTTP helper DRY (in flight) | `pkgs/tasks/handlerhttp` | ADR-0077 + [remaining-cleanup-roi.md](./remaining-cleanup-roi.md) Phase 4 |
| Middleware already extracted | `pkgs/tasks/middleware/` | 16 `.go` files; README maps stack |
| Scheduling already extracted | `pkgs/tasks/scheduling/` | 10 `.go` files; domain docs |
| Policy packages nested | `handler/readpolicy/`, `handler/writepolicy/` | ADR-0026 |
| Test storefake nested | `handler/storefake/` | Phase 5 split done |

### Current `pkgs/tasks/` tree (scale)

| Path | ~Go files | Role |
| --- | --- | --- |
| `handler/` | **100** (+13 `storefake`) | Mux, SSE transport, bootstrap/health/RUM, BC `Register`, **~77 tests** |
| `postgres/` | 29 | Open + AutoMigrate orchestration + upgrade steps |
| `middleware/` | 16 | Outer HTTP stack |
| `apijson/` | 10 | Shared JSON/error helpers |
| `scheduling/` | 10 | Worker readiness predicates |
| `devsim/` | 9 | `HAMIX_SSE_TEST` |
| `realtime/` | 8 | Wire + `Publisher` (**hub still in handler**) |
| `calltrace/` / `logctx/` | 6–7 each | Observability kernels |
| `agentreconcile/` | 5 | **Tests only** (e2e) |
| `service/` | 4 | Bootstrap/git/retry use-cases |
| `handlerhttp/` | 8 | Shared handler HTTP primitives |
| `wire/` | 1 | `HandlerAPI` composition interface |

Sibling BC sizes for comparison: `taskcompose` ~25, `taskchecklist` ~34, `taskevents` ~33, `taskcycles` ~60, `taskcore` ~83. The outlier is **`tasks/handler` tests**, not missing domain packages.

---

## Findings (ranked)

### 1. Flat `handler/` is a test warehouse — ROI 9/10 (High)

- **Location:** `pkgs/tasks/handler/` — **100** `.go` files; **~77** `*_test.go`. Production surface is ~23 files; contract clusters: checklist (~5), compose/drafts (~7), cycles (~6), events (~5), taskcore/cross (~23), SSE (~7), middleware leftovers (~8), shell helpers (~16).
- **Issue:** Matches neither `taskchecklist`’s small package nor a maintainable composition shell. Agents and reviewers scroll a flat dump of `handler_http_*_contract_test.go`. Go requires **same-directory** files for package `handler`, so prefix naming alone does not create a browsable tree.
- **Proposed change:** Move **black-box** BC contract suites into theme packages under `internal/handlertest/` (already has `server.go`, health, commits):
  - `internal/handlertest/checklist/`
  - `internal/handlertest/compose/`
  - `internal/handlertest/cycles/`
  - `internal/handlertest/events/`
  - `internal/handlertest/taskcore/` (CRUD / list / patch / bootstrap pins)
  - Keep **whitebox** tests (unexported helpers, SSE hub internals, `storefake` wiring) in `pkgs/tasks/handler/`.
- **Effort / risk:** 2–3 days; medium — must keep `go-tests (tasks)` green; share `NewServer*` via existing handlertest helpers.
- **Evidence:** `Get-ChildItem pkgs/tasks/handler -Filter *.go` → 100; tests 77; `internal/handlertest` already has 6 Go files as the intended black-box home ([handler/README.md](../../pkgs/tasks/handler/README.md) §Tests).
- **Success signal:** `pkgs/tasks/handler` ≤ ~40 Go files (prod + whitebox + storefake); each handlertest theme package ≤ ~8 files; contract coverage unchanged.
- **PR slot:** PR2–PR4 (one theme wave each) after audit lands.

### 2. Finish SSE transport extract into `realtime` — ROI 8/10 (High)

- **Location:** Production: `handler/sse_hub.go` (311 lines), `sse_stream.go`, `sse_notify.go`. Partial extract: `pkgs/tasks/realtime/` (wire, coalesce, `Publisher`).
- **Issue:** ADR-0020 kept the hub in `handler`; ADR-0070 deferred “optional realtime extract until a second consumer.” `realtime` already has a second-style consumer (`agentworker` via `Publisher`). Hub concurrency + HTTP stream still inflate the composition package and dig against the documented domain article.
- **Proposed change:** Move `SSEHub` + stream framing into `pkgs/tasks/realtime` (or `realtime/hub` subpackage if cycle risk). `handler` keeps thin `streamEvents` method + `notify*` that call `writepolicy` then `hub.Publish`. Update `cmd/taskapi` construct site; keep wire JSON stable.
- **Effort / risk:** 2 days; medium — SSE lossless / trigger contract tests must stay green ([docs/domain/sse-hub.md](../domain/sse-hub.md)).
- **Evidence:** ADR-0020 table; ADR-0070 deferred note; `sse_hub.go` yellow/red border vs CODE_STANDARDS handler limits; `realtime/README.md` states transport still in handler.
- **Success signal:** No ring-buffer / subscribe implementation under `handler/`; handler SSE files are glue ≤ ~100 lines each; agentworker still depends on `realtime.Publisher` only.
- **PR slot:** PR5 `refactor/realtime-hub-extract`

### 3. Middleware tests still under `handler/` — ROI 8/10 (High)

- **Location:** `handler/api_auth_test.go`, `idempotency*_test.go`, `max_body_test.go`, `rate_limit_accesslog_test.go`, `security_headers_test.go`, `stack_test.go`, `observability_test.go` (8 files). Implementations already live in `pkgs/tasks/middleware/`.
- **Issue:** Split ownership: code in `middleware/`, behavior pins in `handler/`. Confuses “is the stack part of the mux package?”
- **Proposed change:** Move these tests to `pkgs/tasks/middleware/` (black-box against `middleware.Stack` / `With*`). Leave handler tests that only assert mux registration of health/SSE.
- **Effort / risk:** ≤1 day; low — import `calltrace.Path` the same way production does.
- **Evidence:** 4 tests already under `middleware/`; 8 named middleware tests remain under `handler/`.
- **Success signal:** Zero middleware-implementation tests under `handler/`; `middleware` README lists the test map.
- **PR slot:** PR1 (can ship with audit) or PR2 first wave

### 4. `agentreconcile` is pkgs-packaged test-only code — ROI 7/10 (Medium)

- **Location:** `pkgs/tasks/agentreconcile/` (~5 files; large e2e: `agent_real_cursor_e2e_test.go` 654 lines, `agentworker_e2e_test.go` 426).
- **Issue:** Importable `pkgs/` path for non-production tests. [agent-map.md](../agent-map.md) already labels it “SQLite integration tests; not production code.” Peers keep e2e under `internal/`.
- **Proposed change:** Move to `internal/taskapi/agentreconcile` (or `internal/agentreconcile`). No public API change.
- **Effort / risk:** ≤1 day; low–medium if CI path filters reference the old import path.
- **Evidence:** Package has only tests + `doc.go`; agent-map row.
- **Success signal:** No `pkgs/tasks/agentreconcile`; CI still runs the e2e group.
- **PR slot:** PR6

### 5. Stale shell docs still describe deleted `store` / `domain` — ROI 6/10 (Medium)

- **Location:** `pkgs/tasks/doc.go` (still lists `domain`, `store`); `.cursor/rules/backend/go/layout.mdc` composition table; [agent-map.md](../agent-map.md) Persistence row (`pkgs/tasks/store/`, store README link); various domain docs still cite pre-extract paths.
- **Issue:** New contributors look for a facade that ADR-0079 deleted; wastes scout time and invites “restore the monolith” PRs.
- **Proposed change:** Doc-only PR: rewrite `tasks/doc.go` to list real subpackages; fix agent-map Persistence → `internal/taskapi/composition` + BC stores + `pkgs/tasks/postgres`; align layout.mdc composition shell table.
- **Effort / risk:** ≤0.5 day; none.
- **Evidence:** `Test-Path pkgs/tasks/store` → false; ADR-0079.
- **Success signal:** Zero primary docs claim `pkgs/tasks/store` or `pkgs/tasks/domain` as live homes.
- **PR slot:** PR1 with this audit (or immediately after)

### 6. `postgres/` migrate steps are flat but prefix-organized — ROI 5/10 (Low–Medium)

- **Location:** `pkgs/tasks/postgres/` — 29 files; many `migrate_*.go` (+ tests) for git tree, compose payloads, removals, seeds.
- **Issue:** Directory is large but **not** a missing BC — it is the composition migrate orchestrator (ADR-0079). Subfolders **cannot** stay `package postgres` (Go rule). A nested `postgres/migrate` package is possible but adds import indirection for little domain clarity.
- **Proposed change:** Prefer **docs map** in `postgres/README.md` (group by theme: git-tree, compose-payload, removals, seeds) + keep `migrate_*` prefixes. Optional later: extract only if a second binary opens a second migrate entrypoint.
- **Effort / risk:** ≤0.5 day docs; higher if mechanical package split.
- **Evidence:** ADR-0079 move of migrate hub; file list is already verb-prefixed.
- **Success signal:** README theme table; newcomers find the right `migrate_*` file in &lt;1 minute.
- **PR slot:** batch with #5

### 7. Git-binding / compose normalize stay shell glue — ROI 4/10 (Low)

- **Location:** `handler_task_git_binding.go`, `handler_compose_normalize.go`, `handler_taskcore_wire.go`, `service/{bootstrap,git,retry}.go`.
- **Issue:** Cross-BC validation (project/worktree/@-mentions) used by taskcore create and compose. Looks “extractable,” but it **orchestrates** `gitinventory` + `gitwork` + taskcore — classic composition.
- **Proposed change:** **Do not** create `pkgs/taskgitbinding`. Keep thin; optionally push pure helpers into `pkgs/gitwork` only when a second caller appears (Phase 6 rule). Ensure `service/` doc matches reality (HTTP-agnostic bootstrap already used by `handler_bootstrap.go`).
- **Effort / risk:** n/a unless a second caller appears.
- **Evidence:** BC extract bar; ADR-0066 bootstrap contracts; `service` imported only from composition shell paths.
- **Success signal:** No new BC ADR for glue; wire callbacks remain in `handler_taskcore_*`.

### 8. Bootstrap / health / RUM as a “system” BC — ROI 3/10 (Low) — **defer / reject**

- **Location:** `handler_bootstrap.go`, `handler_health.go`, `handler_system_health.go`, `handler_rum.go` (~261 lines RUM).
- **Issue:** Tempting package (`pkgs/system`), but bootstrap **aggregates** many BC contracts (ADR-0066). Health/RUM are process operators, not a bounded domain with store/handler/domain quartet.
- **Proposed change:** Keep in shell. If RUM grows past red zone, split **files** (`handler_rum_json.go`) — not a sibling BC.
- **Effort / risk:** Avoid: high rewiring / low clarity gain.
- **Evidence:** ADR-0066; composition shell ownership ADR-0070.
- **Success signal:** No `pkgs/system` without a second independent consumer and owned persistence.

### 9. Scheduling / apijson / calltrace / logctx / handlerhttp / wire — ROI n/a — **verified clean**

Already appropriately sized platform packages. Continue Phase 4 DRY **into** `handlerhttp` / `apijson` ([remaining-cleanup-roi.md](./remaining-cleanup-roi.md)); do not re-home them as domain BCs.

---

## Anti-findings (do not extract)

| Candidate | Why not |
| --- | --- |
| New domain BC for leftover `handler_*` | Domain verticals already extracted (ADR-0045–0059+) |
| Revive `pkgs/tasks/store` facade | Explicitly deleted — ADR-0079 |
| Nested directories inside `package handler` | Impossible in Go without new packages |
| Split `postgres` into per-BC migrate packages | Model lists already BC-local; hub must stay ordered (FK-safe) |
| Extract `devsim` to top-level pkgs | Dev-only; fine under `tasks/` |
| “Organize” by renaming only | Prefixes help search; they do not fix a 77-test flat dir |

---

## Suggested implementation order

| PR | Branch | Slice | Findings |
| --- | --- | --- | --- |
| **1** | `audit/tasks-shell-roi` | This audit + stale `doc.go` / agent-map / layout notes (#5, #6 docs) | docs |
| **2** | `refactor/handlertest-middleware-repatriate` | Middleware tests → `middleware/`; first handlertest theme if room | #3, #1 partial |
| **3** | `refactor/handlertest-bc-contracts-a` | checklist + compose contract suites → `internal/handlertest/*` | #1 |
| **4** | `refactor/handlertest-bc-contracts-b` | cycles + events + taskcore/cross suites | #1 |
| **5** | `refactor/realtime-hub-extract` | SSE hub/stream → `realtime`; thin handler glue | #2 |
| **6** | `chore/agentreconcile-internal` | Move e2e package under `internal/` | #4 |

**Train rule:** Do not start PR5 until PR3–PR4 land (SSE tests are easier to move once black-box suites have a stable handlertest home). Phase 4/6 DRY ([remaining-cleanup-roi.md](./remaining-cleanup-roi.md)) can proceed **in parallel** — it shrinks helpers, this train shrinks **directory shape**.

---

## Relationship to prior audits

| Audit | Overlap | Diff |
| --- | --- | --- |
| [structural-patterns-roi.md](./structural-patterns-roi.md) | God-file **line** splits (done) | This audit targets **file-count / package placement**, not LOC within one file |
| [remaining-cleanup-roi.md](./remaining-cleanup-roi.md) | `handlerhttp` DRY | This audit does not re-rank path-ID / httplog extracts |
| [policy-roi.md](./policy-roi.md) | readpolicy / writepolicy | Policy choke points stay nested under handler — no move |

---

## See also

[cleanup-order.md](../cleanup-order.md) · [audit/README.md](./README.md) · [ADR-0070](../adr/ADR-0070-taskapi-shell-ownership.md) · [ADR-0079](../adr/ADR-0079-facade-deletion.md) · [ADR-0020](../adr/ADR-0020-realtime-sse-layout.md) · [pkgs/taskchecklist/README.md](../../pkgs/taskchecklist/README.md) · [pkgs/tasks/handler/README.md](../../pkgs/tasks/handler/README.md)
