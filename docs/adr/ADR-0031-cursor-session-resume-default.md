# ADR-0031: Cursor session resume by default

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-19  
**Status:** Partially superseded — separate execute/verify session chains and `verify_fresh_after_execute` replaced by [ADR-0085](./ADR-0085-verify-resumes-execute-session.md) (verify resumes the execute Cursor session). Resume-by-default, recovery deltas, settings opt-out, and phase-ledger rules otherwise remain.  
**Deciders:** T2A maintainers

## Context

ADR-0006 treated Cursor `session_id` as audit-only: every `runner.Run` started a fresh CLI chat with a full recomposed prompt. The Cursor CLI supports `--resume <session_id>` so follow-up execute/verify attempts can continue the same conversation with short recovery deltas.

T2A still creates a new `task_cycle_phases` row per `runner.Run` (ADR-0030 correlation, metrics, audit). This ADR adds **CLI session continuity** as an optimization layered on unchanged phase ledger semantics.

## Decision

1. **Resume by default** for execute when a prior execute `session_id` exists; for verify, resume the cycle’s last **execute** session ([ADR-0085](./ADR-0085-verify-resumes-execute-session.md)).
2. **stdin on resume** uses structured recovery deltas (`ComposeRecoveryDelta`) instead of full prompt rehydration.
3. **Deny-list** forces fresh chat: RetryFresh/Start over, first in chain (no session id), HEAD drift, tamper, missing id, workspace mismatch, `--resume` failure (`resume_fallback` + full prompt). First verify after execute **resumes** the execute session ([ADR-0085](./ADR-0085-verify-resumes-execute-session.md)).
4. **Cross-cycle RetryResume** resumes the **parent cycle's** session for the entry phase (execute session for verify-only entry).
5. **Settings opt-out:** `app_settings.cursor_session_resume_enabled` (default `true`).
6. **Storage:** read `session_id` via `LastSessionID(cycleID, phase)` from `details_json`; no new table. Verify reads the **execute** phase session.

## MUST / MUST NOT

**MUST:** new phase row per run; fallback to full prompt on resume failure; skip `ScrubCycleArtifacts` on execute resume only. PhaseVerify resumes the execute Cursor session ([ADR-0085](./ADR-0085-verify-resumes-execute-session.md)).

**MUST NOT:** reuse phase rows; resume on RetryFresh; fail the cycle solely because Cursor resume failed when full-prompt fallback is viable.

## Recovery hint taxonomy

| `RecoveryKind` | Phase | When |
|----------------|-------|------|
| `verify_implementation_fail` | execute | After verify failure → full re-execute |
| `criteria_report_invalid` / `criteria_report_missing` | execute | Invalid/missing self-report (HEAD OK) |
| `process_restart` | execute | In-cycle resume after worker restart |
| `operator_retry_resume` | execute | Cross-cycle RetryResume |
| `verify_infra_retry` / `verify_feedback_carry` | verify | Verify-only / repeat verify |

## Consequences

### Positive

- Smaller prompts on retries; better alignment with interactive Cursor CLI usage.
- Observable via `cursor_resume_mode`, `recovery_hint_kind`, `recovery_hint_bytes` logs.

### Negative / trade-offs

- Cursor ties sessions to `--workspace`; repo_root changes deny resume.
- DB checkpoint / full prompt remain authoritative when CLI session is unavailable.

## Alternatives considered

| Alternative | Reason rejected |
|-------------|-----------------|
| Always fresh (status quo) | Wastes context; poor retry UX |
| DB-only checkpoint without CLI resume | Does not reduce Cursor token re-send |

## Supersedes

ADR-0006 bullet: `session_id` post-hoc audit only — superseded for execute and verify retry paths by this ADR.
