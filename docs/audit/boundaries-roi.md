# Boundaries & contracts audit — ROI-ranked findings

> Read-only audit (July 2026). No code changed. Maps to [cleanup-order.md](../cleanup-order.md) Phase 1.

## Summary

- Items found: **12** (High: 3, Medium: 7, Low: 2)
- Top 3 by ROI:
  1. `tasks/` imports `projects/` and `worktrees/` directly (create flow + sync)
  2. Create/bulk task mutations bypass guarded write + mutation guard
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

### 2. Create/bulk task mutations bypass guarded write — ROI 9/10 (High)

- **Boundary violated:** Mutations policy — guarded writes + mutation guard ([ADR-0025](../adr/ADR-0025-frontend-data-coherence.md), [guardedTaskWrite.ts](../../web/src/tasks/mutations/guardedTaskWrite.ts))
- **Location:**
  - [useTaskCreateMutations.ts](../../web/src/tasks/create/hooks/useTaskCreateMutations.ts) — `createMutation`, draft/template saves: raw `invalidateQueries` on success (lines 66–68, 91–163); no `beginGuardedTaskWrite` / `endGuardedTaskWrite`
  - [useBulkTaskMutation.ts](../../web/src/tasks/components/task-list/bulk/useBulkTaskMutation.ts) — bulk status/priority/reschedule: `invalidateQueries` for `taskQueryKeys.all` + `stats()` (lines 77–78); no mutation guard
- **Issue:** SSE echo suppression and optimistic coherence are bypassed on high-traffic write paths (create task, bulk ops). Race with enriched `task_updated` events can cause stale list rows or double refetch.
- **Proposed change:** Route create/bulk mutations through `runGuardedMutation` / `beginGuardedTaskWrite`; align invalidation with `applySyncEffects` patterns or rely on SSE where enriched events already cover the write.
- **Effort / risk / blast radius:** 1–2 days; medium risk (create + bulk UX); 2 hook files + tests.
- **Evidence:** `rg beginGuardedTaskWrite web/src/tasks/create` → no matches. `useTaskPatchFlow.ts`, `useTaskDeleteFlow.ts`, `useTaskDetailMutations.ts` already use guarded writes — inconsistent policy.

### 3. Project/worktree cache invalidation outside `tasks/sync/` — ROI 8/10 (High)

- **Boundary violated:** Sync policy — cache effects centralized in `tasks/sync/` ([cleanup-order §1](../cleanup-order.md), [web.md](../web.md) §Task sync)
- **Location:**
  - [ProjectListPage.tsx](../../web/src/projects/ProjectListPage.tsx), [ProjectContextPanel.tsx](../../web/src/projects/ProjectContextPanel.tsx), [ProjectSettingsPanel.tsx](../../web/src/projects/ProjectSettingsPanel.tsx), [ProjectDetailPage.tsx](../../web/src/projects/ProjectDetailPage.tsx), [RepositoryProjectsSection.tsx](../../web/src/worktrees/components/RepositoryProjectsSection.tsx)
  - [useGitMutations.ts](../../web/src/worktrees/hooks/useGitMutations.ts) (legacy), [useGlobalGitMutations.ts](../../web/src/worktrees/hooks/useGlobalGitMutations.ts)
- **Issue:** Non-task verticals call `queryClient.invalidateQueries` directly after mutations. No shared invalidation table; project/git writes can miss keys that `decideSyncFrame` knows about (e.g. `projectsByRepo`).
- **Proposed change:** Add `projects/mutations/` + `worktrees/mutations/` mirroring `tasks/mutations/`, or extend `decideSyncFrame` / a shared `web/src/lib/queryInvalidation.ts` with project/git mutation effects. Wire SSE `project_*` hints through the same coordinator where applicable.
- **Effort / risk / blast radius:** 2–3 days; medium risk; 6+ production files.
- **Evidence:** `rg invalidateQueries web/src/projects web/src/worktrees` (exclude `*.test.*`).

### 4. `handler_task_events` publishes hint-only `task_updated` — ROI 7/10 (Medium)

- **Boundary violated:** Write publish policy ([ADR-0026](../adr/ADR-0026-backend-data-coherence.md) — `task_updated` should be enriched when task row changes; [writepolicy/publish_policy.go](../../pkgs/tasks/handler/writepolicy/publish_policy.go))
- **Location:** [handler_task_events.go](../../pkgs/tasks/handler/handler_task_events.go) L235 — `h.notifyChange(TaskUpdated, id)` after `AppendTaskEventResponseMessage`
- **Issue:** Appending a user response to `task_events` does **not** mutate the `tasks` row, but emits id-only `task_updated`, forcing clients to refetch task detail unnecessarily. `EnrichedTaskChangeEvent(TaskUpdated)` is true in policy, but this path skips enrichment.
- **Proposed change:** Remove the publish (event detail returned in HTTP body), or add a dedicated hint type (e.g. `task_event_changed`) and teach `decideSyncFrame` to invalidate `taskQueryKeys.events` only.
- **Effort / risk / blast radius:** 4–8 hours; low–medium risk (event detail page + SSE); 1 handler + sync test.
- **Evidence:** Handler calls `notifyChange` not `notifyTaskUpdatedEnriched`; store method is event append only.

