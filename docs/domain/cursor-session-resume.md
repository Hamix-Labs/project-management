# Cursor session resume

ADR-0031 adds **CLI session continuity** on top of the existing harness phase ledger. A new phase row is still created for every `runner.Run`; continuing a Cursor chat does not reuse a phase row. [ADR-0085](../adr/ADR-0085-verify-resumes-execute-session.md) makes PhaseVerify continue the **execute** chat.

| | |
| --- | --- |
| **Applies to** | Cursor CLI `--resume`, harness phase ledger, `cursor_session_resume_enabled` |
| **Audience** | Contributors touching `pkgs/agents/harness/cursor_resume.go` or cursor adapter |
| **Prerequisite** | [harness.md](./harness.md) — cycle loop and phase model |
| **Decision record** | [ADR-0031](../adr/ADR-0031-cursor-session-resume-default.md), [ADR-0085](../adr/ADR-0085-verify-resumes-execute-session.md) |

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
| **Cursor resume** | `cursor-agent --resume <session_id>` | Optimization; falls back to full prompt |

## Session chains

Execute and verify share **one Cursor chat** per cycle of work: verify `--resume`s the latest terminal **execute** `session_id` (ADR-0085). Phase ledger rows remain separate (one row per `runner.Run`).

Typical in-cycle pattern:

```text
phase 1  execute  →  session E1 (new)
phase 2  verify   →  resume E1 (same chat; recovery delta)
phase 3  execute  →  resume E1 (or fresh on RetryFresh / safety deny)
phase 4  verify   →  resume E1 (last execute session)
phase 5  verify   →  resume E1 (verify-only retry; still execute's session)
```

Polish with verify: polish execute resumes the prior execute session; verify then resumes that polish execute session. Instructions-only polish (`SkipVerify`) never starts a verify run.

## Policy chokepoint

[`pkgs/agents/harness/cursor_resume.go`](../../pkgs/agents/harness/cursor_resume.go) implements `resolveCursorResume` and returns `CursorResumeDecision` with:

| `cursor_resume_mode` | Meaning |
|----------------------|---------|
| `fresh` | No `--resume`; full `composeExecutePrompt` / `buildVerifyPrompt` |
| `resume` | `--resume` + `ComposeRecoveryDelta` stdin |
| `resume_fallback` | Resume failed once; retried with full prompt |

For `PhaseVerify`, session lookup uses `LastSessionID(cycleID, PhaseExecute)` via `SessionPhaseForResume`.

## Storage

- **Write:** `task_cycle_phases.details_json.session_id` on phase complete (cursor adapter).
- **Read (execute):** `store.LastSessionID(ctx, cycleID, PhaseExecute)`.
- **Read (verify):** same — last terminal **execute** session for the cycle (ADR-0085).

Cross-cycle **Resume from failure** reads the **parent** cycle's execute session when the child has no execute session yet (including verify-only entry).

## Operator grep examples

```text
cursor_resume_mode=resume recovery_hint_kind=verify_infra_retry
deny_reason=head_drift cursor_resume_mode=fresh
cursor_resume_mode=resume_fallback
```

## Configuration

`app_settings.cursor_session_resume_enabled` (default `true`). When `false`, behavior matches always-fresh chat. See [configuration.md](../configuration.md).

## See also

- [harness.md](./harness.md) — cycle loop and worker boundary
- [retry-resume.md](./retry-resume.md) — operator Resume from failure
- [ADR-0031](../adr/ADR-0031-cursor-session-resume-default.md)
- [ADR-0085](../adr/ADR-0085-verify-resumes-execute-session.md)
