# ADR-0012: Structured Verify Commands for Done Criteria

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-15  
**Status:** Superseded by [ADR-0092](./ADR-0092-execute-owns-verify-commands.md) for runtime (authoring schema retained)  
**Deciders:** T2A maintainers

## Context

Early T2A stored a `check` column on checklist items for shell verification. That column was merged into free-form `text` and dropped ([migrateChecklistCheckToText](pkgs/tasks/postgres/postgres.go)). ADR-0003 introduced a separate verify LLM pass with no deterministic auto-pass; [ADR-0084](ADR-0084-executor-owned-verify.md) made that pass executor-owned (same agent as execute).

Operators still need repeatable, machine-checkable evidence (test output, lint results) without relying on the execute agent to run commands honestly. The verify phase can interpret structured output, but the worker must produce a stable evidence bundle first.

## Decision

1. **Child table `task_checklist_item_commands`** — optional ordered shell checks per criterion (`command`, `expected_outcome`, optional `timeout_seconds`), capped at five per item. Criterion `text` remains required.

2. **Verify phase execution** — after the execute agent claims `claimed_done: true` and before `runLLMVerify`, the worker runs attached commands sequentially in the task worktree using `adapterkit.DefaultExec` (not the LLM runner). Only criteria with commands are executed.

3. **Temp-file evidence contract** — under `<ReportDir>/<cycleId>/checks/<criterionId>/<seq>.{stdout,stderr,meta.json}`. Streams truncate at 256 KiB; meta records exit code, duration, truncation, errors, and optional timeout. Failures do not skip the LLM verify pass.

4. **LLM remains verdict authority** — success still requires `verified_by=execute_agent`. Exit code 0 does not auto-pass (`deterministic_check` stays legacy-only).

5. **Audit mirror `task_cycle_command_runs`** — exposed on `GET .../verdicts` as `command_runs[]` for the SPA timeline.

6. **Timeout** — per command via optional `timeout_seconds` on the command row. Omit/null means no wall-clock timeout (parent cycle cancel only). The former global `app_settings.verify_command_timeout_seconds` was removed.

## Consequences

### Positive

- Deterministic evidence the verify agent can cite without trusting execute self-report alone.
- Clear separation: commands produce artifacts; LLM judges against text + `expected_outcome`.
- Child-table design allows a future global command template library (copy-on-insert) without schema churn.
- While each command runs, the worker emits live `agent_run_progress` / `task_cycle_stream_events` with `source=worker`, `tool=verify_command`, and subtypes `started` | `running` | `completed` | `failed`, so the SPA cycles ticker can show “Running: …” instead of an empty wait state before the LLM verify agent starts.
- Operators choose per-command cost: unbounded by default, or an explicit seconds cap when they want one.

### Negative / Trade-offs

- Verify commands that mutate the repo may trigger `verify_tampered` on post-snapshot integrity checks — operators must use read-only checks.
- Trusted-operator surface: commands are not agent-authored, but shell injection risk remains; bounded by optional timeout and output caps. Unlimited commands can hold the worker until cycle cancel.
- Extra verify latency when many commands are attached.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| JSON blob on `task_checklist_items` | Poor ordering/CRUD; blocks template library V2 |
| Re-enable auto-pass on exit 0 | Conflicts with executor-owned verify model ([ADR-0084](ADR-0084-executor-owned-verify.md)) |
| Run commands in execute phase | Execute agent could interfere; verify needs post-execute snapshot |
| Inline megabyte payloads in API | Temp files + meta paths keep HTTP and DB bounded |
