# Web SPA

Vite + React client under `web/`. All `fetch` calls live in `web/src/api/`; responses are parsed through typed parsers before use. Wire enums mirrored in `web/src/types/` and `web/src/constants/` are documented in [ADR-0035](./adr/ADR-0035-cross-stack-constant-ownership.md).

| | |
| --- | --- |
| **Applies to** | `web/` SPA routes, data layer, and task UI |
| **Audience** | Frontend contributors and agents on web-only or full-stack slices |
| **Prerequisite** | [architecture.md](./architecture.md) for API/SSE context; [api.md](./api.md) for contracts |

## In this article

- [Routes](#routes)
- [Board view](#board-view)
- [Timeline view](#timeline-view)
- [Cold start](#cold-start)
- [Task sync (SSE cache coherence)](#task-sync-sse-cache-coherence)
- [Task create flow](#task-create-flow)
- [Query policy](#query-policy)
- [Project/worktree mutation invalidation](#projectworktree-mutation-invalidation)
- [Task detail — execution cycles](#task-detail--execution-cycles)
- [CSS ownership](#css-ownership)
- [See also](#see-also)

## Routes

| Path | Module | Notes |
| --- | --- | --- |
| `/` | `web/src/tasks/` | Task home — list (default), board (`?view=board`), or timeline (`?view=timeline`) |
| `/templates` | `web/src/tasks/` | Saved task templates (search, batch instantiate) |
| `/drafts` | `web/src/tasks/` | Saved create-task drafts |
| `/projects` | `web/src/projects/` | Project list |
| `/projects/:id` | `web/src/projects/` | Project detail |
| `/repositories` | `web/src/worktrees/` | Registered git repositories (nav: **Repositories**) |
| `/settings` | `web/src/settings/` | App settings |
| `/tasks/:id` | `web/src/tasks/pages/` | Task detail |

Primary nav links: Tasks, Templates, Drafts, Projects, Repositories (Settings is header gear). Register a repo via `/repositories` or `/repositories?register=1` — see [domain/worktrees-and-branches.md](./domain/worktrees-and-branches.md). Legacy `/worktrees` and `/worktrees/:repositoryId` redirect to `/repositories`.

## Board view

Task Home supports a read-only Kanban **Board** alongside the table **List** (`/?view=board`). Columns are workflow buckets — Backlog (`ready`, `on_hold`), In Progress (`running`), Verification (`review`), Needs Attention (`blocked`, `failed`). **Done tasks are never shown** (volume accumulates; the board is for active execution).

Data loads through `fetchActiveTasksForBoard` (keyset walk of `GET /tasks`, max page size 200) into `taskQueryKeys.board()` under `listRoot()`, so existing SSE / optimistic list invalidation refreshes the board. Caps: 500 active tasks or 10 pages scanned — then a truncation banner. There is no drag-and-drop; status changes come from the execution engine. A future `exclude_status=done` (or allowlist) on `GET /tasks` would avoid scanning Done rows.

## Timeline view

Task Home also supports a read-only **Timeline** (`/?view=timeline`) — a chronological project-activity feed. A date-range dropdown maps to the `since` query parameter sent to `GET /tasks/activity` so the server filters before the client groups by calendar day. The Timeline also reuses the board’s client-side filters (priority, project, tag, title search) against joined task fields on each activity row (`task_priority`, `task_project_id`, `task_tags`, `task_title`).

**Live data via `GET /tasks/activity`** — the three event types surfaced are `status_changed`, `phase_failed`, and `approval_granted`. The client maps each to a `TimelineEvent` via `activityMapper.ts`, then `groupTimelineEvents` buckets them into Today / Yesterday / N days ago groups.

Timeline data is fetched through `useTasksActivity` (hook under `tasks/hooks/`), keyed at `taskQueryKeys.activityRoot()` under `["tasks","activity"]`. The key is invalidated alongside `cycleFailuresRoot()` on every SSE task/event flush in `decideFlushBatch.ts`.

Deep links: when a feed event carries a `seq`, clicking the timestamp navigates to `/tasks/{id}/events/{seq}`. The task ref chip always links to the task detail page.

Data implementation: `web/src/api/tasks.read.ts` (`getTaskActivity`), parser at `web/src/api/parseTaskApiActivity.ts`. Distinct from the per-task audit timeline on task detail.

## Cold start

`web/src/app/hooks/useBootstrap.ts` seeds TanStack Query from `GET /v1/bootstrap`. List/stats/settings queries stay disabled until bootstrap settles (success, unavailable, or failure); then per-resource GETs run only when the cache was not seeded.

**Sync policy exception:** bootstrap calls `seedBootstrapCache` in [`tasks/sync/seedBootstrapCache.ts`](../web/src/tasks/sync/seedBootstrapCache.ts) — intentional direct `setQueryData`, not SSE-driven.

## Task sync (SSE cache coherence)

Live task UI cache policy lives in [`web/src/tasks/sync/`](../web/src/tasks/sync/). Read order:

1. [ADR-0022](./adr/ADR-0022-task-sync-policy.md) — Decide vs Apply boundaries
2. `decideSyncFrame.ts` — per-frame schedule, suppression, enrichment effects
3. `decideFlushBatch.ts` — debounced invalidation targets
4. `taskSyncCoordinator.ts` — pending state + debounce wiring consumed by `useTaskEventStream`

Enriched `task_updated` with terminal status (`done`, `failed`) patches task detail via `setQueryData` and **immediately** invalidates list + stats via `decideTaskInvalidationKeys({ scope: "listStats" })` (`terminalTaskStatus.ts` + `applySyncEffects.ts`) so status badges update without waiting for the debounced cycle flush.

**Sync-owned vs catalog-owned invalidation:** mutation hooks and guarded writes use [`decideTaskInvalidationKeys`](../web/src/lib/queryInvalidation/decideTaskInvalidationKeys.ts) (ADR-0080). SSE sync keeps frame orchestration in `decideSyncFrame.ts` (per-type effects, suppression, enrichment) and debounced flush targets in `decideFlushBatch.ts`, which reuses the catalog for `listStats` keys via `syncListStatsInvalidationKeys()`. Empty pending flush still invalidates `taskQueryKeys.all` (broad resync) — that scope stays sync-owned, not in the mutation catalog. Malformed/unknown SSE frames use `schedule: "ignore"` so they do not debounce into an empty-pending broad invalidate.

Wire decode stays in `web/src/tasks/task-query/sseInvalidate.ts`. Event catalog and operator tuning: [domain/sse-hub.md](./domain/sse-hub.md).

## Task create flow

Create-task policy and hook composition live in [`web/src/tasks/create/`](../web/src/tasks/create/). Read order:

1. [ADR-0024](./adr/ADR-0024-task-create-flow-slice.md) — Decide vs Apply boundaries, invariants I1–I7
2. `decideCreateEntry.ts` — `openCreateModal` routing (loading / error / drafts / fresh)
3. `composePayload.ts`, `validateCreateForm.ts`, `draftPayload.ts`, `buildCreateMutationInput.ts` — shared compose payload, validation, and wire shapes
4. [`lib/composeGitAssignment.ts`](../web/src/lib/composeGitAssignment.ts) + [`hooks/useComposeGitAssignment.ts`](../web/src/hooks/useComposeGitAssignment.ts) — pure git assignment reducer and orchestration hook (repo / project / worktree); see [ADR-0043](./adr/ADR-0043-compose-git-assignment.md)
5. `mapCreateFlowViewModel.ts` — flat public return shape for `useTasksApp`
6. `hooks/useTaskCreateFlow.ts` — composer; shim at `web/src/tasks/hooks/useTaskCreateFlow.ts`

Modal UI stays in `web/src/tasks/components/task-create-modal/` for V1. **`composeTarget`** (`task` | `template`) and **`composeOperation`** (`create` | `edit`) drive one modal for task create/edit and template save/edit. Templates list and batch create: `web/src/tasks/pages/TaskTemplatesPage.tsx` (`GET /task-templates`, `POST /task-templates/instantiate`). API client: `web/src/api/taskTemplates.ts`. Race contracts: `useTasksApp.test.tsx`.

## Query policy

TanStack Query staleTime tiers live in [`web/src/lib/queryPolicy.ts`](../web/src/lib/queryPolicy.ts) (re-exported from [`tasks/queryPolicy.ts`](../web/src/tasks/queryPolicy.ts)). SSE connection policy: [`web/src/lib/queryConnectionPolicy.ts`](../web/src/lib/queryConnectionPolicy.ts). Read order:

1. [ADR-0025](./adr/ADR-0025-frontend-data-coherence.md) — query tiers, mutation guard M1–M3, render isolation
2. `queryPolicy.ts` — `QUERY_POLICY` constants consumed by `queryClient`, list hooks, prefetch
3. [`tasks/mutations/`](../web/src/tasks/mutations/) — guarded optimistic task writes (patch/close/checklist, create/instantiate cache seed, bulk schedule/close)
4. [`tasks/checklist/`](../web/src/tasks/checklist/) — detail checklist mutations with guard
5. [`tasks/app/TasksAppProvider.tsx`](../web/src/tasks/app/TasksAppProvider.tsx) — narrow selector hooks

## Project/worktree mutation invalidation

Project and git writes invalidate React Query through a shared catalog — not inline in pages.

1. [ADR-0044](./adr/ADR-0044-query-invalidation-catalog.md) — catalog scopes and vertical mutation ownership
2. [`lib/queryInvalidation/`](../web/src/lib/queryInvalidation/) — `decideProjectInvalidationKeys`, `decideGitInvalidationKeys`, `applyQueryInvalidations`
3. [`projects/mutations/`](../web/src/projects/mutations/) — project create/delete/patch/context hooks
4. [`worktrees/mutations/`](../web/src/worktrees/mutations/) — global and legacy git hooks
5. [`tasks/sync/decideSyncFrame.ts`](../web/src/tasks/sync/decideSyncFrame.ts) — SSE `project` / `project_context` frames use the same project scopes

## Task detail — execution cycles

Expanded cycle rows in `TaskCyclesPanel` load `GET /tasks/{id}/cycles/{cycleId}/verdicts`. When the worker indexed git commits for the cycle, the panel shows a repo → branch breadcrumb and commit rows (`git_context`, `commits[]`) with **status badges** (`eligible`, `observed`, …) above the per-criterion verdict list.

The task detail page also loads **`GET /tasks/{id}/commits`** via `TaskCommitsPanel` / `useTaskCommits` — task-wide commit history deduped by SHA, refetched on `task_cycle_changed` SSE. Clicking a commit row navigates to **`/tasks/{id}/commits/{sha}`** (`TaskCommitDiffPage`), which loads **`GET /repo/diff?worktree_id=&sha=`** (task `worktree_id`) with GitHub-style summary stats, syntax-highlighted hunks (refractor + `react-diff-view`), unified/split toggle, file navigator, and collapsible large files. Parsers: `web/src/api/parseTaskApiCycles.ts`; types: `web/src/types/cycle.ts`. See [domain/cycle-commits.md](./domain/cycle-commits.md).

## CSS ownership

Global chrome and feature CSS load from [`web/src/app/App.css`](../web/src/app/App.css) via barrels under [`web/src/app/styles/`](../web/src/app/styles/). Design tokens live in `styles/tokens/` (`app-design-tokens.css`).

| Area | Barrel / entry | Section partials |
| --- | --- | --- |
| Tokens | `app-design-tokens.css` | `tokens/app-design-tokens-{foundation,light,dark}.css` |
| Base shell | `app-base.css` | `base/` |
| Task create / fields | `app-task-create-and-fields.css` | `task-create/` |
| Task list / mentions | `app-task-list-and-mentions.css` | `task-list/`, `mentions/` |
| Task detail | `app-task-detail.css` | `task-detail/` |
| Task timeline | `app-task-timeline.css` | `timeline/` |
| Modals | `app-modals.css` | `modals/` |
| **Projects** | `app-projects.css` | `projects/` (`app-project-list`, `detail`, `context`, `sections`) |
| **Repositories / worktrees** | `app-worktrees.css` | `worktrees/` (`app-repositories-list`, `app-repository-detail`) |
| Settings | `settings/settings.css` (imported from `SettingsPage`) | `settings-*.css` colocated with the settings vertical |
| Shared UI primitives | `components/ui/ui.css` (via `App.css`) | `Button` / `Badge` (`.ui-btn`, `.ui-badge`) |

Prefer editing the section partial that owns the surface; keep barrel `@import` order stable (cascade matches the former monoliths). Settings and toast/picker CSS outside `app/styles/` are still in the CSS standards scan.

## See also

- [guide.md](./guide.md) — documentation layers and learning paths
- [README.md](./README.md) — doc index by topic
- [agent-map.md](./agent-map.md) — web code paths
- [CONTRIBUTING.md](../CONTRIBUTING.md) — setup and PR checklist
- [domain/sse-hub.md](./domain/sse-hub.md) — SSE event catalog and operator tuning
- [ADR-0035](./adr/ADR-0035-cross-stack-constant-ownership.md) — cross-stack constant ownership and intentional omissions
