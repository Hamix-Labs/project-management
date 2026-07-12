# ADR-0074: Verdict mirror degradation visibility

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

When `UpsertVerifyReports` or `UpsertCriteriaReports` fails during verify, the harness logs a warning and continues ([HARNESS_IMPROVEMENTS.md](../HARNESS_IMPROVEMENTS.md) P0 #3). Operators had no durable signal that audit mirrors were stale.

## Decision

1. Set `mirror_degraded: true` in verify phase `details_json` when criteria or verify report upserts fail, or when execute criteria mirror upsert fails before verify.
2. **Do not** change verify pass/fail outcomes on mirror errors (phase B gating deferred).

## Consequences

### Positive

- UI and operators can surface degraded audit mirrors without failing the run.

### Negative / Trade-offs

- Clients must read `mirror_degraded` from phase details; no separate SSE event.

## See also

- [ADR-0075](./ADR-0075-non-blocking-harness-notifier.md)