### 5. `handler_settings` bypasses `notifyChange` helper — ROI 7/10 (Medium)

- **Boundary violated:** Write publish — centralized notify path ([ADR-0026](../adr/ADR-0026-backend-data-coherence.md) S1)
- **Location:** [handler_settings.go](../../pkgs/tasks/handler/handler_settings.go) L184 (`SettingsChanged`), L311 (`AgentRunCancelled`) — direct `h.hub.Publish(...)`
- **Issue:** Settings mutations skip `notifyChange`, duplicating hub access and bypassing any future publish instrumentation in [sse_notify.go](../../pkgs/tasks/handler/sse_notify.go).
- **Proposed change:** Replace with `h.notifyChange(SettingsChanged, "")` and `h.notifyChange(AgentRunCancelled, taskID)` (or add typed helpers). `SettingsChanged` is correctly hint-only per `writepolicy.HintOnlyChangeTypes`.
- **Effort / risk / blast radius:** 1–2 hours; low risk; 1 file + settings contract test.
- **Evidence:** `rg 'h\.hub\.Publish' pkgs/tasks/handler` → only `handler_settings.go` and `sse_notify.go`.

### 6. Checklist verify-commands patch uses ad-hoc invalidation — ROI 7/10 (Medium)

- **Boundary violated:** Sync/mutations policy
- **Location:** [useTaskDetailChecklist.ts](../../web/src/tasks/checklist/hooks/useTaskDetailChecklist.ts) L507–511 — `patchChecklistItemVerifyCommands` then manual `invalidateQueries` for checklist key
- **Issue:** Most checklist mutations use `beginGuardedTaskWrite` + optimistic pipeline; verify-command-only edits bypass the shared mutation options and invalidate outside `applySyncEffects`.
- **Proposed change:** Add a `useMutation` for verify-command patch using the same `build*MutationOptions` factory as other checklist ops, or route invalidation through guarded write.
- **Effort / risk / blast radius:** 2–4 hours; low risk; 1 hook file.
- **Evidence:** `beginGuardedTaskWrite` used in same file for other mutations; this path is the exception.

### 7. `TaskEventDetailPage` ad-hoc cache updates — ROI 6/10 (Medium)

