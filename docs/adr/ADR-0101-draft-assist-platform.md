# ADR-0101: Draft-assist platform (session BC + MCP)

**Date:** 2026-08-12
**Status:** Accepted
**Deciders:** Backend / draft-editor maintainers

## Context

Task draft AI (see `docs/design/task-draft-ai.md`) needs a durable, multi-turn
agent that patches the initial-prompt HTML while the operator composes a task.
The SPA cannot run the Cursor SDK, and Go cannot import `@cursor/sdk`. Draft
sessions live only as long as the modal, so there is no persistence
requirement. The stream contract the SPA will trust must ship **before** the
Node sidecar exists so front-end and runner work can proceed in parallel.

## Decision

1. **New bounded context** — `pkgs/draftassist/{domain,contract,store,handler}`
   owns draft-assist sessions. Sessions are keyed by id + per-session `nonce`
   and live entirely in-memory (no Postgres row). One session is bound to one
   `worktree_id` and one form snapshot (title / prompt / criteria / …).
2. **HTTP surface** — routes registered from
   `pkgs/tasks/handler/handler_routes.go`:
   - `POST /draft-assist/sessions` — bind worktree, seed snapshot
   - `PUT  /draft-assist/sessions/{id}/snapshot` — refresh snapshot
   - `GET  /draft-assist/sessions/{id}/events` — SSE (attach before send)
   - `POST /draft-assist/sessions/{id}/runs` — **202** `{ run_id }` immediately
   - `POST /draft-assist/sessions/{id}/runs/{runId}/cancel`
   - `DELETE /draft-assist/sessions/{id}`
   - `GET  /draft-assist/ready` — `{ ready, runner: "fake"|"missing", reason? }`
3. **Stream contract** — named SSE events (`event: …`, `data: …`, `id: …`):
   `session`, `status`, `token`, `tool`, `patch`, `error`, `done`, and a
   `: heartbeat` comment every ≥3s while a run is active. Every event flushes.
   `Last-Event-ID` replay works from the per-session ring.
4. **Runner interface** — `contract.Runner` is the seam between the handler
   and the model. This PR ships a **fake runner** that emits
   `status=thinking` immediately, first `token` within a few ms, an optional
   delayed `tool` for watchdog tests, then `done`. Cancel yields a terminal
   `status=cancelled` + `done`. Plan 3 swaps in the `@cursor/sdk` sidecar
   without handler changes.
5. **MCP host** — `pkgs/draftassist/mcp` + `cmd/hamix-draft-mcp` mirror the
   agent MCP layout (`pkgs/agents/agentmcp`, ADR-0089). Tools:
   `hamix.draft_get`, `hamix.draft_set_prompt`, `hamix.draft_patch_prompt`,
   `hamix.draft_search_repo`, `hamix.draft_read_file`,
   `hamix.draft_list_templates`, `hamix.draft_search_tasks`. Bind file carries
   `{ session_id, nonce, taskapi_base_url }`; every write tool fails closed
   when the nonce does not match the session's live nonce. Prompt is the only
   writable field in v1 (title / criteria writes are out).
6. **Latency** — POST run returns `202` before the runner produces any output,
   consistent with the umbrella latency standards. SSE must already be open;
   readiness is signalled by the first `session` event.

Non-goals for this ADR: `@cursor/sdk`, Node sidecar packaging, TipTap
Space-for-AI, and the SPA compose page. Those live in Plan 3 and Plan 4.

## Consequences

### Positive

- Front-end and Node sidecar work can start against a Go-owned contract with
  green tests instead of a mock.
- Same MCP shape as ADR-0089 keeps the bind / nonce / tool-only story
  consistent for reviewers.
- Fake runner is enough to exercise heartbeat, cancel, and nonce-fail-closed
  contracts in CI without `CURSOR_API_KEY`.

### Negative / trade-offs

- Sessions are process-local; a taskapi restart drops the modal state. Modal
  UX already implies "recompose after crash", so this is acceptable in v1.
- The SSE hub is **per session** (not the global `pkgs/tasks/realtime` hub) —
  new code path to review, and heartbeats are scheduled locally instead of via
  the shared hub. Justified by the different lifecycle (SSE is attached before
  send and only for one modal).

## Alternatives Considered

| Alternative | Reason rejected |
|---|---|
| Reuse `/events` global SSE hub | Different audience (one modal, not all tabs); would need per-session filters + heartbeats grafted into the shared hub. |
| Persist sessions in Postgres | Modal-lifetime state; no cross-tab / cross-device requirement in v1. |
| Ship the SDK sidecar in this PR | Blocks the contract on a `CURSOR_API_KEY`-gated Node process — plan 2 must stay green in CI without live model output. |

## See also

- [docs/design/task-draft-ai.md](../design/task-draft-ai.md)
- [docs/domain/draft-assist.md](../domain/draft-assist.md)
- [ADR-0089](./ADR-0089-agent-mcp-platform.md) — MCP host / bind pattern
