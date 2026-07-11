# `pkgs/tasks/handler`

HTTP surface for `taskapi`: REST + optional `/repo` + `GET /events` (SSE). **Contracts:** [docs/api.md](../../docs/api.md). **How to extend:** [CONTRIBUTING.md](../../CONTRIBUTING.md), [AGENTS.md](../../AGENTS.md).

The returned `http.Handler` from `NewHandler` is the **inner mux** (routes only). `cmd/taskapi` mounts it behind **`middleware.Stack(..., calltrace.Path)`** from **`internal/taskapi.NewHTTPHandler`**. Wiring order and devsim live in **`cmd/taskapi/run.go`**. Taskapi-only env parsing lives in **`internal/taskapiconfig`**.

## Middleware (`With*` — outer stack from `middleware.Stack`)

Implementations live in **[`pkgs/tasks/middleware`](../middleware/)** (no import of `handler`). **`middleware_shim.go`** re-exports the same names for `cmd/taskapi` and tests that still import `handler`. **File map, `Stack` order, and env table:** [`../middleware/README.md`](../middleware/README.md).

| Middleware | Role |
|------------|------|
| `WithRecovery` | Panic → 500 JSON. |
| `WithHTTPMetrics` | Prometheus `taskapi_http_*` + in-flight gauge (health paths excluded from latency). |
| `WithAccessLog` | `http.access` line, `request_id`, `log_seq` scope. |
| `WithRateLimit` | Per-IP token bucket (`HAMIX_RATE_LIMIT_PER_MIN`). |
| `WithAPIAuth` | Optional bearer token (`HAMIX_API_TOKEN`). |
| `WithRequestTimeout` | Context deadline on API routes; SSE exempt. |
| `WithMaxRequestBody` | Body size cap (`HAMIX_MAX_REQUEST_BODY_BYTES`). |
| `WithIdempotency` | `Idempotency-Key` replay cache. |

**`middleware.Stack(inner, callPath)`** in `pkgs/tasks/middleware/stack.go` composes the `With*` layers; production passes **`calltrace.Path`** so access logs include `call_path`.

`GET /metrics` is registered on the **outer** mux in `cmd/taskapi` (not on the inner handler mux).

## Core mux and types

| File | Role |
|------|------|
| `handler.go` | `Handler`, `NewHandler`, route registration, JSON security header helpers. |
| `sse_types.go` | Wire type aliases to `pkgs/tasks/realtime`. |
| `sse_hub.go` | `SSEHub`, publish fanout, ring buffer, eviction. |
| `sse_stream.go` | `streamEvents` (`GET /events`). |
| `sse_notify.go` | `notifyChange` / enriched publish helpers; `notifyScopelessChange` for id-less hints (`settings_changed`, `agent_run_cancelled`). Domain doc: [docs/domain/sse-hub.md](../../docs/domain/sse-hub.md). |

## Route registration (`handler_routes.go`)

Sibling bounded contexts register via `*.Register(m, Deps)` from this package's `registerRoutes`. **Task-core routes** register via [`pkgs/taskcore/handler`](../../taskcore/handler/) `Register`.

| Package | Routes | Registered from |
| --- | --- | --- |
| [`pkgs/taskcore/handler`](../../taskcore/handler/) | `/tasks*` (CRUD, stats, gate, dependencies, retry) | `taskcorehandler.Register` |
| [`pkgs/projects/handler`](../../projects/handler/) | `/projects*` | `projecthandler.Register` |
| [`pkgs/gitinventory/handler`](../../gitinventory/handler/) | `/git/*` | `gitinventoryhandler.Register` |
| [`pkgs/settings/handler`](../../settings/handler/) | `/settings*` | `settingshandler.Register` |
| [`pkgs/taskcompose/handler`](../../taskcompose/handler/) | `/task-drafts*`, `/task-templates*` | `composehandler.Register` |
| [`pkgs/taskchecklist/handler`](../../taskchecklist/handler/) | `/tasks/{id}/checklist*` | `checklisthandler.Register` |
| [`pkgs/taskcycles/handler`](../../taskcycles/handler/) | `/tasks/{id}/cycles*`, commits, `/tasks/cycle-failures` | `taskcycleshandler.Register` |
| [`pkgs/taskevents/handler`](../../taskevents/handler/) | `/tasks/{id}/events*` | `eventhandler.Register` |
| [`pkgs/repo/handler`](../../repo/handler/) | `/repo/*` | `repohandler.Register` |
| [`pkgs/runners/handler`](../../runners/handler/) | `/runners*` | `runnershandler.Register` |