- **Boundary violated:** Sync/mutations policy
- **Location:** [TaskEventDetailPage.tsx](../../web/src/tasks/pages/TaskEventDetailPage.tsx) L41–45 — `setQueryData` + `invalidateQueries` on `taskQueryKeys.detail` after `patchTaskEventUserResponse`
- **Issue:** Page-level mutation success handler duplicates cache policy; no mutation guard for SSE echo on task detail.
- **Proposed change:** Extract to `tasks/mutations/` helper or rely on backend fix (#4) so SSE drives invalidation consistently.
- **Effort / risk / blast radius:** 2–4 hours; low risk; 1 page + test.
- **Evidence:** `rg guardedTaskWrite web/src/tasks/pages` → no matches.

### 8. `useTaskDetailScheduling` ad-hoc list invalidation — ROI 6/10 (Medium)

- **Boundary violated:** Sync/mutations policy
- **Location:** [useTaskDetailScheduling.ts](../../web/src/tasks/hooks/useTaskDetailScheduling.ts) L19 — `invalidateQueries({ queryKey: taskQueryKeys.all })` after schedule patch
- **Issue:** Scheduling mutation invalidates full task list prefix without guarded write; patch flow in `useTaskPatchFlow.ts` already centralizes this pattern.
- **Proposed change:** Reuse guarded patch helper from `useTaskPatchFlow` or shared scheduling mutation module.
- **Effort / risk / blast radius:** 2–3 hours; low risk; 1 hook.
- **Evidence:** Compare with [useTaskPatchFlow.ts](../../web/src/tasks/hooks/useTaskPatchFlow.ts) which uses `beginGuardedTaskWrite`.

### 9. Postgres data migrations still query via `domain.*` GORM models — ROI 6/10 (Medium)

- **Boundary violated:** Domain vs persistence ([ADR-0039](../adr/ADR-0039-domain-persistence-separation.md))
- **Location:** [migrate_repo_root_to_git_repository.go](../../pkgs/tasks/postgres/migrate_repo_root_to_git_repository.go), [migrate_git_common_dir.go](../../pkgs/tasks/postgres/migrate_git_common_dir.go), [migrate_seed_worktree_branch_tree.go](../../pkgs/tasks/postgres/migrate_seed_worktree_branch_tree.go) — `db.Model(&domain.GitRepository{})` etc.
- **Issue:** Runtime store uses [store/model](../../pkgs/tasks/store/model/); one-off SQL migrations still treat domain structs as GORM models. Domain is json-only today, but migrate scripts reintroduce persistence coupling at the seam.
- **Proposed change:** Use `store/model` types + mappers in migrate scripts (follow [migrate_compose_payload_worktree_test.go](../../pkgs/tasks/postgres/migrate_compose_payload_worktree_test.go) pattern with `model.FromDomain*`).
- **Effort / risk / blast radius:** 1 day; medium risk (schema migration); 3–5 migrate files.
- **Evidence:** `rg 'domain\.(Git|Project|Task)' pkgs/tasks/postgres` vs `postgres.go` L88 `model.AutoMigrateAll` for schema.

### 10. Close ADR-0039 as Accepted — ROI 5/10 (Medium)

- **Boundary violated:** Documentation drift
- **Location:** [ADR-0039](../adr/ADR-0039-domain-persistence-separation.md) status **Proposed**; implementation landed in `pkgs/tasks/store/model/`
- **Issue:** Agents and contributors may re-plan a completed split. [store/model/doc.go](../../pkgs/tasks/store/model/doc.go) documents the intended boundary.
- **Proposed change:** Update ADR status to **Accepted**; note residual migrate-script gap (#9) as follow-up, not blocker.
- **Effort / risk / blast radius:** 30 minutes; none; docs only.
- **Evidence:** `rg 'gorm:' pkgs/tasks/domain` → no matches; `postgres.Migrate` uses `model.AutoMigrateAll`.

### 11. `lib/queryClient` coupled to `tasks/sync` — ROI 5/10 (Medium)

- **Boundary violated:** Shell should not depend on feature sync internals
- **Location:** [queryClient.ts](../../web/src/lib/queryClient.ts) imports `isSseLiveForQueries` from `@/tasks/sync/connectionPolicy`; [queryPersist.ts](../../web/src/lib/queryPersist.ts) imports `QUERY_POLICY` + task/project query keys
- **Issue:** Global query client configuration is pinned to tasks-domain policy. Adding a second major vertical or testing query client in isolation requires tasks tree.
- **Proposed change:** Promote `connectionPolicy` + stale-time tiers to `web/src/lib/queryPolicy.ts` (or merge with existing shared constants); keep tasks-specific keys in tasks.
- **Effort / risk / blast radius:** 4–8 hours; low risk; 2 lib files + app bootstrap tests.
- **Evidence:** `rg 'from "@/tasks/' web/src/lib`.

### 12. Bootstrap cold-start `setQueryData` — ROI 4/10 (Low)

- **Boundary violated:** Sync policy (surface-level)
- **Location:** [useBootstrap.ts](../../web/src/app/hooks/useBootstrap.ts) L71–81
- **Issue:** App shell seeds TanStack Query directly instead of going through `tasks/sync/`. Intentional for cold start but undocumented as an exception.
- **Proposed change:** Document in [web.md](../web.md) §Cold start as allowed exception; optionally wrap in `seedBootstrapCache(queryClient, payload)` in `tasks/sync/` for discoverability.
- **Effort / risk / blast radius:** 1–2 hours; low risk; docs + optional thin wrapper.
- **Evidence:** Comments in `useBootstrap.ts` L62–67 already explain parity with query keys; no bug identified.

---

## Verified clean

| Boundary | Check | Result |
| --- | --- | --- |
| Go `domain/` purity | `rg 'gorm:\|gorm\.io\|net/http' pkgs/tasks/domain` | **Clean** — only `database/sql/driver` in [sqltypes.go](../../pkgs/tasks/domain/sqltypes.go) for enum `Scan`/`Value` |
| Go `handler/` no DB | `rg 'database/sql\|gorm\.io' pkgs/tasks/handler` | **Clean** |
| Schema migrate targets store models | [postgres.go](../../pkgs/tasks/postgres/postgres.go) `model.AutoMigrateAll` | **Clean** |
| `fetch` only in `api/` | `rg '\bfetch\s*\(' web/src` | **Clean** — sole call in [api/shared.ts](../../web/src/api/shared.ts) |
| Task CRUD/retry enriched SSE | `notifyTaskChanged` / `notifyTaskUpdatedEnriched` in handlers | **Clean** for checklist, CRUD, gate (with inline task), compose create |
| Guarded writes on detail mutations | `useTaskPatchFlow`, `useTaskDeleteFlow`, `useTaskDetailMutations`, checklist add/update | **Clean** (except finding #6) |
| Reverse vertical imports | `rg 'from "@/tasks/' web/src/projects web/src/worktrees` | **Clean** — one-way coupling only |
| Bootstrap list limit parity | `readpolicy.BootstrapListLimit` (20) vs `TASK_LIST_PAGE_SIZE` (20) | **Aligned** |

---

## Suggested fix order

1. **#10** — Close ADR-0039 (unblocks agent clarity, zero code risk)
2. **#5** — Settings notify helper (small backend win)
3. **#4** + **#7** — Event response SSE/cache (pair backend + frontend)
4. **#2** — Create/bulk guarded mutations (high traffic)
5. **#6** + **#8** — Remaining task hook policy gaps
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
