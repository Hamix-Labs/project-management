# `pkgs/settings`

Bounded context for the singleton `app_settings` row, workspace browse helpers, and `/settings*` HTTP routes. Extracted from `pkgs/tasks` per [ADR-0047](../../docs/adr/ADR-0047-bounded-context-settings.md).

HTTP routes (`/settings`, `/settings/workspace-roots`, …) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/settings/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `AppSettings`, defaults, validation, sentinel errors — stdlib only |
| Contract | [`contract/`](./contract/) | `SettingsStore`, `SettingsPatch`, `AgentWorkerControl` |
| Store | [`store/`](./store/) | GORM persistence facade; `internal/settings/` holds CRUD; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/settings*` REST handlers and wire DTOs |

## Wiring

- **`cmd/taskapi`** still constructs `pkgs/tasks/store.Store` as the composition root.
- `tasks/store.Store` holds `*settingsstore.Store` and implements `contract.SettingsStore` via delegation ([`facade_settings.go`](../tasks/store/facade_settings.go)).
- Supervisor reload/probe/cancel is injected as `settings/contract.AgentWorkerControl` from `handler_routes.go` (same supervisor type as before, without importing `pkgs/tasks/handler` from settings).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | `pkgs/tasks/*`, GORM |
| `store` | `settings/domain`, `settings/contract`, GORM, `pkgs/storekernel`, `pkgs/tasks/store/model` (migrate parity) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `handler` | `settings/domain`, `settings/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx`, `pkgs/tasks/realtime`, `pkgs/gitwork`, `pkgs/repo` | `pkgs/tasks/store` facade, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_settings_boundary`.

## Tests

```powershell
go test ./pkgs/settings/... -count=1
go test ./pkgs/tasks/handler/... -run Settings -count=1
```

Contract coverage for `/settings*` lives in [`handler/handler_http_settings_contract_test.go`](./handler/handler_http_settings_contract_test.go). Cross-route SSE trigger pins remain in [`pkgs/tasks/handler/sse_trigger_surface_test.go`](../tasks/handler/sse_trigger_surface_test.go).

## See also

- [docs/api.md](../../docs/api.md) — `/settings*` contract
- [docs/configuration.md](../../docs/configuration.md) — env vars and `app_settings` columns
- [pkgs/settings/contract/settings.go](./contract/settings.go) — `SettingsStore` interface
- [pkgs/settings/contract/agent_worker_control.go](./contract/agent_worker_control.go) — `AgentWorkerControl`
