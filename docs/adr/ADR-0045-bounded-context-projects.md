# ADR-0045: Extract `pkgs/projects` bounded context

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Project CRUD, curated project context (items + edges), and default-project seeding lived inside `pkgs/tasks` alongside tasks, cycles, git inventory, and settings. The domain was already isolated behind `contract.ProjectStore` and `store/internal/projects/`, but package boundaries still mixed concerns: `grep pkgs/tasks` surfaced project persistence in task logs/tests, and future extractions (git inventory, settings) needed a proven pattern.

Tasks reference projects only via `Task.ProjectID *string` — no import cycle if projects never imports task CRUD.

## Decision

1. **New bounded context** — `pkgs/projects/{domain,store,handler}` owns project types, GORM persistence, and `/projects*` HTTP routes.

2. **Composition root unchanged** — `cmd/taskapi` still wires `pkgs/tasks/store.Store`; that facade embeds/delegates `contract.ProjectStore` to `pkgs/projects/store.Store`.

3. **Route registration** — `pkgs/projects/handler.Register` mounts the same URLs from `pkgs/tasks/handler/handler_routes.go`. No path or JSON shape changes.

4. **Shared kernel (temporary debt)** — `pkgs/projects/store` may import `pkgs/tasks/store/internal/kernel` (DeferLatency, MapNotFound, ResolveID) until a later `pkgs/storekernel/` extraction when a third bounded context needs it.

5. **Domain purity** — `pkgs/projects/domain` imports stdlib only. Project sentinel errors live in `pkgs/projects/domain/errors.go`.

6. **Import gate** — CI rejects `pkgs/projects` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (kernel excepted via allowlist in the check script path — kernel is under `internal/`; the gate documents accepted debt: projects must not import non-kernel tasks store internals).

7. **Postgres migrate** — `pkgs/tasks/postgres` and `pkgs/tasks/store/model` AutoMigrate import GORM models from `pkgs/projects/store/model`; one database, one binary.

## Consequences

### Positive

- First bounded-context extraction template for git inventory and settings
- Project-focused tests and logs narrow to `pkgs/projects`
- `contract.ProjectStore` return types point at `pkgs/projects/domain` — clear ownership for harness and web parsers (JSON unchanged)

### Negative / trade-offs

- Cross-model GORM associations between tasks and projects tables drop pointer fields to avoid `tasks/store/model` ↔ `projects/store/model` import cycles; FK columns and `ModelMigrateExtra` ordering preserve schema
- Harness and worker still reach project methods through the composed tasks store facade (not a separate binary)
- Kernel sharing keeps a dependency from projects store into tasks internal packages until storekernel lands

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Type aliases in `pkgs/tasks/domain` indefinitely | Hides ownership; blocks import lint |
| Split `taskapi` binary | Out of scope; composition root stays in tasks store |
| Move kernel first | Blocks proving bounded-context pattern; kernel debt is documented |

## Future extractions (not this ADR)

| Context | Trigger |
| --- | --- |
| `pkgs/gitinventory/` | Git handler/store complexity + ADR-0037 stable |
| `pkgs/settings/` | Auth / multi-tenant prep |
| `pkgs/storekernel/` | Third store extraction needs shared kernel |

## See also

- [pkgs/projects/README.md](../../pkgs/projects/README.md)
- [docs/agent-map.md](../agent-map.md)
- [ADR-0042](./ADR-0042-repo-default-projects.md) — superseded by [ADR-0094](./ADR-0094-global-default-project.md) (global Default)
