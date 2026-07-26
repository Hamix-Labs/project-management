# Cursor session resume

ADR-0031 adds **CLI session continuity** on top of the existing harness phase ledger. A new phase row is still created for every `runner.Run`; continuing a Cursor chat does not reuse a phase row. [ADR-0086](../adr/ADR-0086-verify-chat-modes.md) selects whether PhaseVerify continues the execute chat (`same_chat`, default) or uses a separate verify chain (`different_chat`).

| | |
| --- | --- |
| **Applies to** | Cursor CLI `--resume`, harness phase ledger, `cursor_session_resume_enabled`, `verify_chat_mode` |
| **Audience** | Contributors touching `pkgs/agents/harness/cursor_resume.go` or cursor adapter |
| **Prerequisite** | [harness.md](./harness.md) — cycle loop and phase model |
| **Decision record** | [ADR-0031](../adr/ADR-0031-cursor-session-resume-default.md), [ADR-0085](../adr/ADR-0085-verify-resumes-execute-session.md), [ADR-0086](../adr/ADR-0086-verify-chat-modes.md) |

## In this article

- [Two resume layers](#two-resume-layers)
- [Session chains](#session-chains)
- [Policy chokepoint](#policy-chokepoint)
- [Storage](#storage)
- [Operator grep examples](#operator-grep-examples)
- [Configuration](#configuration)
- [See also](#see-also)

## Two resume layers

| Layer | Mechanism | Authority |
|-------|-----------|-----------|
| **Harness resume** | Checkpoint, continuation bundle, git state | DB + phase ledger |
| **Cursor resume** | `cursor-agent --resume <session_id>` | Required when resume enabled under `same_chat`; hard-fails instead of opening a new chat |

## Session chains

Effective `verify_chat_mode` (task override, else `app_settings`, else `same_chat`) selects the chain:

**`same_chat` (default):** verify `--resume`s the latest terminal **execute** `session_id` (ADR-0085). Optional `verify_model` may change `--model` without a new chat.

```text
phase 1  execute  →  session E1 (new; session_id required when resume on)
phase 2  verify   →  resume E1 (same chat; optional verify_model; recovery delta)
phase 3  execute  →  resume E1 (or fresh on RetryFresh / safety deny)
phase 4  verify   →  resume E1 (last execute session)
phase 5  verify   →  resume E1 (verify-only retry; still execute's session)
```

**`different_chat`:** separate verify chain + fresh after each execute (ADR-0031 restore).

```text
phase 1  execute  →  session E1 (new/resume execute chain)
phase 2  verify   →  session V1 (fresh; deny verify_fresh_after_execute)
phase 3  execute  →  resume E1
phase 4  verify   →  session V2 (fresh after new execute)
phase 5  verify   →  resume V2 (verify-only retry)
```

Polish with verify: polish execute resumes the prior execute session; verify then follows the effective mode. Instructions-only polish (`SkipVerify`) never starts a verify run.

## Policy chokepoint

[`pkgs/agents/harness/cursor_resume.go`](../../pkgs/agents/harness/cursor_resume.go) implements `resolveCursorResume` and returns `CursorResumeDecision` with:

| `cursor_resume_mode` | Meaning |
|----------------------|---------|
| `fresh` | No `--resume`; full `composeExecutePrompt` / `buildVerifyPrompt` (intentional denials or resume disabled) |
| `resume` | `--resume` + `ComposeRecoveryDelta` stdin |

Hard failures under **`same_chat`** (not soft-fresh):

| Condition | Kind / reason |
|-----------|----------------|
| Execute success, resume on, empty `session_id` | `cursor_missing_session_id` |
| Verify needs resume, empty execute `session_id` | `cursor_missing_session_id` |
| `--resume` rejected (`ErrResumeSession`) | `cursor_resume_session` |

Under **`different_chat`**, first verify after execute intentionally opens a fresh chat (`deny_reason=verify_fresh_after_execute`); missing execute session does not block verify.

For `PhaseVerify`, session lookup uses `SessionPhaseForResume(phase, mode)` → `PhaseExecute` (`same_chat`) or `PhaseVerify` (`different_chat`).

## Storage

- **Write:** `task_cycle_phases.details_json.session_id` is written at two moments:
  1. **First stream sighting (mid-run):** the cursor adapter fires `runner.Request.OnSessionID` once, on the first NDJSON frame that carries a non-empty `session_id`. The harness patches the running phase row via `store.PatchPhaseDetails`. First-wins: a later frame reporting a different id does not overwrite the stored value.
  2. **Any terminal outcome:** the cursor adapter surfaces `session_id` on `Result.Details` for success, `is_error=true`, non-zero exit, timeout/cancel, and exec failure — extracted from captured stdout (init frame is enough). `CompletePhase` merges those details into the row.
- **Read (execute):** `store.LastSessionID(ctx, cycleID, PhaseExecute)`.
- **Read (verify):** last terminal session for `SessionPhaseForResume` (execute under `same_chat`, verify under `different_chat`).

The mid-run patch guarantees a durable id even when the run terminates before emitting a `result` event (timeout, panic, kill). Terminal-outcome writes remain the audit anchor.

Cross-cycle **Resume from failure** reads the **parent** cycle's session for the entry phase when the child has none yet.

## Operator grep examples

```text
cursor_resume_mode=resume recovery_hint_kind=verify_infra_retry
deny_reason=verify_fresh_after_execute cursor_resume_mode=fresh
deny_reason=head_drift cursor_resume_mode=fresh
failure_kind=cursor_missing_session_id
failure_kind=cursor_resume_session
```

## Configuration

- `app_settings.cursor_session_resume_enabled` (default `true`). When `false`, behavior matches always-fresh chat (no session_id hard-require).
- `app_settings.verify_chat_mode` (default `same_chat`) — global verify chat policy; task `verify_chat_mode` overrides when non-empty.
- `app_settings.verify_model` — optional Cursor `--model` for PhaseVerify; empty inherits execute effective model.

See [configuration.md](../configuration.md).

## See also

- [harness.md](./harness.md) — cycle loop and worker boundary
- [retry-resume.md](./retry-resume.md) — operator Resume from failure
- [ADR-0086](../adr/ADR-0086-verify-chat-modes.md) — configurable verify chat modes
- [ADR-0085](../adr/ADR-0085-verify-resumes-execute-session.md) — same-chat verify default
- [ADR-0031](../adr/ADR-0031-cursor-session-resume-default.md)
