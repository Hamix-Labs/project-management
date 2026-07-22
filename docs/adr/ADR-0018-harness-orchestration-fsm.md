# ADR-0018: Harness Orchestration State Machine

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-18
**Status:** Accepted
**Deciders:** Engineering

## Context

After ADR-0017 split harness into `internal/*` domain packages, verify retry semantics still lived as inline conditionals in `runCycleLoopVerify`. Contributors extending cycle behavior must read imperative control flow across `cycle_loop.go` rather than a named transition table.

Industry outer-harness patterns (LangGraph FSM, SWE-AF durable steps) separate **pure decisions** from **effect application** (store writes, runner invocation, git subprocesses).

## Decision

Introduce `pkgs/agents/harness/internal/orchestration` as a **pure Decide** package (event/facts → effects):

| Type / function | Role |
|------|------|
| `VerifyResult` | Classify one verify pipeline outcome (pass, retryable fail, tamper) |
| `VerifyEffects` | Retry loop, terminal failure, or tamper flags |
| `DecideVerifyRetry` / `DecideVerifyRetryWithValidity` | Map `(attempt, maxRetries, result[, executeStillValid])` → effects |

The harness **root applies effects**: increment `verifyAttempt` on `processState.verify`, call `terminateCycle`, run `completeChecklistLegacy`, etc. Orchestration imports **domain types only** — no store, runner, or filesystem.

**Live in-memory scratch** for one run is nested `processState` on the harness root (`cycle.go`): verify counters live on `processState.verify`, not in orchestration. Decide functions take **scalar / DTO projections** at the I/O boundary.

Initial scope covers verify retry/tamper decisions wired from `runCycleLoopVerify`. Execute-phase and loop-level finalize/legacy decisions were added in [ADR-0021](ADR-0021-harness-execute-orchestration.md).

### Historical note (deleted types)

An early sketch included `LoopState` (phase + verify attempt bag) and a `VerifyDisabled(enabled bool)` helper in orchestration. Both were **unused and removed** (commit `5bced0a7`, 2026-07-20). Do not revive them: verify-disabled is gated with `!state.verify.verifySnap.Enabled` at the root; legacy completion uses `DecideVerifyDisabledLegacy(checklistErr)` from ADR-0021. A full `Decide(LoopState, Event)` graph remains **Track C deferred** (ADR-0021) — not a type-alias cleanup.

## Consequences

### Positive

- Verify retry budget is table-tested without harness/store setup.
- New terminal paths add rows to `DecideVerifyRetry` rather than nested `if` chains.
- Clear seam for future cycle timeline / lease effects at machine boundaries.

### Negative / Trade-offs

- Execute transitions originally lagged verify (split brain) — largely closed by ADR-0021 DecideExecute / finalize / legacy.
- Intentional DTO projection between `processState` and orchestration inputs (`ExecutePostRunInput`, `ClassifyInput`, etc.); do not merge `processState` into orchestration (would break leaf purity).

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Full graph in orchestration day one | Too large for Track B; verify retry is highest-risk branch |
| Keep all logic in cycle_loop | Already at reviewability limits; no testable contract |
| Public `harness/orchestration` package | No external importers; `internal/` enforces boundary |

## Related

- [ADR-0017](ADR-0017-harness-internal-domains.md) — internal package layout
- [ADR-0021](ADR-0021-harness-execute-orchestration.md) — execute Decide + Track C deferral
- [docs/domain/harness.md](../domain/harness.md) — durability tiers and cycle lifecycle
