# ADR-0003: Verify component upgrade

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-05
**Status:** Partially superseded — adversarial `VerifyRunner` / `verify_runner_name` replaced by [ADR-0084](./ADR-0084-executor-owned-verify.md) (executor-owned verify). Integrity checks, locked passes, and verdict observability from this ADR remain in force.
**Deciders:** Backend / agents-worker maintainers

## Context

The V1 verify pass had three structural gaps relative to what
`docs/data-model.md` already promised:

1. **`verifier_kind=verify_agent` was not adversarial.** The worker used
   the same `runner.Runner` instance for execute and verify, even when
   the operator set `app_settings.verify_runner_name`. The setting was
   logged-and-ignored (`pkgs/agents/worker/verification.go::loadVerificationSnapshot`).
2. **No tampering safety net.** Docs spoke of "a separate git worktree"
   for the verify pass, but execute leaves uncommitted changes in
   `RepoRoot`; a fresh worktree at HEAD would be empty and the verifier
   would lose the ability to inspect actual file contents. There was no
   defense against a misbehaving verifier modifying source.
3. **No retry efficiency.** On verify failure, the next attempt re-asked
   the agent to re-prove every criterion — including ones the verifier
   had already accepted — wasting tokens and risking flaky re-evaluation.
4. **Unmeasurable.** No metrics distinguished `agent_self`/`failed`
   (disagreement) from `verify_agent`/`failed` (verifier rejection), no
   verify-phase duration histogram, and no retries-per-cycle distribution.
5. **Opaque terminal failures.** The cycle's `terminate_reason` was the
   bare string `verification_failed`, forcing the SPA to round-trip to
   the checklist API to discover which criterion actually failed.

## Decision

Three sequential, single-concern PRs:

### PR1 — Adversarial runner + post-verify integrity check + verify observability

- `worker.Options.VerifyRunner runner.Runner` is the runner the verify
  pass uses. The supervisor (`cmd/taskapi/run_agentworker.go`) builds
  and probes it from `app_settings.verify_runner_name`; build/probe
  failure demotes verify to reuse the execute runner with a loud warn,
  not a worker-blocking error.
- **No git worktree.** The verifier runs in the execute working dir
  (where the uncommitted changes are) so it can inspect file contents
  directly. Tampering safety lives in a pre/post integrity check:
  capture `git status --porcelain -z` plus `git rev-parse HEAD` before
  `StartPhase(verify)`; re-capture after `CompletePhase(verify)`. Any
  HEAD movement, any post-snapshot error, or any working-tree change
  outside `.t2a/<cycleID>/verify-report.json` terminates the cycle as
  `verify_tampered`. Non-git working dirs degrade to a no-op.
- **Fail-safe under uncertainty.** A failed pre-snapshot is treated as
  tampered. A safety property cannot be defeated by "the check threw
  an exception."
- The verify runner's `OnProgress` callback is wired through to the
  cycle stream under the verify phase's `phase_seq`, so the SPA's P3
  panel shows live verify activity instead of being silently empty.

### PR2 — Per-criterion progress across retries (atomic-decision-preserving)

- `processState.previouslyPassed map[string]criterionVerdict` carries
  passes across retries in memory.
- `injectCriteria` renders an `## Already verified (do not re-do)`
  header and excludes locked criteria from the active checklist.
  `parseCriteriaReport`'s expected-IDs set narrows to the unfinished
  items, so a retry report that legitimately omits locked IDs no
  longer parse-fails.
- `runVerifyChecks` short-circuits any criterion in `previouslyPassed`
  to its original verdict (verifier kind + reasoning preserved) before
  any deterministic check or LLM verify runs.
- **Atomic-decision contract preserved verbatim.** On terminal-success,
  `applyVerifiedCompletions` sees the union of all retry attempts and
  writes every row in one transaction. On terminal failure, NOTHING is
  committed — `previouslyPassed` is in-memory only.

