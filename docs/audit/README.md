# Codebase quality audit — consolidated index

Read-only audits completed 2026-07-05. Each report ranks findings by ROI (1–10). No code was changed during investigation.

## Reports

| Report | Focus | Items | High ROI (≥8) |
| --- | --- | --- | --- |
| [duplication-roi.md](./duplication-roi.md) | Shared code for duplicated logic | 15 | 6 |
| [abstractions-roi.md](./abstractions-roi.md) | Interfaces and dependency inversion | 12 | 4 |
| [logic-simplification-roi.md](./logic-simplification-roi.md) | Simpler control flow without new layers | 12 | 5 |
| [lld-patterns-roi.md](./lld-patterns-roi.md) | SOLID, layering, design patterns | 12 | 4 |
| [dead-code-roi.md](./dead-code-roi.md) | Unused exports, legacy stacks, pointless lines | 15 | 5 |

## Top 10 overall (cross-report)

Ranked by composite ROI considering effort, risk, and lines/clarity impact. Primary report owns each item; others may cross-reference.

| Rank | ROI | Finding | Primary report | Effort |
| --- | --- | --- | --- | --- |
| 1 | 10 | Remove legacy project-scoped git stack (~900–1,100 lines, zero UI consumers) | [dead-code](./dead-code-roi.md#1-legacy-project-scoped-git-stack--roi-1010-high) | 1–2 days |
| 2 | 10 | Domain layer persistence split (GORM out of `domain/`) | [lld-patterns](./lld-patterns-roi.md#1-domain-layer-is-persistence-aware--roi-1010-high) | 3–5 days |
| 3 | 9 | Consolidate `mustCreateTask` test helpers (5 variants, 40+ sites) | [duplication](./duplication-roi.md#1-mustcreatetask--post-tasks-test-helpers--roi-910-high) | 4–6 hours |
| 4 | 9 | Centralize settings cursor-model query keys | [abstractions](./abstractions-roi.md#1-centralize-settings-runnermodel-query-keys--roi-910-high) | Low |
| 5 | 9 | Git store → facade/internal pattern (`reconcile_git.go` Red) | [lld-patterns](./lld-patterns-roi.md#2-git-persistence-bypasses-facadeinternal-pattern--roi-910-high) | 3–4 days |
| 6 | 9 | Reconcile worktree matching simplification | [logic-simplification](./logic-simplification-roi.md#1-reconcile-worktree-row-matching--triple-skip-paths--roi-910-high) | 4–6 hours |
| 7 | 9 | Checklist mutation optimistic pipeline factory | [logic-simplification](./logic-simplification-roi.md#2-checklist-mutations--three-optimistic-pipelines--roi-910-high) | 4–6 hours |
| 8 | 9 | Unify global vs project git HTTP handlers | [duplication](./duplication-roi.md#2-git-handlers-global-vs-project-scoped-parallel-surfaces--roi-910-high) | 2–3 days |
| 9 | 9 | Handler `*store.Store` → composed interfaces | [abstractions](./abstractions-roi.md#2-handler-storestore--composed-store-interface--roi-910-high) | Multi-PR |
| 10 | 9 | Delete subtask-era CSS with no DOM (~120–150 lines) | [dead-code](./dead-code-roi.md#2-subtask-era-css-with-no-dom--roi-910-high) | 1–2 hours |

## Quick wins (≤1 day, low risk)

Do these first for immediate payoff without architectural risk:

1. Migrate `agentreconcile` git seeding to `gittest` — [duplication #3](./duplication-roi.md)
2. Template handler `normalizeComposePayload` helper — [duplication #10](./duplication-roi.md)
3. Task create modal footer collapse (`isEdit`) — [logic-simplification #4](./logic-simplification-roi.md)
4. Reconcile `stillLive` → `liveByPath` lookup — [logic-simplification #8](./logic-simplification-roi.md)
5. `getTaskCycle` use `assertCycleBelongsToTask` — [logic-simplification #9](./logic-simplification-roi.md)
6. Harness timeout: use or delete `withOptionalRunTimeout` — [logic-simplification #12](./logic-simplification-roi.md)
7. Delete `taskDescendantCount` stub — [dead-code #3](./dead-code-roi.md)
8. `GitWorktreeResolver` narrow interface — [abstractions #4](./abstractions-roi.md)
9. Route registration split in `handler.go` — [lld-patterns #3](./lld-patterns-roi.md)

## Structural themes (multi-PR)

These recur across reports — batch related PRs:

| Theme | Reports | Suggested sequence |
| --- | --- | --- |
| **Git vertical cleanup** | dead-code #1, duplication #2/#4, lld #2/#6, abstractions #9/#10 | Delete legacy → extract helpers → facade/internal → optional service layer |
| **Handler test consolidation** | duplication #1/#6/#13, lld #9 | Shared test helpers → contract harness → optional handlertest cycle break |
| **Web god-file splits** | lld #5/#8/#11, logic #10 | Cycle detail page → create modal → task detail page |
| **Store/handler DIP** | abstractions #2/#5/#8, lld #4/#7 | Query keys → GitWorktreeResolver → handler store slices → worker Store → fakes |
| **Select/widget unification** | duplication #5 | Spike a11y parity → merge CustomSelect + SettingsSelect |

## Overlap rules

| If found in… | Primary owner | Secondary may reference |
| --- | --- | --- |
| Same code copied 3+ times | [duplication](./duplication-roi.md) | logic-simplification |
| Missing interface on hot path | [abstractions](./abstractions-roi.md) | lld-patterns |
| Over-nested conditionals | [logic-simplification](./logic-simplification-roi.md) | — |
| God file / SRP violation | [lld-patterns](./lld-patterns-roi.md) | abstractions |
| Unused export / pass-through barrel | [dead-code](./dead-code-roi.md) | duplication |

## ROI legend (shared)

| Score | Effort | Risk | Typical action |
| --- | --- | --- | --- |
| 8–10 | ≤1 day (or high strategic value) | Low–medium | Do next sprint |
| 5–7 | 1–3 days | Medium | Batch with related work |
| 1–4 | >3 days or high risk | Defer or spike first |

## Verification after implementing fixes

Run the local check script from repo root:

```powershell
.\scripts\check.ps1
```

For scoped changes: `-GoOnly` or `-WebOnly`. See [CONTRIBUTING.md](../CONTRIBUTING.md#before-you-open-a-pr).
