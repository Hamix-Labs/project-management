# ADR-0064: Web parser and types BC finish

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`web/src/api/parseTaskApi.ts` was a barrel, but `parseTaskApiTasks.ts` (~516 lines) mixed task CRUD, checklist, stats, and cycle-failures parsers. Checklist wire types lived inside `types/taskCore.ts`.

## Decision

1. **Parser modules** — `parseTaskApiTasks.ts` (task list/get), `parseTaskApiStats.ts` (stats + cycle failures), `parseTaskApiChecklist.ts` (checklist); barrel unchanged at `parseTaskApi.ts`.
2. **Types** — `web/src/types/checklist.ts` owns checklist wire types; `taskCore.ts` re-exports for backward compat; `types/index.ts` exports the new module.
3. **Tests** — `parseTaskApiStats.test.ts` holds stats/cycle-failure parser tests; core barrel tests remain in `parseTaskApi.test.ts`.
4. **God-file splits** — workspace picker, rich prompt editor, depends-on picker, project context panel split into hooks/subcomponents (Track G).

## Consequences

### Positive

- Parser ownership mirrors backend BC map in [web-layout.mdc](../../.cursor/rules/web-layout.mdc).
- Smaller files meet CODE_STANDARDS reviewability targets.

### Negative / Trade-offs

- `taskCore.ts` re-exports checklist types until all imports migrate to `@/types/checklist`.

## See also

- [docs/web.md](../web.md)
