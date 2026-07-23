# ADR-0083: Task token usage accounting

**Date:** 2026-07-23  
**Status:** Accepted  
**Deciders:** Engineering (task cycles / harness)

## Context

Long resume loops need operators to see how many tokens a task has consumed (execute vs verify) and how much each attempt contributes to that total. Cursor's headless CLI already emits aggregate `usage` on the terminal `result` event. Hamix already persists that object under execute phase `details_json.usage`, but verify overwrites phase details with a verification snapshot and discards the runner `Result`, so verify usage is lost. The IDE “context composition” breakdown (system / rules / tools / conversation) is not available from CLI `usage`.

## Decision

1. **Canonical persistence** — Keep Cursor's camelCase `usage` object on phase `details_json` for both execute and verify. No new DB columns in v1.
2. **Domain type** — `taskcycles/domain.TokenUsage` owns parse / `Consumed` / `Add` / `Known`. **Consumed** = `totalTokens` when present and > 0; otherwise sum of `inputTokens + outputTokens + cacheReadTokens + cacheWriteTokens`.
3. **Verify** — Capture runner `Result` usage from Cursor verify runs (sum across resume retries) and shallow-merge into `EncodePhaseDetails` without dropping `run_correlation_id` (ADR-0030).
4. **Aggregation** — Server-side helpers list/sum usage from phase rows by task, cycle, and phase kind. The SPA must not parse opaque `details_json` for totals.
5. **Missing usage** — Absent or unparseable usage is omitted from sums (`Known() == false`), not coerced to zero.

## Consequences

### Positive

- Stable accounting rules shared by store, API, and UI.
- Historical execute usage already on disk remains readable.
- Fits existing `details_json` / CompletePhase merge patterns.

### Negative / Trade-offs

- Historical verify phases lack usage until re-run.
- `Result.Details` byte caps can still drop `usage` on oversized payloads.
- Aggregate usage overestimates true context-window fill for multi-step turns (Cursor limitation).

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Cursor SDK migration | Wrong stack for Hamix Go harness; same aggregate signal |
| Dedicated token columns / table | Premature; `details_json` already holds execute usage |
| Client-only parse of phase details | Fragile; verify currently lacks usage; duplicates rules |
| CLI `statusLine` / user config | Interactive TTY only; not for desktop app users |
| IDE composition buckets | Not exposed by headless CLI `usage` |
