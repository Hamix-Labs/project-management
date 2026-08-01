# ADR-0021: Harness Execute Orchestration (Decide vs Apply)

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-19
**Status:** Accepted (amended 2026-08-01 — ExecuteVisitPolicy)
**Deciders:** Engineering

## Context

[ADR-0018](ADR-0018-harness-orchestration-fsm.md) introduced `internal/orchestration` for verify retry/tamper decisions (`DecideVerifyRetry`). Execute post-run policy (runner error classification, operator cancel overlay, commit ingest failure) remained imperative in `cycle_loop.go` and `cycle.go`. Verify-disabled legacy and finalize downgrade paths were also inline.

Contributors extending cycle behavior must read two styles: table-driven verify retry vs nested conditionals for execute and loop terminal paths. Terminal writes (`terminateCycle`, `transitionTask`) were duplicated across execute, verify, and finalize branches.

## Decision

Extend `pkgs/agents/harness/internal/orchestration` with execute and loop-level **Decide** functions; consolidate **Apply** in harness root [`cycle_effects.go`](../../pkgs/agents/harness/cycle_effects.go).

| Function | Role |
|----------|------|
| `DecideExecutePostRun` | Runner outcome + cancel + commit ingest → `ExecuteEffects` (run-kind-blind) |
| `ResolveExecuteVisitPolicy` | `run_kind` + skip-claim → `CommitIngestMode` + `PostExecutePath` |
| `DecideVerifyRetry` | Unchanged (verify retry/tamper) |
| `DecideVerifyDisabledLegacy` | Legacy checklist completion err → terminal |
| `DecideFinalizeSuccess` | Completion ledger err → downgrade to failed |

**Boundary rule (unchanged):** orchestration imports `domain` only. Harness maps `runner.Result`/`error` → `ExecuteRunnerOutcome` before Decide. Git ingest takes `CommitIngestMode` only (no knowledge of `open_pr`).

**Effect applier** owns store ordering: `CompletePhase` before `TerminateCycle`; publish/metrics after successful writes. Cycle loop branches finalize on `PostExecutePath` (`ClaimAcceptance` / `ReviewSkipClaims` / `OpenPR`).

**Track C deferred:** a unified event-graph `Decide(...)` over the full cycle (including recovery) is still out of scope. The historical `LoopState` type was never wired and was deleted as unused (`5bced0a7`); do not revive it as a unify vehicle. Shutdown/panic recovery stays imperative in `recovery.go`. Live counters remain on harness-root `processState` (see ADR-0018).

## Consequences

### Positive

- Execute policy is table-tested without harness/store setup.
- Loop policy (execute, verify retry, verify-disabled, finalize) shares one Decide vs Apply model.
- Terminal paths converge on `applyExecuteEffects` / `applyVerifyEffects` / `applyFinalizeEffects`.

### Negative / Trade-offs

- Intentional DTO projection between `processState` and orchestration inputs (`ExecutePostRunInput`, etc.); keep orchestration free of harness scratch bags.
- `funclogmeasure` allowlist updates when symbols move.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Import `runner` in orchestration | Breaks leaf purity established in ADR-0018 |
| Execute-only (skip loop-level Decide) | Leaves verify-disabled/finalize imperative |
| Full event graph day one | ADR-0018 rejected; high regression risk |
| Effect applier in orchestration | Would require store/runner imports |

## Related

- [ADR-0017](ADR-0017-harness-internal-domains.md) — internal package layout
- [ADR-0018](ADR-0018-harness-orchestration-fsm.md) — verify retry machine (Track B start)
- [docs/domain/harness.md](../domain/harness.md) — cycle lifecycle reference
