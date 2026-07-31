# ADR-0092: Execute owns verify commands

**Date:** 2026-07-30  
**Status:** Accepted  
**Deciders:** Backend / agent maintainers

## Context

Command-backed criteria previously ran worker shell checks and a same-chat PhaseVerify LLM that matched `expected_outcome` to captured output ([ADR-0090](./ADR-0090-command-only-verify.md)). The execute agent was not shown those commands, so it could not self-check before claiming done. A second Cursor pass and harness shell runner added cost and failure surface without independent trust value operators wanted to keep — agent honesty is treated as Cursor’s concern.

## Decision

1. **Inject verify commands into the execute prompt** — each active criterion lists `command` and `expected_outcome`. The agent must run those commands in the worktree and only claim done when outcomes match.
2. **Acceptance is claim-only for all criteria** — `claimed_done: true` → `verified_by=execute_claim`. No worker `RunCriterionCommands`. No PhaseVerify Cursor run. No `hamix.submit_verify_report`.
3. **MCP** — keep `hamix.submit_criteria_report` (`claimed_done` + free-text `evidence`). Evidence should summarize work and command self-checks.
4. **No synthetic PhaseVerify row** on the happy path — execute success → claim gate → completions → task done. Historical `phase=verify` / `execute_agent` / `command_runs` rows remain readable.
5. **Authoring unchanged** — operators still attach `verify_commands` on checklist items; semantics are “agent must self-check,” not “worker re-runs.”

## Consequences

### Positive

- Execute agent knows the pass bar before claiming.
- Removes PhaseVerify LLM cost, same-chat resume complexity, and worker shell orchestration.
- Single acceptance path (`execute_claim`).

### Negative / trade-offs

- Harness no longer independently re-runs commands; false claims are not caught by Hamix.
- `app_settings.verify_model` is unused for new cycles (kept for API compatibility).
- `verified_by=execute_agent` is legacy for new successes.

## Supersedes

- Command-backed PhaseVerify path of [ADR-0090](./ADR-0090-command-only-verify.md).
- Worker-run shell evidence of [ADR-0012](./ADR-0012-structured-verify-commands.md) (authoring schema retained).
- Executor-owned PhaseVerify Cursor resume of [ADR-0084](./ADR-0084-executor-owned-verify.md) / [ADR-0085](./ADR-0085-verify-resumes-execute-session.md) for new cycles.
