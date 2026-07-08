# Codebase quality audit — consolidated index

Read-only audits completed 2026-07-05. Each report ranks findings by ROI (1–10). No code was changed during investigation.

**Cleanup phase order:** [cleanup-order.md](../cleanup-order.md) — agents pick tasks from this index but follow that priority stack (boundaries → delete → policy → DRY → structure → abstractions).

## Reports

| Report | Focus | Items | High ROI (≥8) |
| --- | --- | --- | --- |
| [boundaries-roi.md](./boundaries-roi.md) | Layer boundaries & contract seams (Phase 1) | 12 | 3 |
| ~~duplication-roi.md~~ | Shared code for duplicated logic | 15 (completed) | — |
| ~~abstractions-roi.md~~ | Interfaces and dependency inversion | 12 (completed) | — |
| ~~logic-simplification-roi.md~~ | Simpler control flow without new layers | 12 (completed) | — |
| ~~lld-patterns-roi.md~~ | SOLID, layering, design patterns | 12 (completed) | — |
| [dead-code-roi.md](./dead-code-roi.md) | Unused exports, legacy stacks, pointless lines | 15 | 5 |

## Top 10 overall (cross-report)

Ranked by composite ROI considering effort, risk, and lines/clarity impact. Primary report owns each item; others may cross-reference.

| Rank | ROI | Finding | Primary report | Effort |
| --- | --- | --- | --- | --- |
| 1 | 10 | Remove legacy project-scoped git stack (~900–1,100 lines, zero UI consumers) | [dead-code](./dead-code-roi.md#1-legacy-project-scoped-git-stack--roi-1010-high) | 1–2 days |
| 2 | 9 | ~~Web vertical coupling (`tasks/` → `projects/` + `worktrees/`)~~ **done** | [boundaries](./boundaries-roi.md#1-web-vertical-coupling-tasks--projects--worktrees--roi-910-high--status-done-2026-07-08) | 2–3 days |
| 3 | 9 | ~~Create/bulk mutations bypass guarded write + mutation guard~~ **done** | [boundaries](./boundaries-roi.md#2-createbulk-task-mutations-bypass-guarded-write--roi-910-high--status-done-2026-07-08) | 1–2 days |
| 4 | 9 | Delete subtask-era CSS with no DOM (~120–150 lines) | [dead-code](./dead-code-roi.md#2-subtask-era-css-with-no-dom--roi-910-high) | 1–2 hours |
| 5 | ~~8~~ | ~~Project/worktree cache invalidation outside `tasks/sync/`~~ **done** | [boundaries](./boundaries-roi.md#3-projectworktree-cache-invalidation-outside-taskssync--roi-810-high--status-done-2026-07-08) | 2–3 days |
| 6 | 8 | Delete `taskDescendantCount()` stub | [dead-code](./dead-code-roi.md#3-taskdescendantcount-always-returns-0--roi-810-high) | 15 min |
| 7 | ~~7~~ | ~~`handler_task_events` hint-only `task_updated` on event append~~ **done** | [boundaries](./boundaries-roi.md#4-handler_task_events-publishes-hint-only-task_updated--roi-710-medium--status-done-2026-07-08) | 4–8 hours |
| 8 | ~~7~~ | ~~`handler_settings` bypasses `notifyChange` helper~~ **done** | [boundaries](./boundaries-roi.md#5-handler_settings-bypasses-notifychange-helper--roi-710-medium--status-done-2026-07-08) | 1–2 hours |
| 9 | 7 | Checklist verify-commands ad-hoc invalidation | [boundaries](./boundaries-roi.md#6-checklist-verify-commands-patch-uses-ad-hoc-invalidation--roi-710-medium) | 2–4 hours |
| 10 | 5 | Close ADR-0039 as Accepted (store/model landed) | [boundaries](./boundaries-roi.md#10-close-adr-0039-as-accepted--roi-510-medium) | 30 min |

## Quick wins (≤1 day, low risk)

Do these first for immediate payoff without architectural risk:

1. ~~Migrate `agentreconcile` git seeding to `gittest`~~ — done
2. ~~Template handler `normalizeComposePayload` helper~~ — done
3. ~~Task create modal footer collapse (`isEdit`)~~ — done
4. ~~Reconcile `stillLive` → `liveByPath` lookup~~ — done
5. ~~`getTaskCycle` use `assertCycleBelongsToTask`~~ — done
6. ~~Harness timeout: use or delete `withOptionalRunTimeout`~~ — done
7. Delete `taskDescendantCount` stub — [dead-code #3](./dead-code-roi.md)
8. ~~`GitWorktreeResolver` narrow interface~~ — done
9. ~~Route registration split in `handler.go`~~ — done

## Structural themes (multi-PR)

These recur across reports — batch related PRs:

| Theme | Reports | Suggested sequence |
| --- | --- | --- |
| **Git vertical cleanup** | dead-code #1, duplication #2/#4, lld #2/#6 | Delete legacy → extract helpers → facade/internal |
| **Web vertical boundaries** | boundaries #1, #3 | Decouple tasks↔projects/worktrees → unify cache invalidation |
| **Frontend mutation policy** | boundaries #2, #6–#8 | Guarded create/bulk → checklist/scheduling gaps |
| **Handler test consolidation** | duplication (done), lld #9 | Contract harness done; tasktestserver extracted |
| **Web god-file splits** | lld #5/#8/#11 | Cycle detail page → create modal → task detail page |
| **Store/handler DIP** | abstractions (done) | Query keys → GitWorktreeResolver → handler store slices → worker Store → fakes |
| **Select/widget unification** | duplication (done) | Combobox keyboard layer shared; full BaseCombobox deferred |

## Overlap rules

| If found in… | Primary owner | Secondary may reference |
| --- | --- | --- |
| Same code copied 3+ times | duplication (completed) | logic-simplification |
| Missing interface on hot path | abstractions (completed) | lld-patterns |
| Over-nested conditionals | logic-simplification (completed) | — |
| God file / SRP violation | abstractions (completed) | lld-patterns (completed) |
| Unused export / pass-through barrel | [dead-code](./dead-code-roi.md) | duplication |
| Layer / contract seam violation | [boundaries](./boundaries-roi.md) | dead-code (if delete fixes seam) |

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
