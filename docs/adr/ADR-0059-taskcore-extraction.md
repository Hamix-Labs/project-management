# ADR-0059: Taskcore bounded context extraction

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Tier 3 Track C required extracting task CRUD, dependencies, gate, retry, ready queue, stats, dev-mirror, and health persistence from the monolithic `pkgs/tasks` package. Sibling BCs (`taskchecklist`, `taskevents`, `taskcycles`) already owned their domain/model layers; task row types and `/tasks*` core routes remained in `pkgs/tasks`.

Harness and the SPA still import `pkgs/tasks/domain` aliases during the migration window.

## Decision

1. **New package** — `pkgs/taskcore/{domain,contract,store,handler}` owns task CRUD, dependencies, gate, retry, ready queue, stats, dev-mirror row updates, and health probes.

2. **Actor ownership** — `Actor` / `ActorUser` / `ActorAgent` live in `pkgs/taskcore/domain`. `pkgs/taskevents/domain` aliases them; `pkgs/tasks/domain` re-exports for harness compatibility.

3. **Composition shell** — `pkgs/tasks/store` embeds `*taskcore/store.Store` and retains notify/pickup-wake wiring. `pkgs/tasks/handler` registers `taskcore/handler.Register` and keeps SSE, bootstrap, writepolicy, readpolicy.

4. **Contract hub** — `pkgs/tasks/contract` aliases `TaskCRUDStore`, stats, and health types from `pkgs/taskcore/contract`. Narrow read-only lookup is `taskcore/contract.TaskGetter` (`Get(ctx, id) (*Task, error)`); sibling BCs and agentworker alias or import it rather than redefining identical interfaces.

5. **Migrate hub** — `pkgs/tasks/store/model/migrate_models.go` registers `taskcore/store/model` task tables in FK-safe order.

6. **Import gates** — `scripts/check-go.sh` → `step_taskcore_boundary` forbids `pkgs/tasks/handler`, `pkgs/tasks/store/internal`, GORM, and `pkgs/tasks/domain` inside `pkgs/taskcore/domain`.

## Consequences

### Positive

- Task core matches the four-layer BC layout used by checklist, cycles, and events.
- `pkgs/tasks/handler` file count drops; route table documents `taskcorehandler.Register`.

### Negative / Trade-offs

- Temporary duplicate type names at `pkgs/tasks/domain` alias layer until harness imports `taskcore/domain` directly.
- Bootstrap and compose paths use exported helpers from `taskcore/handler` via thin shims in `pkgs/tasks/handler`.

## See also

- [pkgs/taskcore/README.md](../../pkgs/taskcore/README.md)
- [Tier 3 blueprint plan](../../.cursor/plans/tier_3_bc_blueprint_5ff0fc21.plan.md)
