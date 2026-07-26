# ADR-0085: Verify resumes the execute Cursor session

**Date:** 2026-07-25
**Status:** Accepted (amended: optional verify model; hard-fail missing session / resume; default mode under [ADR-0086](./ADR-0086-verify-chat-modes.md))
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

Soft-falling back to a full fresh verify (or execute) after a missing
`session_id` or `ErrResumeSession` re-sent full context and broke the
same-chat product rule.

## Decision

1. **Same chat (default mode)** — When effective `verify_chat_mode` is
   `same_chat` ([ADR-0086](./ADR-0086-verify-chat-modes.md)), PhaseVerify
   Cursor `--resume` uses the cycle’s **last terminal execute** `session_id`
   (`LastSessionID(cycle, PhaseExecute)`), including verify-only retries and
   polish-then-verify.
2. **Remove sole-policy denials** — The unconditional `verify_fresh_after_execute`
   deny and always-on separate chains are no longer the only product path;
   `different_chat` restores them when selected (ADR-0086).
3. **Supersede** ADR-0031 rules that *required* separate execute/verify session
   chains for all tasks; those rules apply only under `different_chat`.
4. **Optional verify model** — `app_settings.verify_model` may pin a different
   `--model` on the verify `Run` while still `--resume`ing the execute chat
   under `same_chat`. Empty inherits the execute effective model (task pin,
   else `cursor_model`).
5. **Hard-fail session contract** (Cursor + `cursor_session_resume_enabled` +
   effective `same_chat`):
   - Successful execute without `session_id` → fail execute phase
     (`cursor_missing_session_id`); do not enter verify.
   - Verify with deny `no_session_id` → fail verify; **zero** Cursor Runs.
   - `ErrResumeSession` on execute or verify → fail; **no** soft-fresh full
     prompt retry (avoids double token spend / new chat).
6. **Retain** — new phase row per `runner.Run`; recovery deltas on resume;
   intentional fresh execute paths (`settings_disabled`, Start over /
   `retry_fresh`, tamper, `head_drift`, workspace mismatch); checklist/verdict
   UX unchanged. When resume is **disabled**, verify may still run fresh.

## Consequences

### Positive

- Execute → verify continues one Cursor conversation (same agent and chat).
- Operators can pin a cheaper/stronger model for verify without a second chat.
- Missing chat ids and resume failures surface as operator-readable
  `failure_kind` / `standardized_message` / `cycle_failed.failure_summary`.

### Negative / trade-offs

- Verify recovery deltas must make sense when continuing an **execute** chat.
- Infra/CLI bugs that omit `session_id` become hard cycle failures instead of
  silent fresh chats.

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Keep separate verify chain, only remove first-verify-fresh | Still starts a new chat after execute |
| Resume last verify session when present | First verify still needs execute’s id; execute id is sufficient for later retries if resume succeeded |
| Soft-fresh on missing id / `ErrResumeSession` | Burns ~full context again; violates same-chat hard requirement |

## See also

- [ADR-0031](./ADR-0031-cursor-session-resume-default.md) — resume framework (partially superseded here)
- [ADR-0084](./ADR-0084-executor-owned-verify.md) — same execute runner for verify
- [cursor-session-resume.md](../domain/cursor-session-resume.md)
