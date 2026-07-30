# ADR-0090: Command-only verify and execute_claim

**Date:** 2026-07-28  
**Status:** Superseded by [ADR-0091](./ADR-0091-execute-owns-verify-commands.md)  
**Deciders:** Backend / agent maintainers

## Context

Execute agents submit structured criteria claims via MCP (`hamix.submit_criteria_report`). A second narrative PhaseVerify pass (especially `same_chat`) largely re-graded the same claims. Worker `verify_commands` remain the valuable independent evidence. In-cycle execute↔verify retries (`verify_max_retries`, verify-only vs full re-execute) existed to recover from narrative disagreement and are obsolete under tool-only claims.

## Decision

1. **Claim-only criteria** (no `verify_commands`) with `claimed_done: true` are accepted by the harness without a Cursor verify run. Completions use `verified_by=execute_claim`.
2. **Command-backed criteria** still run worker commands, then a PhaseVerify LLM whose **only** job is to judge whether each command’s `expected_outcome` matches captured output. On pass, completion evidence composes execute claim + verify interpretation; `verified_by=execute_agent`.
3. **One-shot cycle:** one execute, at most one command-verify. Any verify/gate failure terminates the cycle (`verification_failed…`). No in-cycle re-execute or verify-only retry. Operators recover via Retry / Start over.
4. **Command-verify chat:** hardcoded `same_chat` (resume execute session). `app_settings.verify_chat_mode` and `tasks.verify_chat_mode` were removed.
5. Exit code 0 on verify commands does **not** auto-pass ([ADR-0012](./ADR-0012-structured-verify-commands.md)).

## Consequences

### Positive

- Removes redundant self-grading for claim-only tasks.
- Aligns verify LLM cost with command evidence.
- Clear audit: `execute_claim` vs `execute_agent`.

### Negative / trade-offs

- Stricter failures (no in-cycle retry budget); recovery is operator **Retry** / **Start over** (new cycle).
- `verify_max_retries` setting removed with in-cycle retry code.

## Supersedes

- In-cycle retry policy of [ADR-0028](./ADR-0028-in-cycle-verify-only-retry.md).
- Operator-facing chat-mode choice of [ADR-0086](./ADR-0086-verify-chat-modes.md) (hardcoded `same_chat`).
