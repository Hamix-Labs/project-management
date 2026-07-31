# ADR-0091: Stream-idle fail-clean

**Date:** 2026-07-30
**Status:** Accepted
**Deciders:** Harness / agent-worker maintainers

## Context

Cursor execute/verify runs block until `cursor-agent` exits. Production runs can hang after the CLI stops emitting stdout (for example a Shell tool that soft-timeouts into the background and never completes). With `max_run_duration_seconds = 0`, the harness waits indefinitely.

[ADR-0027](ADR-0027-stream-idle-evidence-recovery.md) previously killed on stdout silence and attempted evidence recovery. That stack was removed; only wall-clock `max_run_duration_seconds` remained. Operators still need a liveness kill that is independent of wall-clock, without reintroducing recovery complexity.

## Decision

1. **Two orthogonal budgets** on every `runner.Run`:
   - `max_run_duration_seconds` → wall-clock → `runner.ErrTimeout` → terminate reason `runner_timeout`
   - `stream_idle_stuck_seconds` → stdout line silence after the first line → `runner.ErrStale` → terminate reason `stream_idle`
2. **Liveness** lives in `adapterkit.DefaultStreamExecWithIdle` (tiers: suspicious at `stuck/2`, kill-pending near stuck, cancel with `adapterkit.ErrStreamIdle`).
3. **Grace:** idle detection starts only after the first stdout line.
4. **Fail clean:** on `ErrStale`, fail the phase/cycle/task. Do **not** run evidence-recovery gates.
5. **Default:** `stream_idle_stuck_seconds = 900` (15 minutes). `0` disables. Long silent shells may false-positive; operators raise the knob or set `0`.
6. **Settings:** field under Execute → Runner, beside Max execute duration. Supervisor hot-swaps when the value changes (`InstanceMatchesSettings`).

## Consequences

### Positive

- Hung silent agents free the worker slot without operator cancel
- Audit reason `stream_idle` is distinct from wall-clock timeout
- Configurable / disableable

### Negative / Trade-offs

- Long silent tools can trip the watchdog; mitigated by high default and `0` off
- No automatic salvage of commits/reports after kill (by design)

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Evidence recovery (ADR-0027) | Removed once; over-coupled to git/report gates |
| Collapse into `runner_timeout` | Wrong audit semantics |
| Tool-call-aware idle | Over-engineered for v1 |
