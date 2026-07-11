# `pkgs/runners/handler`

HTTP surface for the runner registry used by the SPA (`/runners/*`). **Contracts:** [docs/api.md](../../docs/api.md) (Runners section), [docs/domain/runner-adapters.md](../../docs/domain/runner-adapters.md).

## Routes

| Method | Path | Handler |
| --- | --- | --- |
| GET | `/runners` | List registered runners and optional config schemas |
| GET | `/runners/{id}/config-schema` | Runner-specific config schema |
| POST | `/runners/{id}/probe` | Probe binary path / version |
| POST | `/runners/{id}/list-models` | List models for a runner binary |
| POST | `/runners/{id}/validate-config` | Validate runner config JSON |

Probe and list-models fall back to `AppSettings.CursorBin` when the request omits `binary_path` and the runner is `cursor`.

## Registration

`runnershandler.Register(m, runnershandler.Deps{Settings: …})` is called from `pkgs/tasks/handler/handler_routes.go`. HTTP contract tests remain in `pkgs/tasks/handler/handler_http_runners_contract_test.go` (full taskapi mux).

## Files

| File | Role |
| --- | --- |
| `handler.go` | `Register`, `Deps`, `Handler` |
| `handler_runners.go` | Route handlers and wire types |
| `http_helpers.go` | JSON responses, store error mapping, debug request logging |
