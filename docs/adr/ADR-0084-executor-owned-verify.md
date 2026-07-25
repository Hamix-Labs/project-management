# ADR-0084: Executor-owned verify phase

**Date:** 2026-07-25
**Status:** Accepted
**Deciders:** Backend / agents-worker maintainers

## Context

ADR-0003 required an optional adversarial verify runner (`verify_runner_name` /
`Options.VerifyRunner`) so `verifier_kind=verify_agent` meant a second model
judged execute’s work. In practice operators often left verify on the execute
runner, Cursor resume already continues the same session into `PhaseVerify`,
and product copy claiming an “independent verification agent” overstated the
control plane.

Acceptance is already a **second phase** (gate → worker shell evidence → LLM
verdict → git integrity), not a second identity. Keeping a configurable second
runner that may be ignored or demoted is worse than an honest self-grading
design with worker-owned evidence.

## Decision

1. **Same agent** — The LLM that writes `verify-report.json` is the execute
   agent. `PhaseVerify` always invokes the worker’s execute `runner.Runner`.
   Prefer Cursor `--resume` of the execute session when policy allows
   ([ADR-0085](./ADR-0085-verify-resumes-execute-session.md)). Optional
   `verify_model` may pass a different `--model` on that resume; it does
   **not** select a different runner. When resume is disabled, a full verify
   prompt with criteria, evidence, command previews, and diff is used.
2. **Wire kind** — Successful LLM verdicts persist
   `verified_by=execute_agent` (replaces `verify_agent`). `agent_self` remains
   failure-only for unclaimed criteria.
3. **Prompts** — Verify and execute criteria copy describe continuation of the
   implementer (“You implemented this task; now verify…”), not a new persona.
4. **Supersedes ADR-0003** only for adversarial `VerifyRunner` /
   `verify_runner_name` as a trust requirement. Settings deletion follows in a
   subsequent PR; this decision requires the harness to **stop selecting** a
   separate verify runner/model immediately.
5. **Retains** from ADR-0003: pre/post git integrity (`verify_tampered`),
   in-memory locked passes, verdict metrics, enriched
   `verification_failed:<ids>` terminate reasons, structured verify commands
   as evidence (ADR-0012 — exit 0 does not auto-complete).

## Consequences

### Positive

- Product language matches the control plane (self-grading + worker evidence).
- One runner to probe, configure, and resume; fewer demotion paths.
- Clear `execute_agent` label for operators reviewing completions.

### Negative / Trade-offs

- Self-grading bias is accepted. Mitigations: worker-owned command evidence in
  the verify prompt, minimum reasoning length when `verified=true`, git
  integrity during the verify phase, and honest documentation.
- Removing adversarial runner settings is a deliberate API/settings break
  (follow-up PR).

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Fold verify into a single execute `Run` | Worker cannot inject shell evidence between implement and judge without a phase barrier. |
| Trust `claimed_done` alone | Removes the judgment pass. |
| Auto-pass on shell exit 0 | Contradicts ADR-0012. |
| Keep `verify_agent` wire value | Misnames the role; no production rows to preserve. |
| Keep configurable verify runner “for later” | Dead operator surface. |

## See also

- [ADR-0003](./ADR-0003-verify-component-upgrade.md) — integrity, locked passes (retained); adversarial runner (superseded here)
- [ADR-0012](./ADR-0012-structured-verify-commands.md) — shell evidence, not auto-pass
- [ADR-0028](./ADR-0028-in-cycle-verify-only-retry.md) — verify-only retry unchanged
