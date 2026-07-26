# ADR-0087: Remove project context (memory)

**Date:** 2026-07-26
**Status:** Accepted
**Deciders:** Backend / web maintainers
**Supersedes:** [ADR-0001](./ADR-0001-project-context.md)

## Context

[ADR-0001](./ADR-0001-project-context.md) introduced project-scoped relational
memory: curated context items/edges, task selection (`project_context_item_ids`),
harness `<project_context>` injection, and cycle-scoped `task_context_snapshots`.

That surface added REST/UI/SSE/schema weight without becoming a validated
product loop. Operators still need projects as repo-bound containers for tasks
(`project_id`, per-project `#N`), and `@` repository-file mentions remain useful
in prompts. The memory subsystem itself is unused enough to justify a full
teardown rather than a partial freeze.

## Decision

1. **Delete** project memory end-to-end: SPA entry/`/context` UI, `#` TipTap
   mentions, task `project_context_item_ids`, harness load/wrap/snapshots,
   `/projects/{id}/context*` REST, SSE `project_context_changed`, tables
   `project_context_items` / `project_context_edges` / `task_context_snapshots`,
   and `projects.context_summary`.
2. **Keep** projects CRUD (list/detail/settings without memory) and `@`
   repo-file attach/search.
3. **Schema** — bump `SchemaRevision` and drop leftover columns/tables via an
   idempotent migrate (`migrateRemoveProjectContext`).
4. **No replacement** — do not ship a new memory product in the same change.

## Consequences

### Positive

- Smaller API, SPA, harness, and schema surface.
- Projects BC stays focused on membership and repo defaults.
- Prompt attach remains via `@` file mentions only.

### Negative / trade-offs

- Existing curated memory rows are dropped on upgrade (no export/migration path).
- Historical execute prompts that relied on `<project_context>` cannot be
  reconstructed from snapshots after the tables are gone.
- ADR-0001’s “shared memory across months of tasks” goal is deferred until a
  future product decision.

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Soft-deprecate UI, keep tables | Leaves dead schema, handlers, and harness paths. |
| Replace with embeddings/vector memory | Out of scope; same reason ADR-0001 deferred vectors. |
| Keep snapshots for audit only | Still requires selection/render code and storage. |

## See also

- [ADR-0001](./ADR-0001-project-context.md) — original decision (superseded)
- [ADR-0045](./ADR-0045-bounded-context-projects.md) — projects BC extraction
- [api.md](../api.md) — `/projects*` without context routes
- [data-model.md](../data-model.md) — projects without memory tables
