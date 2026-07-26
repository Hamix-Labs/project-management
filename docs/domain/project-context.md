# Project context (removed)

| | |
| --- | --- |
| **Status** | Removed |
| **Superseding ADR** | [ADR-0087](../adr/ADR-0087-remove-project-context.md) |
| **Original decision** | [ADR-0001](../adr/ADR-0001-project-context.md) (superseded) |

Project memory — curated context items/edges, task `project_context_item_ids`,
harness `<project_context>` injection, and `task_context_snapshots` — has been
deleted. Projects remain as repo-bound task containers (CRUD + `#N`). Prompt
file attach continues via `@` repository-file mentions only.

For current project HTTP and schema, see [api.md](../api.md) and
[data-model.md](../data-model.md).
