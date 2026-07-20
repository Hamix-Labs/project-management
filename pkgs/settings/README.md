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

- **`cmd/taskapi`** constructs `internal/taskapi/composition.API` via `composition.NewAPI(db)` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- Composition holds `*settingsstore.Store` and satisfies `contract.SettingsStore`.
- Supervisor reload/probe/cancel is injected as `settings/contract.AgentWorkerControl` from `handler_routes.go` (same supervisor type as before, without importing `pkgs/tasks/handler` from settings).
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | `pkgs/tasks/*`, GORM |
| `store` | `settings/domain`, `settings/contract`, GORM, `pkgs/storekernel`, `pkgs/tasks/calltrace` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `handler` | `settings/domain`, `settings/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx`, `pkgs/tasks/realtime`, `pkgs/tasks/handlerhttp`, `pkgs/gitwork`, `pkgs/repo` | `internal/taskapi/composition`, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_settings_boundary`.

## Tests

```powershell
go test ./pkgs/settings/... -count=1
```

Contract coverage for `/settings*` lives in [`handler/handler_http_settings_contract_test.go`](./handler/handler_http_settings_contract_test.go).

## See also

- [docs/api.md](../../docs/api.md) — `/settings*` contract
- [docs/configuration.md](../../docs/configuration.md) — env vars and `app_settings` columns
- [pkgs/settings/contract/settings.go](./contract/settings.go) — `SettingsStore` interface
- [pkgs/settings/contract/agent_worker_control.go](./contract/agent_worker_control.go) — `AgentWorkerControl`
