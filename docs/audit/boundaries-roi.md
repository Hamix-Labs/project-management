# Boundaries & contracts audit — ROI-ranked findings

> Read-only audit (July 2026). No code changed. Maps to [cleanup-order.md](../cleanup-order.md) Phase 1.

## Summary

- Items found: **12** (High: 3, Medium: 7, Low: 2)
- Top 3 by ROI:
  1. ~~`tasks/` imports `projects/` and `worktrees/` directly~~ **done** ([#147](https://github.com/AlexsanderHamir/Hamix/pull/147))
  2. ~~Create/bulk task mutations bypass guarded write + mutation guard~~ **done** ([#148](https://github.com/AlexsanderHamir/Hamix/pull/148))
  3. `projects/` and `worktrees/` hooks invalidate cache outside `tasks/sync/`

## ROI legend

| Score | Effort | Risk | Typical action |
| --- | --- | --- | --- |
| 8–10 | ≤1 day (or high strategic value) | Low–medium | Do next sprint |
| 5–7 | 1–3 days | Medium | Batch with related work |
| 1–4 | >3 days or high risk | Defer or spike first |

**Formula:** `(clarity_gain × blast_radius) / (effort × risk)` — scores rounded 1–10.

---

## Findings (ranked)

### 1. Web vertical coupling (`tasks/` → `projects/` + `worktrees/`) — ROI 9/10 (High) — **Status: done (2026-07-08)**

- **PR:** [#147](https://github.com/AlexsanderHamir/Hamix/pull/147) (`cleanup/boundaries-vertical-decouple`)
- **Boundary violated:** Vertical slices — no cross-feature imports ([cleanup-order §1](../cleanup-order.md), `CODE_STANDARDS.mdc` Part 4)
- **Resolution:** Cross-cutting git-assignment and project-picker code promoted to `web/src/lib/`, `web/src/hooks/`, and `web/src/components/`. `tasks/` imports only the inner ring; `check-code-standards.ps1` enforces the boundary.
- **New locations (production):**
  - **lib:** [composeGitAssignment.ts](../../web/src/lib/composeGitAssignment.ts), [gitWorktreeRegistration.ts](../../web/src/lib/gitWorktreeRegistration.ts), [ensureRepositoriesRegistered.ts](../../web/src/lib/ensureRepositoriesRegistered.ts), [projectQueryKeys.ts](../../web/src/lib/projectQueryKeys.ts)
  - **hooks:** [useComposeGitAssignment.ts](../../web/src/hooks/useComposeGitAssignment.ts), [useProject.ts](../../web/src/hooks/useProject.ts), [useProjects.ts](../../web/src/hooks/useProjects.ts), [useProjectContextPromptBinding.ts](../../web/src/hooks/useProjectContextPromptBinding.ts), global git hooks under [hooks/](../../web/src/hooks/)
  - **components:** [ProjectSelect](../../web/src/components/project/ProjectSelect.tsx), [GitWorktreeIcons](../../web/src/components/git/GitWorktreeIcons.tsx), [RepositorySetupPrompt](../../web/src/components/git/RepositorySetupPrompt.tsx), [ProjectContextPicker](../../web/src/components/project-context/ProjectContextPicker.tsx)
  - **tasks consumers:** [TaskCreateAssignmentFields.tsx](../../web/src/tasks/components/task-create-modal/fields/TaskCreateAssignmentFields.tsx), [TaskCreateModalsLayer.tsx](../../web/src/tasks/pages/TaskCreateModalsLayer.tsx), [decideSyncFrame.ts](../../web/src/tasks/sync/decideSyncFrame.ts), [useTaskGitBinding.ts](../../web/src/tasks/hooks/useTaskGitBinding.ts), [useTaskCreateEntryActions.ts](../../web/src/tasks/create/hooks/useTaskCreateEntryActions.ts)
- **Deleted:** `WorktreeSelector.tsx` (zero consumers; superseded by `TaskCreateAssignmentFields`)
- **Evidence (exit gate):** `rg 'from ["'']@/(projects|worktrees)/' web/src/tasks` → zero matches

### 2. Create/bulk task mutations bypass guarded write — ROI 9/10 (High) — **Status: done (2026-07-08)**

- **PR:** [#148](https://github.com/AlexsanderHamir/Hamix/pull/148)
- **Boundary violated:** Mutations policy — guarded writes + mutation guard ([ADR-0025](../adr/ADR-0025-frontend-data-coherence.md), [guardedTaskWrite.ts](../../web/src/tasks/mutations/guardedTaskWrite.ts))
- **Resolution:** Create and template-instantiate mutations seed cache under `beginGuardedTaskWrite`; bulk schedule/delete use `beginBulkTaskMutationGuard` (ADR M2) with optional optimistic list surgery. Broad `taskQueryKeys.all` invalidation replaced by [`invalidateTaskListAndStats`](../../web/src/tasks/mutations/invalidateTaskListCoherence.ts) (`listRoot` + `stats`).
- **New / updated modules:**
  - [optimisticTaskList.ts](../../web/src/tasks/mutations/optimisticTaskList.ts) — `applyCreatedTaskToCache`, list insert/remove/schedule patches
  - [mutationGuard.ts](../../web/src/tasks/sync/mutationGuard.ts) — `beginBulkTaskMutationGuard` / `endBulkTaskMutationGuard`
  - [useTaskCreateMutations.ts](../../web/src/tasks/create/hooks/useTaskCreateMutations.ts) — guarded `createMutation` + `instantiateTemplatesMutation`
  - [useBulkTaskMutation.ts](../../web/src/tasks/components/task-list/bulk/useBulkTaskMutation.ts) — bulk guard session + narrowed invalidation
- **Out of scope (ADR M3):** draft/template autosave mutations still invalidate `drafts()` / `templates()` only — no task guard.
- **Evidence (exit gate):** `rg 'taskQueryKeys\.all' web/src/tasks/create/hooks/useTaskCreateMutations.ts web/src/tasks/components/task-list/bulk` → zero matches; `rg beginGuardedTaskWrite web/src/tasks/create` → matches create path; bulk guard tests green.

### 3. Project/worktree cache invalidation outside `tasks/sync/` — ROI 8/10 (High) — **Status: done (2026-07-08)**

- **PR:** [#149](https://github.com/AlexsanderHamir/Hamix/pull/149) (`cleanup/boundaries-cache-invalidation`)
- **Boundary violated:** Sync policy — cache effects centralized in `tasks/sync/` ([cleanup-order §1](../cleanup-order.md), [web.md](../web.md) §Task sync)
- **Resolution:** Shared catalog in `web/src/lib/queryInvalidation/`; `projects/mutations/` and `worktrees/mutations/` own production invalidation; `decideSyncFrame` uses the same project scopes for SSE. CI gate blocks stray `invalidateQueries` in project/worktree pages.
- **New locations:**
  - **lib:** [queryInvalidation/](../../web/src/lib/queryInvalidation/) — `decideProjectInvalidationKeys`, `decideGitInvalidationKeys`, `applyQueryInvalidations`
  - **projects:** [mutations/](../../web/src/projects/mutations/) — create/delete/patch/context hooks
  - **worktrees:** [mutations/](../../web/src/worktrees/mutations/) — global and legacy git hooks
  - **sync:** [decideSyncFrame.ts](../../web/src/tasks/sync/decideSyncFrame.ts) — project frames via catalog
- **ADR:** [ADR-0044](../adr/ADR-0044-query-invalidation-catalog.md)
- **Evidence (exit gate):** `rg 'invalidateQueries' web/src/projects web/src/worktrees --glob '!*.test.*' --glob '!**/mutations/**'` → zero matches

### 4. `handler_task_events` publishes hint-only `task_updated` — ROI 7/10 (Medium) — **Status: done (2026-07-08)**

- **PR:** [#150](https://github.com/AlexsanderHamir/Hamix/pull/150)
- **Boundary violated:** Write publish policy ([ADR-0026](../adr/ADR-0026-backend-data-coherence.md) — `task_updated` should be enriched when task row changes; [writepolicy/publish_policy.go](../../pkgs/tasks/handler/writepolicy/publish_policy.go))
- **Resolution:** Added dedicated `task_event_changed` SSE type with `event_seq` for audit-log mutations. `PATCH /tasks/{id}/events/{seq}` now publishes `notifyTaskEventChanged` instead of misleading `task_updated`. Web sync invalidates `taskQueryKeys.eventsRoot` only (immediate, no detail/list storm).
- **New locations:**
  - **realtime:** [wire.go](../../pkgs/tasks/realtime/wire.go) — `TaskEventChanged`, `EventSeq`
  - **handler:** [sse_notify.go](../../pkgs/tasks/handler/sse_notify.go) — `notifyTaskEventChanged`; [handler_task_events.go](../../pkgs/tasks/handler/handler_task_events.go) — publish swap
  - **web:** [sseInvalidate.ts](../../web/src/tasks/task-query/sseInvalidate.ts), [decideSyncFrame.ts](../../web/src/tasks/sync/decideSyncFrame.ts), [taskQueryKeys.ts](../../web/src/lib/taskQueryKeys.ts) — `eventsRoot`
- **Evidence (exit gate):** `rg 'notifyChange\(TaskUpdated' pkgs/tasks/handler/handler_task_events.go` → zero; `rg 'task_event_changed' pkgs/tasks/realtime web/src` → wire + handler + sync matches

### 5. `handler_settings` bypasses `notifyChange` helper — ROI 7/10 (Medium) — **Status: done (2026-07-08)**

- **PR:** [#151](https://github.com/AlexsanderHamir/Hamix/pull/151)
- **Boundary violated:** Write publish — centralized notify path ([ADR-0026](../adr/ADR-0026-backend-data-coherence.md) S1)
- **Resolution:** Added `notifyScopelessChange` for id-less hint-only frames (`settings_changed`, `agent_run_cancelled`). `handler_settings.go` now routes both publish sites through the helper; CI gate blocks direct `h.hub.Publish` outside `sse_notify.go`.
- **New locations:**
  - **writepolicy:** [publish_policy.go](../../pkgs/tasks/handler/writepolicy/publish_policy.go) — `ScopelessHintChangeTypes`, `IsScopelessHint`
  - **handler:** [sse_notify.go](../../pkgs/tasks/handler/sse_notify.go) — `notifyScopelessChange`; [handler_settings.go](../../pkgs/tasks/handler/handler_settings.go) — publish swap
  - **ci:** [check-go.sh](../../scripts/check-go.sh), [check-go.ps1](../../scripts/check-go.ps1) — `sse publish boundary` step
- **Evidence (exit gate):** `rg 'h\.hub\.Publish' pkgs/tasks/handler -g '*.go' -g '!*_test.go'` → only `sse_notify.go`

### 6. Checklist verify-commands patch uses ad-hoc invalidation — ROI 7/10 (Medium) — **Status: done (2026-07-08)**

- **PR:** [#152](https://github.com/AlexsanderHamir/Hamix/pull/152) (`cleanup/boundaries-checklist-verify-guard`)
- **Boundary violated:** Sync/mutations policy
- **Resolution:** Added `buildUpdateChecklistVerifyCommandsMutationOptions` via `buildGuardedChecklistMutation`; `submitEditChecklistCriterionForm` routes verify-only and combined edits through `updateChecklistVerifyCommandsMutation`. Verify edits now get mutation guard, optimistic `verify_commands` cache surgery, and `invalidateTaskChecklistQueries` (checklist + detail).
- **Location (was):** [useTaskDetailChecklist.ts](../../web/src/tasks/checklist/hooks/useTaskDetailChecklist.ts) — raw `patchChecklistItemVerifyCommands` + manual checklist-only invalidation
- **Evidence (exit gate):** `rg 'patchChecklistItemVerifyCommands' web/src/tasks --glob '!*.test.*'` → only `mutationFn` in `useTaskDetailChecklist.ts`; `rg 'buildGuardedChecklistMutation' web/src/tasks/checklist/hooks/useTaskDetailChecklist.ts` → four mutation builders (add, text, verify, delete)

### 7. `TaskEventDetailPage` ad-hoc cache updates — ROI 6/10 (Medium) — **Status: done (2026-07-08)**

- **Resolution:** `buildPatchTaskEventUserResponseMutationOptions` applies mutation guard, updates event detail cache, invalidates `eventsRoot` only (pairs with #4 `task_event_changed`).
- **Evidence (exit gate):** `rg 'taskQueryKeys\.detail' web/src/tasks/pages/TaskEventDetailPage.tsx` → zero

### 8. `useTaskDetailScheduling` ad-hoc list invalidation — ROI 6/10 (Medium) — **Status: done (2026-07-08)**

- **Resolution:** Guarded scheduling mutations; hint ops use `invalidateTaskDetailCoherence`; tags/milestone use optimistic detail + `invalidateTaskListAndStats`.
- **Evidence (exit gate):** `rg 'taskQueryKeys\.all' web/src/tasks/hooks/useTaskDetailScheduling.ts` → zero

### 9. Postgres data migrations still query via `domain.*` GORM models — ROI 6/10 (Medium) — **Status: done (2026-07-08)**

- **Resolution:** Data migrations use `store/model` for GORM; domain values via `FromDomain*` mappers.
- **Evidence (exit gate):** `rg 'Model\(&domain\.(Git|Project)' pkgs/tasks/postgres/migrate_*.go` → zero on targeted files

### 10. Close ADR-0039 as Accepted — ROI 5/10 (Medium) — **Status: done (2026-07-08)**

- **Resolution:** ADR-0039 status **Accepted**.
- **Evidence (exit gate):** ADR status Accepted; `rg 'gorm:' pkgs/tasks/domain` → no matches

### 11. `lib/queryClient` coupled to `tasks/sync` — ROI 5/10 (Medium) — **Status: done (2026-07-08)**

- **Resolution:** `QUERY_POLICY` and SSE connection flag promoted to `web/src/lib/`; tasks modules re-export.
- **Evidence (exit gate):** `rg 'from "@/tasks/' web/src/lib/queryClient.ts` → no matches

### 12. Bootstrap cold-start `setQueryData` — ROI 4/10 (Low) — **Status: done (2026-07-08)**

- **Resolution:** `seedBootstrapCache` in `tasks/sync/`; documented in web.md §Cold start.
- **Evidence (exit gate):** `rg 'seedBootstrapCache' web/src` → sync module + useBootstrap

---

## Verified clean

| Boundary | Check | Result |
| --- | --- | --- |
| Go `domain/` purity | `rg 'gorm:\|gorm\.io\|net/http' pkgs/tasks/domain` | **Clean** — only `database/sql/driver` in [sqltypes.go](../../pkgs/tasks/domain/sqltypes.go) for enum `Scan`/`Value` |
| Go `handler/` no DB | `rg 'database/sql\|gorm\.io' pkgs/tasks/handler` | **Clean** |
| Schema migrate targets store models | [postgres.go](../../pkgs/tasks/postgres/postgres.go) `model.AutoMigrateAll` | **Clean** |
| `fetch` only in `api/` | `rg '\bfetch\s*\(' web/src` | **Clean** — sole call in [api/shared.ts](../../web/src/api/shared.ts) |
| Task CRUD/retry enriched SSE | `notifyTaskChanged` / `notifyTaskUpdatedEnriched` in handlers | **Clean** for checklist, CRUD, gate (with inline task), compose create |
| Guarded writes on detail mutations | `useTaskPatchFlow`, `useTaskDeleteFlow`, `useTaskDetailMutations`, checklist add/update/delete/verify | **Clean** |
| Reverse vertical imports | `rg 'from "@/tasks/' web/src/projects web/src/worktrees` | **Clean** — one-way coupling only |
| Bootstrap list limit parity | `readpolicy.BootstrapListLimit` (20) vs `TASK_LIST_PAGE_SIZE` (20) | **Aligned** |

---

## Suggested fix order

1. **#10** — Close ADR-0039 (unblocks agent clarity, zero code risk)
2. **#5** — Settings notify helper (small backend win)
3. **#4** + **#7** — Event response SSE/cache (pair backend + frontend)
4. **#2** — Create/bulk guarded mutations (high traffic)
5. ~~**#6**~~ + **#8** — Remaining task hook policy gaps
6. **#3** — Project/worktree invalidation modules
7. **#1** — Vertical decoupling (largest structural slice; may span multiple PRs)
8. **#9** — Postgres migrate model alignment
9. **#11** — Lib/query policy promotion
10. **#12** — Document bootstrap exception

---

## Not flagged (intentional per ADR / design)

| Item | Reason |
| --- | --- |
| `notifyChange(TaskGateChanged)` / `TaskDependencyChanged` | Hint-only per [ADR-0026](../adr/ADR-0026-backend-data-coherence.md) S3 |
| `notifyChange(Project*)` | Hint-only project events |
| `notifyTaskChanged` with inline `domain.Task` on PATCH/create | Enriched publish without extra store round-trip |
| `app/App.tsx` imports `@/tasks/*` | Shell wires SSE + task routes — expected |
| React Query `.refetch()` in pages | Not `fetch()` — not an api/ boundary violation |
| `useGitMutations.ts` (project-scoped) | Legacy stack; tracked in [dead-code-roi.md](./dead-code-roi.md) #1 — delete with legacy git, not boundary fix |

---

## Re-run evidence (agent commands)

```powershell
# Domain / handler / fetch
rg 'gorm:|gorm\.io|net/http' pkgs/tasks/domain
rg 'database/sql|gorm\.io' pkgs/tasks/handler
rg '\bfetch\s*\(' web/src

# Vertical imports (production)
rg 'from "@/projects/' web/src/tasks --glob '!*.test.*'
rg 'from "@/worktrees/' web/src/tasks --glob '!*.test.*'

# Cache policy (production)
rg 'invalidateQueries|setQueryData' web/src --glob '!*.test.*' --glob '!**/test/**'

# Backend publish
rg 'h\.hub\.Publish|notifyChange\(' pkgs/tasks/handler
```

After implementing fixes: `.\scripts\check.ps1` from repo root.
