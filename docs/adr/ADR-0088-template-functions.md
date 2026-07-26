# ADR-0088: Template function inputs and soft scope

**Date:** 2026-07-26
**Status:** Accepted
**Deciders:** Backend / web maintainers

## Context

Templates are frozen compose blueprints. Operators need reusable “template
functions” that accept create-time inputs (directory, file, or function
reference) so the resulting agent task is scoped to those targets. There is no
agent filesystem sandbox in v1.

`normalizeComposePayloadRaw` decode→remarshal drops unknown JSON keys, so any
schema must be a first-class typed compose field.

## Decision

1. **Schema** — optional `function_inputs` on the compose/template payload:
   `{ id, kind: dir|file|function, label, required?, multiple? }`.
2. **Persistence** — stored inside `task_templates.payload_json` (no new column).
3. **Task create** — non-empty `function_inputs` on `POST /tasks` is **400**
   (schema is template-only).
4. **List summary** — `is_function` + `input_kinds[]` peeked from payload
   (same pattern as `primary_tag`).
5. **Instantiate** — items may include `function_bindings`; validate against
   schema; append a soft-scope block to `initial_prompt` (dirs as paths, files
   and functions as `@`-mentions); strip `function_inputs` before
   `CreateTaskFromComposeJSON`.
6. **Restriction** — soft (prompt + mentions) only in v1.

## Consequences

### Positive

- Parameterized templates without a second persistence model.
- Backward-compatible instantiate body when schema is empty.

### Negative / trade-offs

- Soft scope is advisory; agents can still leave the paths.
- Symbol search quality is best-effort (see `/repo/symbols`).

## See also

- [ADR-0048](./ADR-0048-bounded-context-taskcompose.md)
- [ADR-0043](./ADR-0043-compose-git-assignment.md)
- [docs/api.md](../api.md) Task templates / Workspace repo
