# ADR-0047: Extract `pkgs/settings` bounded context

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

App settings (singleton `app_settings` row), workspace browse helpers, and agent-worker control endpoints lived inside `pkgs/tasks/handler` and `pkgs/tasks/store/internal/settings/`. The row was already isolated behind `contract.SettingsStore`, but HTTP routes and GORM models still mixed with tasks, projects, and git inventory. ADR-0045 and ADR-0046 proved the bounded-context extraction pattern; settings was the next slice (see ADR-0046 future extractions).

Supervisor, harness, and bootstrap still read settings through the composed tasks store facade — no import cycle if settings never imports task CRUD or `pkgs/tasks/handler`.

## Decision

1. **New bounded context** — `pkgs/settings/{domain,store,handler}` owns `AppSettings` types, GORM persistence, and `/settings*` HTTP routes (including workspace-roots, browse-dirs, git-probe).

2. **Composition root unchanged** — `cmd/taskapi` still wires `pkgs/tasks/store.Store`; that facade embeds/delegates `contract.SettingsStore` to `pkgs/settings/store.Store`.

3. **Route registration** — `pkgs/settings/handler.Register` mounts the same URLs from `pkgs/tasks/handler/handler_routes.go`. No path or JSON shape changes.

4. **Contract kernel** — `pkgs/tasks/contract` keeps `SettingsStore` and `SettingsPatch`. Handlers depend on contract interfaces, not the tasks store facade.

5. **Agent worker control** — `pkgs/settings/contract` owns the narrow `AgentWorkerControl` interface (cancel, reload, probe). Handlers and `cmd/taskapi` use that contract; settings handler does not import `pkgs/tasks/handler`.

6. **Cross-context reads** — Workspace roots listing uses `contract.GitReadStore` injected at registration (registered repositories from git inventory).

7. **Domain purity** — `pkgs/settings/domain` imports stdlib only. Settings sentinel errors use the `settings:` prefix; handlers strip detail for client-facing 400 messages.

8. **Import gate** — CI rejects `pkgs/settings` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (`scripts/check-go.sh` → `step_settings_boundary`).

9. **Postgres migrate** — `pkgs/tasks/postgres` and `pkgs/tasks/store/model` AutoMigrate import GORM models from `pkgs/settings/store/model`; one database, one binary.

## Consequences

### Positive

- Settings-focused tests, logs, and file ownership narrow to `pkgs/settings`
- Completes the projects → gitinventory → settings structural split ahead of Phase 2 dead-code work
- Web and operator docs unchanged at the HTTP contract layer

### Negative / trade-offs

- Supervisor and harness still reach settings through the composed tasks store facade (not a separate binary)
- `pkgs/settings/store` may import `pkgs/tasks/kernel` and `pkgs/tasks/store/model` for shared migrate parity until `pkgs/storekernel/` lands (same debt as projects and gitinventory)
- SSE cross-route tests in `pkgs/tasks/handler` keep thin settings test helpers that wire the full taskapi mux

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Type aliases in `pkgs/tasks/domain` indefinitely | Hides ownership; blocks import lint |
| Split `taskapi` binary | Out of scope; composition root stays in tasks store |
| Move contract into settings | Breaks supervisor/harness imports that already use `pkgs/tasks/contract` |
| Keep handlers in tasks, store-only extraction | Leaves mixed route ownership and duplicate HTTP helpers |

## Future extractions (not this ADR)

| Context | Trigger |
| --- | --- |
| `pkgs/storekernel/` | Third store extraction needs shared kernel |
| Auth / multi-tenant | May split settings further or add middleware-owned config |

## See also

- [pkgs/settings/README.md](../../pkgs/settings/README.md)
- [docs/agent-map.md](../agent-map.md)
- [ADR-0046](./ADR-0046-bounded-context-gitinventory.md) — prior extraction
- [ADR-0045](./ADR-0045-bounded-context-projects.md) — extraction template
