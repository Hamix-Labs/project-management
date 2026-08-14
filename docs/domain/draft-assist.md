# Draft assist

In-memory sessions and SSE for prompt-box LLM help while composing a task
(ADR-0101). Execute/verify still uses the Cursor CLI; this path is separate.

| | |
| --- | --- |
| **Applies to** | `pkgs/draftassist`, `cmd/hamix-draft-mcp`, taskapi `/draft-assist/*` |
| **Audience** | Contributors extending draft-assist; Plan 3/4 implementers |
| **Prerequisite** | [ADR-0101](../adr/ADR-0101-draft-assist-platform.md), [task-draft-ai design](../design/task-draft-ai.md) |

## Overview

SPA opens a session, attaches SSE, POSTs a run (202), and streams named
events. A **fake runner** ships in Wave A so CI proves the stream contract
without `CURSOR_KEY`. Plan 3 swaps the runner for `@cursor/sdk`.

## Stream events

`session`, `status`, `token`, `tool`, `patch`, `error`, `done`, plus SSE
comment heartbeats every 3s while a run is active.

## Status machine

`idle → thinking → streaming | tool → idle` (also `cancelled`, `failed`).

## MCP tools (v1)

`hamix.draft_get`, `hamix.draft_set_prompt`, `hamix.draft_patch_prompt`.
Prompt write only; nonce fail-closed.

## Ready probe

`GET /draft-assist/ready` → `{ ready, runner: "fake"|"missing", reason? }`.

## See also

- [api.md](../api.md) — route table
- [agent-mcp.md](./agent-mcp.md) — execute/verify MCP (different tools)
