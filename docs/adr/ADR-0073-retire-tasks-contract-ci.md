# ADR-0073: Enforce pkgs/tasks/contract retirement in CI

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Tier 4 deleted `pkgs/tasks/contract` per [ADR-0062](./ADR-0062-retire-tasks-contract-hub.md). `step_tasks_domain_retired` already guards the parallel `pkgs/tasks/domain` hub, but no CI step mirrored that enforcement for `tasks/contract` — a revived alias tree could land silently.

## Decision

1. **CI gate** — `scripts/check-go.sh` → `step_tasks_contract_retired` fails if `pkgs/tasks/contract` exists or any `*.go` imports `github.com/.../pkgs/tasks/contract`.
2. **No revival** — handler composition uses BC `contract` packages and `pkgs/tasks/wire` at the composition root; do not reintroduce a tasks-level contract hub.

## Consequences

### Positive

- Post–Tier 5 handoff work cannot accidentally restore the second alias hub.
- Symmetric enforcement with `step_tasks_domain_retired`.

### Negative / Trade-offs

- Doc links to `pkgs/tasks/contract` must stay cleaned (see handoff Child 5).

## See also

- [ADR-0062](./ADR-0062-retire-tasks-contract-hub.md)
- [Post–Tier 5 handoff train](../../.cursor/plans/post-tier5_handoff_train_d063a5c6.plan.md)
