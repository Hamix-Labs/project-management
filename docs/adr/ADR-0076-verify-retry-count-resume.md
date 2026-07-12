# ADR-0076: Explicit verify retry count on resume

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

On in-cycle resume, `VerifyAttempt` was derived from the max `attempt_seq` in `task_cycle_verify_reports`. That conflates DB attempt sequence with the harness retry budget counter ([HARNESS_IMPROVEMENTS.md](../HARNESS_IMPROVEMENTS.md) P0 #5).

## Decision

1. Persist `verify_retry_count` (harness `verifyAttempt` before the verify pass) in verify phase `details_json` via `EncodePhaseDetails`.
2. `ReconstructCheckpoint` reads `verify_retry_count` from the latest verify phase when present; otherwise falls back to `max(attempt_seq)` from verify report rows (legacy cycles).

## Consequences

### Positive

- Resume after `process_restart` respects `verify_max_retries` with an explicit counter.

### Negative / Trade-offs

- Cross-cycle operator resume still resets `VerifyAttempt` to 0 per [ADR-0015](./ADR-0015-dual-retry-modes.md).

## See also

- [pkgs/agents/harness/internal/resume/checkpoint.go](../../pkgs/agents/harness/internal/resume/checkpoint.go)
