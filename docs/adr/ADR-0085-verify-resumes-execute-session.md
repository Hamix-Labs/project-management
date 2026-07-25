# ADR-0085: Verify resumes the execute Cursor session

**Date:** 2026-07-25
**Status:** Accepted
**Deciders:** Backend / agents-worker maintainers

## Context

[ADR-0084](./ADR-0084-executor-owned-verify.md) made PhaseVerify the same
execute agent (same runner). [ADR-0031](./ADR-0031-cursor-session-resume-default.md)
still kept **separate** execute vs verify Cursor session chains and forced a
**fresh** chat on the first verify after a new execute
(`verify_fresh_after_execute`). Operators therefore saw a new chat for judgment
even though the product identity was already one agent.

Polish already resumes the prior execute session for further execute work;
when polish runs verify (flagged/new criteria), it hit the same first-verify
fresh gap.

## Decision

1. **Same chat** — For `PhaseVerify`, Cursor `--resume` uses the cycle’s
   **last terminal execute** `session_id` (`LastSessionID(cycle, PhaseExecute)`),
   including verify-only retries and polish-then-verify.
2. **Delete** the `verify_fresh_after_execute` deny and
   `FirstVerifyAfterNewExecute` helper.
3. **Supersede** ADR-0031 rules that required separate execute/verify session
   chains and forbade cross-wiring execute session ids into verify runs.
4. **Retain** — new phase row per `runner.Run`; recovery deltas on resume;
   full verify prompt on `ErrResumeSession` / missing id / settings off /
   safety denials (tamper, head_drift, …); checklist/verdict UX unchanged.

## Consequences

### Positive

- Execute → verify continues one Cursor conversation (same agent and chat).
- Polish+verify shares the polish execute session for judgment.

### Negative / trade-offs

- Verify recovery deltas must make sense when continuing an **execute** chat
  (wording updated; full prompt remains the fallback).
- If execute never recorded a `session_id`, verify stays fresh (unchanged).

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Keep separate verify chain, only remove first-verify-fresh | Still starts a new chat after execute |
| Resume last verify session when present | First verify still needs execute’s id; execute id is sufficient for later retries if resume succeeded |

## See also

- [ADR-0031](./ADR-0031-cursor-session-resume-default.md) — resume framework (partially superseded here)
- [ADR-0084](./ADR-0084-executor-owned-verify.md) — same execute runner for verify
- [cursor-session-resume.md](../domain/cursor-session-resume.md)
