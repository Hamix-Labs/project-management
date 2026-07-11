# `pkgs/projects/handler`

REST handlers for `/projects*` routes. Registered from [`pkgs/tasks/handler/handler_routes.go`](../../tasks/handler/handler_routes.go) via `Register(mux, Deps)`.

| File | Role |
| --- | --- |
| `handler.go` | `Register`, `Deps`, SSE `NotifyFunc` |
| `handler_projects.go` | Route handlers |
| `handler_json.go` | Request/response wire types |
| `handler_params.go` | Query/path parsing |
| `http_helpers.go` | Local JSON/error helpers (no import of `pkgs/tasks/handler`) |

Store dependency is `contract.ProjectStore` — satisfied by `pkgs/tasks/store.Store` at wire time.

Contract tests remain under `pkgs/tasks/handler/handler_http_projects_*`.
