# ADR-0060: Retire pkgs/tasks/domain compat shim

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Tier 3 extracted bounded contexts with real `domain/` packages, but `pkgs/tasks/domain` remained a 16-file alias tree (~250 import sites). New contributors grepped `Task` and landed in shims instead of `taskcore/domain`.

## Decision

1. **Delete** `pkgs/tasks/domain/` entirely after migrating all Go imports to BC owners (`taskcore`, `taskcycles`, `taskchecklist`, `taskevents`, `projects`, `settings`).
2. **Actor canonical owner** — `pkgs/taskcore/domain.Actor`; BC contracts use `taskcoredomain.Actor`, never the compat path.
3. **CI gate** — `scripts/check-go.sh` → `step_tasks_domain_retired` fails if `pkgs/tasks/domain` exists or any `*.go` imports it.

## Consequences

### Positive

- Grep and agent-map routes point engineers to real ownership.
- Import lint is enforceable at CI.

### Negative / Trade-offs

- Historical doc links to `pkgs/tasks/domain/*` need periodic cleanup (see agent-map).

## See also

- [Tier 4 umbrella plan](../../.cursor/plans/tier_4_compat_retirement_ddd9ab15.plan.md)
- [ADR-0059](./ADR-0059-taskcore-extraction.md)
