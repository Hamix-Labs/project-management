# ADR-0052: Extract `/runners/*` HTTP to `pkgs/runners/handler`

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Runner registry HTTP routes (list, config schema, probe, list-models, validate-config) lived in `pkgs/tasks/handler/handler_runners.go`. Runner execution and adapter contracts already live under `pkgs/agents/runner/`; only HTTP adapters and settings lookup for default binary paths were mixed into the tasks handler package.

Prior extractions (repo — ADR-0049; settings, gitinventory, taskchecklist — ADR-0045–0048) established the pattern: register sibling routes from `handler_routes.go`, keep contract tests on the full mux in `pkgs/tasks/handler`, and depend on narrow store contracts rather than the tasks handler.

## Decision

1. **HTTP only** — `pkgs/runners/handler` owns the five `/runners/*` routes. No new store layer; registry logic stays in `pkgs/agents/runner/registry`.

2. **Settings dependency** — Handlers need only `contract.SettingsStore.GetSettings` (for default `CursorBin`). Wired as `Deps.Settings`; field name `settings` on `Handler`.

3. **Route registration** — `runnershandler.Register(m, runnershandler.Deps{Settings: h.store})` from `pkgs/tasks/handler/handler_routes.go`. No URL or JSON shape changes.

4. **Contract tests** — `handler_http_runners_contract_test.go` stays in `pkgs/tasks/handler` (full mux dependency; same approach as repo/settings extractions).

5. **Import gate** — CI rejects `pkgs/runners/handler` importing `pkgs/tasks/handler` (`scripts/check-go.sh` → `step_runners_handler_boundary`).

## Consequences

### Positive

- `/runners/*` ownership aligns with runner adapter documentation and future `pkgs/runners/` growth
- Smaller tasks handler surface; settings read is explicit via contract
- No web or `docs/api.md` contract changes

### Negative / trade-offs

- HTTP helpers duplicated lightly vs `pkgs/taskchecklist/handler` until a shared thin helper package exists
- Default binary resolution still reads app settings through the tasks store facade (intentional — single settings row)

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Keep handlers in tasks | Leaves route ownership split from runner domain docs |
| Move contract tests to `pkgs/runners/handler` | Would import tasks handler or duplicate mux wiring |
| New runners store package | No persistence; registry is in-memory |

## See also

- [pkgs/runners/handler/README.md](../../pkgs/runners/handler/README.md)
- [docs/domain/runner-adapters.md](../domain/runner-adapters.md)
- [ADR-0049](./ADR-0049-repo-http-handler.md) — prior extraction pattern
