# ADR-0050: Extract `pkgs/storekernel`

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/tasks/kernel` held shared store helpers (DeferLatency, MapNotFound, event append, validators) imported by `pkgs/projects/store`, `pkgs/gitinventory/store`, `pkgs/settings/store`, and `pkgs/taskcompose/store`. ADR-0045–0048 documented this as temporary debt until a fourth bounded context needed a shared kernel.

Tier 2 extractions (taskchecklist, taskcycles, taskevents) require the same helpers without implying ownership by `pkgs/tasks`.

## Decision

1. **New package** — `pkgs/storekernel` owns all former `pkgs/tasks/kernel` implementation files (same API, package name `storekernel`).

2. **Delete shim path** — Remove `pkgs/tasks/kernel`; callers import `pkgs/storekernel` directly.

3. **Import rules** — `pkgs/storekernel` may import `pkgs/tasks/domain` and `pkgs/tasks/store/model` for event append and LoadTask; it must not import `pkgs/tasks/store/internal` or handlers.

4. **CI** — `scripts/check-go.sh` → `step_storekernel_boundary` rejects `pkgs/storekernel` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal`.

## Consequences

### Positive

- BC stores depend on a neutral kernel package, not a tasks subfolder.
- Clear home for shared persistence primitives as more tables extract.

### Negative / trade-offs

- Kernel still depends on `pkgs/tasks/domain` and `pkgs/tasks/store/model` until event/task row types move or narrow interfaces are introduced.

## See also

- [pkgs/storekernel/README.md](../../pkgs/storekernel/README.md)
- Tier 2 umbrella plan (taskchecklist, taskcycles, taskevents)
