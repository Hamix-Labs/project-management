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
events. A **fake runner** ships so CI proves the stream contract without
`CURSOR_API_KEY`. Later plans swap the runner for `@cursor/sdk`.

## Stream events

Named events: `session`, `status`, `token`, `tool`, `patch`, `error`, `done`.

Heartbeats are **SSE comments** (`: heartbeat`) every 3s while a run is
active — not a named `event: heartbeat` frame.

`session` payload includes `schema_version` (currently `1`). The SPA must
assert this before trusting the stream.

### Replay

Each session keeps a ring of the last **256** events with monotonic `id`
values. `GET /draft-assist/sessions/{id}/events` honours `Last-Event-ID`
and replays ring entries with `id > Last-Event-ID`, then tails live
events.

## Status machine

`idle → thinking → streaming | tool → idle`

Also: `cancelling` (cancel accepted), then terminal `done` with
`status=cancelled` | `done` | `failed`.

Cancel yields two frames: `status=cancelling`, then `done{status=cancelled}`.

## Concurrent runs

`POST /runs` while a run is active → **409** (`ErrRunActive`).

## MCP tools (v1)

Complete tool table for `hamix-draft-mcp`. All writes carry the session
nonce via `X-Hamix-Draft-Nonce`; a stale nonce fails closed (`ErrUnauthorized`).

| Tool | Kind | Backing endpoint | Notes |
| --- | --- | --- | --- |
| `hamix.draft_get` | read | `GET /draft-assist/sessions/{id}` | Snapshot: prompt, title, priority, criteria, tags, git binding, model |
| `hamix.draft_set_prompt` | write | `PATCH /draft-assist/sessions/{id}/prompt` | Replace prompt HTML; validated against the TipTap subset (see below) |
| `hamix.draft_patch_prompt` | write | `PATCH …/prompt` (client fetches, computes, replaces) | Bounded find/replace / append / set on current prompt |
| `hamix.draft_search_repo` | read | `GET /repo/files?worktree_id=…&q=…` | Scoped to the session's bound worktree |
| `hamix.draft_read_file` | read | `GET /repo/file?worktree_id=…&path=…` | 32 MiB preview; `warning` surfaces binary/truncation |
| `hamix.draft_list_templates` | read | `GET /task-templates` | Raw templates payload forwarded |
| `hamix.draft_search_tasks` | read | `GET /tasks` (title-substring filtered client-side) | Read-only; no dependency on server-side `q=` |

### Prompt subset validator

`pkgs/draftassist/domain/promptsubset.go::ValidateHTML` is enforced on both
sides — the MCP tool and the `PATCH …/prompt` handler. Allowed tags:
`h2`–`h4`, `p`, `ul`, `ol`, `li`, `blockquote`, `a[href]`, `br`, and the
`repo-file` mention chip (`span[data-repo-file="true"]`). `script`,
`iframe`, `style`, form controls, event handlers, and `javascript:`/`data:`
hrefs are rejected.

## Ready probe

`GET /draft-assist/ready` → `{ ready, runner: "sdk"|"fake"|"missing", reason? }`.

| `ready` | `runner` | `reason` |
| --- | --- | --- |
| true | `fake` or `sdk` | omitted |
| false | `missing` | `no_runner` |
| false | `sdk` | `missing_key` or `sidecar_down` |

## Metrics

- `taskapi_draftassist_run_first_event_ms` — histogram, accept → first status/token/tool/error
- `taskapi_draftassist_watchdog_total` — counter, silence watchdog firings

## See also

- [api.md](../api.md) — route table
- [agent-mcp.md](./agent-mcp.md) — execute/verify MCP (different tools)