## Route handlers — tasks shell (inner mux)

| Area | Files |
|------|--------|
| Health + SSE | `handler_health.go`, `handler_system_health.go`; `sse_*.go` (`GET /events`) |
| Taskcore wiring | `handler_taskcore_wire.go`, `handler_taskcore_compat.go`, `handler_taskcore_compose.go` |
| Bootstrap + RUM | `handler_bootstrap.go`, `handler_rum.go` |
| Write policy | `handler_writepolicy.go`, `writepolicy/` |
| Read policy | `readpolicy/` |

Core `/tasks*` handlers live in [`pkgs/taskcore/handler`](../../taskcore/handler/).

## Request/response helpers

| File | Role |
|------|------|
| `handler_http_json.go` | `decodeJSON`, `writeJSON` / `writeError`, `actorFromRequest`, store error → HTTP. |
| `handler_task_json.go` | Request/response DTOs (`taskCreateJSON`, flat task encoding, etc.). |
| `handler_path_ids.go` | Path UUID / segment parsing and abuse-guard caps. |
| `patch_fields.go` | `PATCH` helpers (e.g. nullable `project_id`, `gate`). |
| `server_version.go` | Build/version string for health JSON. |

## Observability and debug logging

| File | Role |
|------|------|
| (sibling package) | **[`pkgs/tasks/calltrace`](../calltrace/)** — `WithRequestRoot`, `Push`, `Path`, `RunObserved`, `HelperIOIn` / `HelperIOOut` for `call_path` and helper.io traces. File map: [`../calltrace/README.md`](../calltrace/README.md). |
| `httplog_io.go` | `http.io` debug summaries (uses `calltrace.Path`). |
| (sibling package) | **[`pkgs/tasks/logctx`](../logctx/)** — `ContextWithLogSeq`, `ContextWithRequestID`, `RequestIDFromContext`, slog wrappers (`WrapSlogHandlerWithLogSequence`, `WrapSlogHandlerWithRequestContext`). Used from middleware, `handler_http_json.go`, and `cmd/taskapi/run.go` (no import cycle). |
| (sibling package) | **[`pkgs/tasks/apijson`](../apijson/)** — `ApplySecurityHeaders`, `WriteJSONError` (JSON `{"error", "request_id"}` + `http.io` debug). `handler` passes `calltrace.Path` into `WriteJSONError`; middleware receives the same `Path` function from `internal/taskapi`. |

## Tests

| Where | What |
|-------|------|
| **[`internal/handlertest`](../../internal/handlertest/)** | Black-box HTTP against exported `NewHandler` / `With*` only (health, metrics scrape, health security headers). Helpers: `handlertest.NewServer`, `NewServerWithStore`, `NewServerWithRepo`. |
| **[`internal/httpsecurityexpect`](../../internal/httpsecurityexpect/)** | Shared `AssertBaselineHeaders` for handler whitebox tests and `handlertest` (avoids import cycles). |
| **`pkgs/tasks/handler/*_test.go`** | Whitebox tests (unexported helpers, `decodeJSON`, path parsing, SSE handler internals). `handler_http_*_test.go` beside the route area; **`handler_http_testserver_test.go`** has `newTaskTestServer*` for tests not yet moved. **`stack_test.go`** asserts production **`middleware.Stack(..., calltrace.Path)`**. Call-stack unit tests live in **`pkgs/tasks/calltrace`**. |

When adding a **new** route or middleware file, extend this README in the same PR. Prefer **`internal/handlertest`** for new black-box HTTP tests.

## Scaling this package

`handler` stays a **single package** (one directory in Go). To avoid an unmaintainable mix of routes and tests over time, follow the **When a file feels too large** section below — what already lives in `middleware`, `calltrace`, and `internal/middlewaretest`, conventions for **whitebox vs black-box** tests, and **ordered next extractions** (e.g. task JSON types).
