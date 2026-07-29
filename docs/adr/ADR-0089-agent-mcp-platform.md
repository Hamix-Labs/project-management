# ADR-0089: Hamix agent MCP platform

**Date:** 2026-07-27
**Status:** Accepted
**Deciders:** Backend / agent maintainers

## Context

Execute/verify agents freeform-wrote side-channel JSON (`criteria-report.json` /
`verify-report.json`). Path/schema/ID mistakes caused full re-execute loops.
Cursor CLI supports MCP tools; Hamix needed an agent-facing tool host that can
grow beyond report submit without re-architecture.

## Decision

1. **Packages** — `pkgs/agents/sidecar` owns shared report parse/validate/write
   and submit receipts. `pkgs/agents/agentmcp` is the MCP host + tool registry;
   binary `cmd/hamix-agent-mcp` serves stdio MCP with `--bind <path>`.
2. **Tool-only default** — `app_settings.agent_mcp_enabled` defaults to **true**.
   Harness accepts reports only with a matching MCP submit receipt nonce. Freeform
   Write alone is treated as missing/invalid. Set the flag **false** only as an
   emergency kill-switch to restore legacy Write prompts and probes.
3. **Bind file** — Per phase, harness writes `agent-tool-bind.json` under
   `ReportDir/<cycle_id>/` (never via `HAMIX_*` env into Cursor; the adapter
   strips those keys). Bind carries task/cycle/phase, report_dir, working_dir,
   active/locked IDs, and `submit_nonce`.
4. **MCP config in workspace** — Cursor CLI loads project MCP only from
   `<workspace>/.cursor/mcp.json` (plus `~/.cursor/mcp.json`). `--add-dir` does
   **not** discover MCP config (confirmed spike failure). Harness merges
   `hamix-agent` into the worktree `.cursor/mcp.json` for the run and
   restores/removes it on cycle terminate. `.cursor/mcp.json` is exempt from
   verify integrity added-path checks.
5. **Cursor flags** — `--approve-mcps` and `--trust` for headless MCP approval.
6. **Trust boundary** — Tools never mark checklist done, never run
   `verify_commands`, never mutate cycle lifecycle.
7. **Extensibility** — Tools register via a small `Tool` interface + groups;
   bind schema is versioned (`bind_schema_version`).
8. **v1 tools** — `hamix.submit_criteria_report`, `hamix.submit_verify_report`
   only. Future tools (`get_cycle_contract`, evidence, git) plug into the same
   registry.

### Spike (headless Cursor flags)

Confirmed on pinned `cursor-agent` (2026.07.23):

- `--approve-mcps` + `--trust` + workspace `.cursor/mcp.json` → tools callable
- `--add-dir` with MCP config outside the workspace → **not** loaded (only
  global `~/.cursor/mcp.json` servers appear)

## Consequences

### Positive

- Single parser contract for harness + MCP (no drift).
- Deterministic report ingest; fewer path/schema re-execute loops.
- Extensible host without one-off scripts.

### Negative / trade-offs

- Requires `hamix-agent-mcp` on PATH for Cursor runs when MCP is enabled;
  missing binary fails loud (no silent freeform fallback while kill-switch is on).
- Real Cursor MCP smoke remains an operator-gated merge check
  (`HAMIX_TEST_REAL_CURSOR=1`).
