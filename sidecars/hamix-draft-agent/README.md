# hamix-draft-agent

Node sidecar that hosts `@cursor/sdk` for the Hamix draft-assist prompt agent
(see [`docs/design/task-draft-ai.md`](../../docs/design/task-draft-ai.md) and
[`docs/domain/draft-assist.md`](../../docs/domain/draft-assist.md)).

Loopback-only. Taskapi is the only intended caller.

## Endpoints

| Method + path | Purpose |
| --- | --- |
| `POST /runs` | Start (or follow-up) a run on a session; SSE stream |
| `POST /runs/{run_id}/cancel` | Ask the SDK to cancel; 202 accepted |
| `GET  /healthz` | `{ ok, sdk_version?, agents_active }` |
| `GET  /readyz` | `{ ready, reason? }` — `missing_key` when `CURSOR_API_KEY` is unset |

`POST /runs` body:

```json
{
  "session_id": "sess-…",
  "run_id":     "run-…",
  "user_message": "tighten this prompt",
  "snapshot":     { "title": "…", "prompt": "<p>…</p>", "priority": "P2" },
  "worktree_cwd": "C:/repos/some-worktree",
  "model":     "composer-2.5",
  "agent_id":  "opt — Agent.resume target"
}
```

SSE named events match the Go domain (`session`, `status`, `token`, `tool`,
`patch`, `error`, `done`). Heartbeat is an SSE comment every 3s while the run
is active.

## Runtime shape

- `Agent.create` per draft-assist session, `local.cwd = worktree_cwd`,
  `settingSources: []`, tools `read, grep, glob, ls, mcp`, disallowed
  `shell, edit, task`.
- MCP: `mcpServers["hamix-draft"] = { command: "hamix-draft-mcp", args: ["--bind", <tmp>] }`.
  The sidecar writes the bind JSON (see
  [`pkgs/draftassist/mcp/BindFile`](../../pkgs/draftassist/mcp/server.go)) to a
  per-session tmp dir. Plan 4 replaces the local writer with `taskapi`'s
  canonical `WriteBind`.
- Cancel: `run.cancel` if supported. The stream ends with `status=cancelling`
  then `done{status=cancelled}`.
- Resume: pass `agent_id` (previously observed on any run) to `Agent.resume`.
  Inline MCP servers are always re-passed because the SDK does not persist them.

## Build

```powershell
pnpm --dir sidecars/hamix-draft-agent install
pnpm --dir sidecars/hamix-draft-agent build
pnpm --dir sidecars/hamix-draft-agent test
```

`scripts/dev.ps1` / `scripts/dev.sh` at repo root build the bundle to
`dist/hamix-draft-agent.js`, copy it (with a `.cmd` shim on Windows) to the
repo root, and prepend the repo root to `PATH` so `taskapi` can spawn it.

Ephemeral port for supervisor discovery:

```powershell
node .\sidecars\hamix-draft-agent\dist\hamix-draft-agent.js --port 0
# stdout: listening on 63421
```

## Env

| Var | Required | Notes |
| --- | --- | --- |
| `CURSOR_API_KEY` | yes for real SDK | `readyz` reports `missing_key` when unset |
| `PORT` | no | Overridden by `--port` if both are set |

## Tests

Vitest specs live under `src/__tests__/`. They use `MockAgentPort` — no real
SDK install and no `CURSOR_API_KEY` are required. The build itself leaves
`@cursor/sdk` external so the bundle succeeds even offline.
