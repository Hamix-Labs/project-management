# Code path index

Repository paths grouped by subsystem. Read only the rows relevant to your task.

| | |
| --- | --- |
| **Applies to** | Locating code when editing the repo |
| **Audience** | Contributors and agents (after [AGENTS.md](../AGENTS.md) scoped path) |
| **Prerequisite** | Pick a scoped path in [AGENTS.md](../AGENTS.md) first |

## In this article

- [Overview](#overview)
- [Backend](#backend)
- [Web](#web)
- [Infra and test helpers](#infra-and-test-helpers)
- [See also](#see-also)

## Overview

> **Tip** — Open one to three rows below. API contracts: [api.md](./api.md). System overview: [architecture.md](./architecture.md).

## Backend

| Area | Path | Purpose | Deep dive |
| --- | --- | --- | --- |
| HTTP API + SSE | `pkgs/tasks/handler/` | SSE hub, bootstrap, writepolicy; delegates core `/tasks*` to taskcore | [handler/README.md](../pkgs/tasks/handler/README.md), [domain/sse-hub.md](./domain/sse-hub.md) |
| Task core BC | `pkgs/taskcore/` | Task CRUD, dependencies, gate, retry, stats, ready queue; `/tasks*` routes | [ADR-0059](./adr/ADR-0059-taskcore-extraction.md), [taskcore/README.md](../pkgs/taskcore/README.md) |
| Domain types | `pkgs/<bc>/domain/` | BC-owned entities (`taskcore`, `taskcycles`, `taskchecklist`, `taskevents`, `projects`, `settings`) — **no** `pkgs/tasks/domain` | [data-model.md](./data-model.md), [ADR-0060](./adr/ADR-0060-retire-tasks-domain.md) |
| Handler store contract | `pkgs/tasks/handler/handler_store.go` | Composes BC `contract` slices for HTTP — **no** `pkgs/tasks/contract` hub | [ADR-0062](./adr/ADR-0062-retire-tasks-contract-hub.md) |
| Harness store contract | `pkgs/agents/harness/internal/contract/` | BC-sliced persistence surface for worker orchestration | [ADR-0061](./adr/ADR-0061-harness-contract-slices.md), [domain/harness.md](./domain/harness.md) |
| Projects bounded context | `pkgs/projects/` | Project CRUD, context graph, `/projects*` HTTP | [ADR-0045](./adr/ADR-0045-bounded-context-projects.md), [projects/README.md](../pkgs/projects/README.md) |
| Git inventory bounded context | `pkgs/gitinventory/` | Registered repos/worktrees/branches, `/git/*` HTTP, reconcile | [ADR-0046](./adr/ADR-0046-bounded-context-gitinventory.md), [gitinventory/README.md](../pkgs/gitinventory/README.md) |
| Settings bounded context | `pkgs/settings/` | App settings row, `/settings*` HTTP, workspace browse, agent probe/cancel | [ADR-0047](./adr/ADR-0047-bounded-context-settings.md), [settings/README.md](../pkgs/settings/README.md) |
| Task compose bounded context | `pkgs/taskcompose/` | Task drafts + templates, `/task-drafts*` and `/task-templates*` HTTP | [ADR-0048](./adr/ADR-0048-bounded-context-taskcompose.md), [taskcompose/README.md](../pkgs/taskcompose/README.md) |
| Workspace repo HTTP | `pkgs/repo/handler/` | `/repo/*` search, file preview, validate-range, diff | [ADR-0049](./adr/ADR-0049-repo-http-handler.md), [repo/handler/README.md](../pkgs/repo/handler/README.md) |
| Store kernel (shared) | `pkgs/storekernel/` | DeferLatency, MapNotFound, event seq append, validators | [ADR-0050](./adr/ADR-0050-storekernel-extraction.md), [storekernel/README.md](../pkgs/storekernel/README.md) |
| Task checklist BC | `pkgs/taskchecklist/` | Done criteria CRUD, `/tasks/{id}/checklist*`; local `domain/` + `store/model/` | [ADR-0051](./adr/ADR-0051-bounded-context-taskchecklist.md), [ADR-0056](./adr/ADR-0056-taskchecklist-domain-model.md), [taskchecklist/README.md](../pkgs/taskchecklist/README.md) |
| Runners HTTP | `pkgs/runners/handler/` | `/runners*` probe, list-models, config-schema | [ADR-0052](./adr/ADR-0052-runners-http-handler.md), [runners/handler/README.md](../pkgs/runners/handler/README.md) |
| Execution cycles BC | `pkgs/taskcycles/` | Cycles, phases, stream, verdicts, commits; local `domain/` + `store/model/` | [ADR-0053](./adr/ADR-0053-bounded-context-taskcycles.md), [ADR-0058](./adr/ADR-0058-taskcycles-domain-model.md), [taskcycles/README.md](../pkgs/taskcycles/README.md) |
| Task audit events BC | `pkgs/taskevents/` | Task event thread, `/tasks/{id}/events*`; local `domain/` + `store/model/` | [ADR-0054](./adr/ADR-0054-bounded-context-taskevents.md), [ADR-0057](./adr/ADR-0057-taskevents-domain-model.md), [taskevents/README.md](../pkgs/taskevents/README.md) |
| Persistence | `pkgs/tasks/store/`, `pkgs/projects/store/`, `pkgs/gitinventory/store/`, `pkgs/settings/store/`, `pkgs/taskcompose/store/`, `pkgs/taskchecklist/store/`, `pkgs/taskcycles/store/`, `pkgs/taskevents/store/`, `pkgs/tasks/postgres/` | Store facade, GORM migrate, dual-write to `task_events` | [domain/persistence.md](./domain/persistence.md), [store/README.md](../pkgs/tasks/store/README.md) |
| Task scheduling | `pkgs/tasks/scheduling/` | Worker readiness predicates, pickup gate, post-commit notify | [domain/task-scheduling.md](./domain/task-scheduling.md), ADR-0023 |
| Execution cycles HTTP | `pkgs/taskcycles/handler/` (was `pkgs/tasks/handler/handler_cycles.go`) | Cycle and phase REST surface | [api.md](./api.md), [data-model.md](./data-model.md) |
| Operator retry | `pkgs/taskcore/handler/`, `pkgs/taskcore/domain/retry.go`, `harness/retry_run.go` | Start over / resume after failure | [retry-start-over.md](./domain/retry-start-over.md), [retry-resume.md](./domain/retry-resume.md) |
| Read/write policy | `pkgs/tasks/handler/readpolicy/`, `writepolicy/` | Bootstrap limits, commit-then-notify SSE enrichment | ADR-0026 |
| Workspace search | `pkgs/repo/` | Path resolution, `@`-mentions; HTTP in `pkgs/repo/handler/` | [domain/workspace-repo.md](./domain/workspace-repo.md), [ADR-0049](./adr/ADR-0049-repo-http-handler.md) |
| Agent queue | `pkgs/agents/` (notifier hook) | Ready-task enqueue, reconcile tick, queue cap | [domain/agent-queue.md](./domain/agent-queue.md) |
| Agent harness | `pkgs/agents/harness/` | Execute/verify loop, criteria, git integrity, retry modes | [domain/harness.md](./domain/harness.md), [cursor-session-resume.md](./domain/cursor-session-resume.md) |
| Cycle commits | `harness/internal/git/commits.go`, `store/internal/commits/` | Agent-claimed commit ledger for verify | [cycle-commits.md](./domain/cycle-commits.md), ADR-0032 |
| Agent worker | `pkgs/agents/worker/` | Single-goroutine queue consumer; calls harness | [configuration.md](./configuration.md) |
| Worker supervisor | `internal/taskapi/agentworker/` | Boot, reload, probe, hot-swap worker from settings | [domain/agent-supervisor.md](./domain/agent-supervisor.md) |
| Agent runner stack | `pkgs/agents/runner/` | `Runner` interface, cursor adapter, registry, adapterkit, runnerfake | [domain/runner-adapters.md](./domain/runner-adapters.md) |
| Request logging | `pkgs/tasks/logctx/`, `pkgs/tasks/calltrace/` | `request_id`, `log_seq`, `call_path` in logs | [observability-trace-lines.md](./domain/observability-trace-lines.md) |
| JSON response helpers | `pkgs/tasks/apijson/` | Security headers, `WriteJSONError` | — |
| HTTP middleware | `pkgs/tasks/middleware/`, `internal/taskapi/` | Middleware stack, server assembly | [middleware/README.md](../pkgs/tasks/middleware/README.md) |
| Dev SSE simulation | `pkgs/tasks/devsim/` | `HAMIX_SSE_TEST` synthetic events for local UI | [api.md](./api.md), [configuration.md](./configuration.md) |

## Web

| Area | Path | Purpose | Deep dive |
| --- | --- | --- | --- |
| API client | `web/src/api/` | All `fetch` calls; `parseTaskApi*` parsers | [web.md](./web.md) |
| Task sync (SSE) | `web/src/tasks/sync/` | Cache invalidation, debounce, enrichment on SSE | [web.md](./web.md) §Task sync, ADR-0022 |
| Query keys / hooks | `web/src/tasks/task-query/` | TanStack Query keys, list/detail hooks, SSE bridge | [web.md](./web.md) |
| Query policy | `web/src/tasks/queryPolicy.ts`, `tasks/mutations/`, `tasks/checklist/`, `tasks/app/` | StaleTime tiers, guarded mutations, TasksAppProvider | ADR-0025 |
| Task create flow | `web/src/tasks/create/`, `components/task-create-modal/` | Compose payload, drafts, modal policy | [web.md](./web.md) §Task create, ADR-0024 |
| Task pages | `web/src/tasks/pages/` | Route-level containers (home, detail, templates, commits) | [web.md](./web.md) §Routes |
| Task list UI | `web/src/tasks/components/task-list/` | Home table, filters, bulk actions, stats | [web.md](./web.md) |
| Task detail UI | `web/src/tasks/components/task-detail/` | Header, schedule, cycles, commits, checklist panels | [web.md](./web.md) §Task detail |
| Task display helpers | `web/src/tasks/task-display/` | Shared labels, edit guards, priority display | — |
| Task templates | `web/src/api/taskTemplates.ts`, `pages/TaskTemplatesPage.tsx` | Template CRUD and batch instantiate | [api.md](./api.md), [web.md](./web.md) |
| App shell | `web/src/app/` | Router, providers, bootstrap, global styles | [web.md](./web.md) |
| Projects / settings | `web/src/projects/`, `web/src/settings/` | Non-task feature modules | [web.md](./web.md) §Routes |

## Infra and test helpers

| Area | Path | Purpose | Deep dive |
| --- | --- | --- | --- |
| Binaries | `cmd/taskapi/`, `cmd/dbcheck/` | Entrypoints only; wire deps and run | [cmd/taskapi/README.md](../cmd/taskapi/README.md) |
| Env loading | `internal/envload/` | Resolve `.env` from repo root | [configuration.md](./configuration.md) |
| taskapi config | `internal/taskapiconfig/` | Listen host, log level, queue cap, dev SSE interval | [configuration.md](./configuration.md) |
| SQLite test DB | `internal/tasktestdb/` | In-memory GORM for default Go tests | [CONTRIBUTING.md](../CONTRIBUTING.md#before-you-open-a-pr) |
| Middleware tests | `internal/middlewaretest/` | Black-box tests for middleware stack | [middleware/README.md](../pkgs/tasks/middleware/README.md) |
| Handler tests | `internal/handlertest/` | Black-box HTTP tests for exported handler API | — |
| Security header tests | `internal/httpsecurityexpect/` | Shared baseline header assertions | — |
| Agent reconcile tests | `pkgs/tasks/agentreconcile/` | SQLite integration tests; not production code | — |

## See also

- [guide.md](./guide.md) — documentation layers and learning paths
- [README.md](./README.md) — doc index by topic
- [AGENTS.md](../AGENTS.md) — scoped paths when editing
- [AGENTS.md](../AGENTS.md) §Where to find X — intent-based lookup (route, harness, sync, …)
