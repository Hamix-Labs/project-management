# ADR-0078: Retire pkgs/tasks/store facade type aliases

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Tier 5 split task persistence into bounded contexts (`taskcore`, `taskcycles`, `taskchecklist`, etc.). The composition facade at `pkgs/tasks/store/facade_*.go` re-exported BC input/DTO types as `type X = …` aliases so callers could write `store.CreateTaskInput` instead of importing each BC package.

That second naming layer duplicated the retired `pkgs/tasks/contract` hub pattern: importers depended on facade type names rather than BC contracts, and every new BC field required updating alias blocks plus grep-based audits.

## Decision

1. **Remove all `type X = …` alias blocks** from `pkgs/tasks/store/facade_*.go`.
2. **Keep `func (s *Store) …` delegations** until [facade deletion](./post_tier5_facade_deletion.plan.md).
3. **Migrate importers** to BC `contract`, `store`, or `domain` types directly (e.g. `taskcorestore.CreateTaskInput`, `checklistcontract.ChecklistVerifyItem`, `gitinventorystore.ReconcileGitInput`).
4. **Retain** `ThreadEntriesForDisplay` and `FindWorktreeInInventory` var/func re-exports until facade deletion (devsim and git startup depend on stable `store.*` entry points).

Facade method signatures use fully qualified BC types internally; callers that only invoke `*store.Store` methods need not import BC packages unless they construct input structs.

## Consequences

### Positive

- One canonical type per DTO — no parallel `store.Foo` / `taskcorestore.Foo` naming.
- Importers align with BC boundary checks in `scripts/check-go.sh`.
- Facade files shrink to pure delegation; alias retirement is a prerequisite for deleting `facade_*.go`.

### Negative / Trade-offs

- Test and harness code gains BC imports (more import lines at call sites that build inputs).
- Grep gate `store.(CreateTaskInput|…)` no longer catches alias drift; CI relies on absent alias blocks plus BC boundary steps.

## Alternatives

- **Keep aliases until facade deletion** — rejected; prolongs dual naming and blocks handler/service migration.
- **Reintroduce `pkgs/tasks/contract` hub** — rejected per [ADR-0062](./ADR-0062-retire-tasks-contract-hub.md).

## See also

- [Post–Tier 5 handoff train](../../.cursor/plans/post_tier5_handoff_train_d063a5c6.plan.md) — Child 3
- [ADR-0062](./ADR-0062-retire-tasks-contract-hub.md) — contract hub retirement
