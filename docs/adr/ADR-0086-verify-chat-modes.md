# ADR-0086: Configurable verify chat modes

**Date:** 2026-07-26
**Status:** Accepted
**Deciders:** Backend / agents-worker maintainers

## Context

[ADR-0085](./ADR-0085-verify-resumes-execute-session.md) made PhaseVerify always
`--resume` the execute Cursor session (same chat). Operators also need the older
[ADR-0031](./ADR-0031-cursor-session-resume-default.md) separate verify session
chain (fresh chat after each execute; resume verify-only retries on the verify
chain). Both remain executor-owned ([ADR-0084](./ADR-0084-executor-owned-verify.md)).

## Decision

1. **Modes** — `same_chat` (ADR-0085) and `different_chat` (ADR-0031 restore:
   separate verify chain + `verify_fresh_after_execute`).
2. **Global default** — `app_settings.verify_chat_mode`, default `same_chat`.
3. **Per-task override** — `tasks.verify_chat_mode`: empty inherits settings;
   otherwise pins `same_chat` or `different_chat`.
4. **Resolve** — `settings/domain.EffectiveVerifyChatMode(task, settings)`.
5. **Hard-fail session contract** (missing execute `session_id` /
   `ErrResumeSession`) applies only when the **effective** mode is `same_chat`.
6. **Same runner** — ADR-0084 unchanged; wire kind stays `execute_agent`.

## Consequences

### Positive

- Operators choose continuity vs independent judgment chat globally or per task.
- Default preserves current production behavior.

### Negative / trade-offs

- Two resume policies to document and test.
- Different-chat spends a fresh verify context after each execute (intentional).

## See also

- [ADR-0085](./ADR-0085-verify-resumes-execute-session.md) — amended: same-chat is the default mode, not the sole policy
- [ADR-0031](./ADR-0031-cursor-session-resume-default.md) — resume framework
- [cursor-session-resume.md](../domain/cursor-session-resume.md)
