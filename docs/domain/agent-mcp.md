# Agent MCP platform

How Hamix exposes MCP tools to Cursor execute/verify agents for tool-only
criteria and verify report submit.

| | |
| --- | --- |
| **Applies to** | `pkgs/agents/agentmcp`, `cmd/hamix-agent-mcp`, harness Cursor wiring, `pkgs/agents/sidecar` |
| **Audience** | Operators enabling/debugging agent MCP; contributors adding tools |
| **Prerequisite** | [harness.md](./harness.md), [execute-agent.md](./execute-agent.md), [verify-agent.md](./verify-agent.md), [ADR-0089](../adr/ADR-0089-agent-mcp-platform.md) |
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
to the current cycle. The agent must call `hamix.submit_criteria_report` or
`hamix.submit_verify_report`. The MCP host validates args, writes the report
under `ReportDir`, and writes a submit receipt with the bind nonce. The harness
probe **requires** that receipt before accepting the JSON.

## Key concepts

| Term | Meaning |
| --- | --- |
| **Bind file** | `ReportDir/<cycle_id>/agent-tool-bind.json` — session contract for tools |
| **Receipt** | `criteria-report.submitted` / `verify-report.submitted` next to the report |
| **Sidecar** | Shared parse/write package used by harness and MCP |
| **Tool-only** | Orphan JSON without a matching receipt is rejected |

## Workflow

1. Before `runner.Run`, harness writes bind under `ReportDir/<cycle_id>/` and
   merges `hamix-agent` into **`<WorkingDir>/.cursor/mcp.json`** (Cursor only
   loads project MCP from the workspace root — `--add-dir` does **not**).
2. Cursor is invoked with `--approve-mcps` and `--trust` (workspace remains the
   task worktree).
3. Agent calls the submit tool; MCP writes report + receipt under ReportDir.
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