### PR3 — Observability + sharper terminate reason

- Three metrics on `worker.RunMetrics`, registered through the existing
  `cmd/taskapi/RegisterAgentWorkerMetrics` seam:
  - `t2a_verify_verdict_total{verifier_kind, verdict}` — one counter
    that doubles as the disagreement signal. Disagreement is the
    derived query `verifier_kind="agent_self",verdict="failed"`. No
    separate disagreement counter.
  - `t2a_verify_phase_duration_seconds` — wall-clock from
    `StartPhase(verify)` to `CompletePhase(verify)`.
  - `t2a_verify_retries_per_cycle` — observed once per terminate,
    value = `state.verifyAttempt`.
- Terminal verify failure emits
  `verification_failed:<id1>,<id2>,…` (sorted, deduped failing IDs)
  instead of the bare `verification_failed`. The `verification_failed`
  prefix is contract-stable: SPA and other clients MUST use prefix
  matching. Reason column stays ≤ 256 chars; long lists truncate with
  a trailing `…` while keeping the prefix intact.

## Consequences

### Positive

- The docs' "adversarial verify phase" claim is now structurally true
  whenever an operator sets `verify_runner_name`.
- Source mutations during verify are detectable and terminal — the
  trust property is enforceable, not aspirational.
- Retry attempts only re-do the unfinished items, cutting token cost
  proportionally to verifier accuracy on the first pass.
- The verify pass's behavior is measurable (verdict rate, latency,
  retries) so future tuning is data-driven.
- Cycle list views can render the failing criterion IDs without an
  extra round-trip.

### Negative / Trade-offs

- The integrity check requires a working `git` binary in the execute
  working dir for any safety enforcement. Non-git working dirs (test
  fixtures, spike work) bypass the check; this is logged once at
  startup but is otherwise silent per cycle.
- `previouslyPassed` lives in memory only. A worker restart re-runs
  the cycle from scratch via the orphan sweep (`SweepOrphanRunningCycles`),
  re-spending verifier tokens on already-proven criteria once. Persisting
  partial completions across restarts was rejected as out of scope —
  the orphan sweep contract is sufficient at current scale.
- The `terminate_reason` enrichment is technically additive but does
  change the wire shape of an existing field. The contract requires
  prefix-matching on the consumer side, which the SPA does not
  currently rely on (it does not render `terminate_reason` directly);
  the prefix-stability test in
  `pkgs/agents/worker/verification_reason_test.go` is the regression
  boundary if any future change breaks the prefix.

## Alternatives Considered

| Alternative | Reason Rejected |
|---|---|
| Run verify in a fresh `git worktree add HEAD` | Empty worktree because execute leaves changes uncommitted; verifier loses access to actual file contents. The integrity check delivers the same property at lower complexity. |
| Auto-commit execute changes per cycle, then real worktree isolation | Reshapes the operator workflow (would surface intermediate commits in user repos). Separate decision; not on the verify-component path. |
| Persist `previouslyPassed` to disk across restarts | Premature for current scale. The orphan sweep already re-runs cycles from scratch on restart; redoing verifier work for one criterion is bounded and observable in `t2a_verify_verdict_total`. |
| Separate `t2a_verify_disagreement_total` counter | Derivable from the existing verdict counter as `verifier_kind="agent_self",verdict="failed"`. Two metrics for one fact is a maintenance tax. |
| Multi-judge verify ensemble / self-consistency voting | Out of scope. The adversarial separation between execute and verify runners is the leverage point; multi-judge is a future PR if needed. |
| Per-criterion verify timeout setting | Operator surface bloat. The existing `check_command_timeout_seconds` covers deterministic checks; the LLM verify call is bounded by `MaxRunDurationSeconds`. |
| Evidence-citation regex enforcement in the parser | Verifier-prompt tuning concern, not a structural fix. Tightening the prompt is zero new code paths and zero new settings. |
