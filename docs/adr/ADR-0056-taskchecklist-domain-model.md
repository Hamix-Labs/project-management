# ADR-0056: Task checklist domain and store model ownership

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

[ADR-0051](./ADR-0051-bounded-context-taskchecklist.md) extracted checklist HTTP and store facades but left checklist **domain types** in `pkgs/tasks/domain` and **GORM models** in `pkgs/tasks/store/model`. That split contradicted the four-layer BC blueprint established by `pkgs/taskcompose/` and `pkgs/settings/` (local `domain/` + `store/model/`).

Harness and tasks facade still import checklist shapes through `pkgs/tasks/domain` during the Tier 3 migration window.

## Decision

1. **Domain ownership** — `TaskChecklistItem`, `TaskChecklistItemCommand`, `TaskChecklistCompletion`, verify-command limits, and `VerifierKind` live in `pkgs/taskchecklist/domain/`. `TaskChecklistCompletion.By` is a string holding `tasks/domain.Actor` wire values (no import cycle).

2. **Model ownership** — GORM structs and domain↔model mappers for checklist tables live in `pkgs/taskchecklist/store/model/`.

3. **Compatibility aliases** — `pkgs/tasks/domain` re-exports checklist types and verifier symbols as type/const aliases so harness and existing imports compile unchanged.

4. **Migrate hub** — `pkgs/tasks/store/model/migrate_models.go` registers `checklistmodel.*` types in FK-safe order alongside other BC models.

5. **Parity registry** — Checklist `ParityPairs` entries in `pkgs/tasks/store/model/parity.go` reference `checklistdomain` + `checklistmodel` (schema guards stay centralized until a dedicated checklist parity package is warranted).

6. **Import gates** — `scripts/check-go.sh` → `step_taskchecklist_boundary` forbids GORM and `pkgs/tasks/domain` imports in `taskchecklist/domain`.

7. **Unchanged** — Checklist HTTP routes/JSON, `notifyUnblockedDependents` in tasks facade, cross-TX helpers exported from `taskchecklist/store`. Store/handler may import `pkgs/tasks/domain` for `Actor`, task rows, and shared errors.

## Consequences

### Positive

- Checklist BC matches taskcompose/settings navigation and dependency rules.
- Domain layer is stdlib + `Actor` only; persistence tags stay in `store/model`.

### Negative / trade-offs

- Temporary duplicate type names at `tasks/domain` alias layer until taskcore (#186).
- Cycle report models in `tasks/store/model` still FK-reference `checklistmodel.TaskChecklistItem`.

## See also

- [pkgs/taskchecklist/README.md](../../pkgs/taskchecklist/README.md)
- [ADR-0051](./ADR-0051-bounded-context-taskchecklist.md)
- [ADR-0039](./ADR-0039-domain-persistence-separation.md)
