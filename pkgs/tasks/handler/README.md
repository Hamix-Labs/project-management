# `pkgs/tasks/handler`

HTTP surface for `taskapi`: REST + optional `/repo` + `GET /events` (SSE). **Contracts:** [docs/api.md](../../../docs/api.md). **How to extend:** [CONTRIBUTING.md](../../../CONTRIBUTING.md), [docs/agent-map.md](../../../docs/agent-map.md).

`NewHandler` returns the **inner mux** (routes only). `cmd/taskapi` mounts it behind **`middleware.Stack(..., calltrace.Path)`** via **`internal/taskapi.NewHTTPHandler`**. Hub construction uses **`pkgs/tasks/realtime`** (`*realtime.SSEHub`); this package owns stream framing and notify only.

## Area → files (disk)

| Prefix / area | Files |
| --- | --- |
| Mux composition | `handler.go`, `handler_routes.go`, `handler_*_wire.go`, `handler_projects_wire.go` |
| SSE (HTTP + notify) | `sse_stream.go`, `sse_notify.go` (+ whitebox `sse_*_test.go`, `sse_whitebox_helpers_test.go`) |
| Health | `handler_health.go`, `handler_system_health.go` |
| SPA / operator shell | `handler_bootstrap.go`, `handler_rum.go` |
| Policy | `handler_writepolicy.go`, `policy/` |
| Cross-BC glue | `handler_task_*`, `handler_compose_*`, `handler_git_helpers.go` |
| Shell utilities | `repo_compat.go`, `server_version.go`, `handler_http_json.go`, `handler_store.go` (http.io via `handlerhttp`) |
| Fakes | `storefake/` |

Root stays one Go package. Domain grouping is by **filename prefix**, not subdirectories with `package handler`.

## Route registration (`handler_routes.go`)

Sibling BCs register via `*.Register(m, Deps)`. Task-core routes register via [`pkgs/taskcore/handler`](../../taskcore/handler/).

| Package | Routes |
| --- | --- |
| [`pkgs/taskcore/handler`](../../taskcore/handler/) | `/tasks*` (CRUD, stats, gate, dependencies, retry) |
| [`pkgs/projects/handler`](../../projects/handler/) | `/projects*` |
| [`pkgs/gitinventory/handler`](../../gitinventory/handler/) | `/git/*` |
| [`pkgs/settings/handler`](../../settings/handler/) | `/settings*` |
| [`pkgs/taskcompose/handler`](../../taskcompose/handler/) | `/task-drafts*`, `/task-templates*` |
| [`pkgs/taskchecklist/handler`](../../taskchecklist/handler/) | `/tasks/{id}/checklist*` |
| [`pkgs/taskcycles/handler`](../../taskcycles/handler/) | `/tasks/{id}/cycles*`, commits, `/tasks/cycle-failures` |
| [`pkgs/taskevents/handler`](../../taskevents/handler/) | `/tasks/{id}/events*` |
| [`pkgs/repo/handler`](../../repo/handler/) | `/repo/*` |
| [`pkgs/runners/handler`](../../runners/handler/) | `/runners*` |

## Middleware

Implementations live in [`pkgs/tasks/middleware`](../middleware/). Production stack: **`middleware.Stack(inner, calltrace.Path)`**. File map and env table: [`../middleware/README.md`](../middleware/README.md).

## Tests

| Where | What |
| --- | --- |
| [`internal/handlertest`](../../../internal/handlertest/) | Blackbox HTTP (`NewServer`, `NewCreateServer`, `NewSSETriggerServer`, …) |
| [`internal/handlertest/sse`](../../../internal/handlertest/sse/) | Blackbox SSE HTTP (lossless, trigger surface, headers/events) |
| [`internal/handlertest/shell`](../../../internal/handlertest/shell/) | Shell HTTP (RUM, list/error logging) |
| [`internal/handlertest/taskcore`](../../../internal/handlertest/taskcore/) (+ checklist/compose/cycles/events) | BC contract suites |
| `pkgs/tasks/handler/*_test.go` | Whitebox only: fold RUM, WriteJSON helpers, SSE notify/hub units, security headers, writepolicy |

Whitebox suites **cannot** import `handlertest` (import cycle with `handler`). They use package-local `sse_whitebox_helpers_test.go` for drain/assert.

## Navigability metrics (ROI cleanup)

Target: root **&lt;25** files and test ratio **&lt;30%**. After ROI drains/aliases, remaining root files are mostly production shell + unavoidable whitebox tests (notify/writepolicy/security/WriteJSON/RUM fold/hub units). Blackbox HTTP is in `internal/handlertest/**`.

## Do not extract shell glue as a new BC

Bootstrap, health/RUM, and cross-BC git/compose normalize stay here (or thin `pkgs/tasks/service`). See [`.cursor/rules/go-layout.mdc`](../../../.cursor/rules/go-layout.mdc).

When adding a route or production file, update this README in the same PR. Prefer **`internal/handlertest`** theme packages for new blackbox HTTP tests.
