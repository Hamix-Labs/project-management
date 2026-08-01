# Agent MCP platform

How Hamix exposes MCP tools to Cursor execute/verify agents for tool-only
criteria/verify report submit and authoritative git commits.

| | |
| --- | --- |
| **Applies to** | `pkgs/agents/agentmcp`, `cmd/hamix-agent-mcp`, harness Cursor wiring, `pkgs/agents/sidecar` |
| **Audience** | Operators enabling/debugging agent MCP; contributors adding tools |
| **Prerequisite** | [harness.md](./harness.md), [execute-agent.md](./execute-agent.md), [verify-agent.md](./verify-agent.md), [ADR-0089](../adr/ADR-0089-agent-mcp-platform.md), [ADR-0093](../adr/ADR-0093-mcp-commit-register.md) |
| **Status** | Product default **on** (`agent_mcp_enabled=true`) |

## In this article

- [Overview](#overview)
- [Key concepts](#key-concepts)
- [Workflow](#workflow)
- [Configuration](#configuration)
- [Cursor flags (spike)](#cursor-flags-spike)
- [Trust boundary](#trust-boundary)
- [Kill-switch](#kill-switch)
- [See also](#see-also)

## Overview

By default every execute/verify Cursor run gets a Hamix stdio MCP server bound
to the current cycle. Execute agents must call `hamix.commit` for every new git
commit (index only; stage via Shell) and `hamix.submit_criteria_report` for the
criteria report. Open-PR runs call `hamix.create_pull_request` (push + `gh pr create`
+ receipt) instead of criteria submit. Verify agents call `hamix.submit_verify_report`.
The MCP host validates args, writes artifacts under `ReportDir`, and writes submit
receipts with the bind nonce where applicable. The harness requires the criteria/verify
or pull-request receipt before accepting those outcomes, and validates the commit register
against `cycle_base_sha..HEAD` after execute ([cycle-commits.md](./cycle-commits.md)).

## Key concepts

| Term | Meaning |
| --- | --- |
| **Bind file** | `ReportDir/<cycle_id>/agent-tool-bind.json` — session contract for tools |
| **Receipt** | `criteria-report.submitted` / `verify-report.submitted` / `pull-request.submitted` next to the report |
| **Commit register** | `commit-register.json` — SHAs appended by `hamix.commit` (ADR-0093) |
| **Sidecar** | Shared parse/write package used by harness and MCP |
| **Tool-only** | Orphan criteria/verify JSON without a matching receipt is rejected |

## Workflow

1. Before `runner.Run`, harness writes bind under `ReportDir/<cycle_id>/` and
   merges `hamix-agent` into **`<WorkingDir>/.cursor/mcp.json`** (Cursor only
   loads project MCP from the workspace root — `--add-dir` does **not**).
2. Cursor is invoked with `--approve-mcps` and `--trust` (workspace remains the
   task worktree).
3. Agent calls `hamix.commit` as needed (append register) and the submit tool;
   MCP writes report + receipt under ReportDir.
4. Harness requires receipt nonce match, then parses via sidecar.
5. On cycle terminate, harness restores/removes the managed workspace mcp.json.

## Configuration

| Setting | Default | Role |
| --- | --- | --- |
| `agent_mcp_enabled` | `true` | Product path. `false` = emergency legacy freeform Write |

Binary: `hamix-agent-mcp` must be on `PATH` when MCP is enabled (`dev.ps1` /
`dev.sh` build it into the repo root and prepend PATH).

## Cursor flags (spike)

Verified on current `cursor-agent` (2026.07.23):

- `--approve-mcps` — auto-approve MCP servers in headless/`--force` runs
- `--trust` — trust workspace without prompting (needed for headless MCP)
- `--workspace` — task worktree; **must** contain `.cursor/mcp.json` for Hamix tools
- `--add-dir` — does **not** load MCP config from the added directory (failed spike)

## Trust boundary

MCP tools must not:

- Mark checklist items done
- Run `verify_commands`
- Mutate cycle lifecycle / task status

## Kill-switch

Set `agent_mcp_enabled=false` only to restore legacy prompt instructions and
accept freeform report Write without receipts. Do not leave this off in normal
operation.

## See also

- [ADR-0089](../adr/ADR-0089-agent-mcp-platform.md)
- [agent-tools-audit.md](./agent-tools-audit.md)
- [configuration.md](../configuration.md)
