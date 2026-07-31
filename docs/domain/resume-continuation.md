# Resume continuation bundle

Cross-cycle **Resume from failure** loads one `ContinuationBundle` from the parent cycle instead of ad-hoc checkpoint fields.

| | |
| --- | --- |
| **Applies to** | `loadContinuationBundle`, resume retry prompt assembly |
| **Audience** | Contributors touching `pkgs/agents/harness` resume paths |
| **Prerequisite** | [retry-resume.md](./retry-resume.md) — operator Resume from failure |
| **Decision record** | [ADR-0093](../adr/ADR-0093-mcp-commit-register.md) (commit register); historical claim index [ADR-0032](../adr/ADR-0032-agent-claimed-commit-index.md) |

## In this article

- [Loader](#loader)
- [Routing](#routing)
- [Prompt order](#prompt-order)
- [See also](#see-also)

## Loader

`loadContinuationBundle(parentCycleID)` classifies the parent outcome, checks sufficiency, routes resume entry, and assembles prompt blocks.

| Field | Purpose |
|-------|---------|
| `entry` | `execute` or `verifyOnly` |
| `failureClass` | runner, executeGate, verify, infrastructure, operator |
| `scopeFiles` | Paths from `git diff --name-only cycle_base..HEAD` |
| `commits` | Task-wide indexed commits (`ListCommitsForTask`) |
| `executeFeedback` | Prior execute failure context |
| `verifyFeedback` | Last verify attempt failures |
| `previouslyPassed` | Locked verify passes |

## Routing

- Parent execute **succeeded**, verify **failed**, task-wide ledger has commits → `verifyOnly` (`skipFirstExecute`)
- Otherwise → `execute` with continuation prompt (scope lock, known commits, additive-only git policy)

## Prompt order

Lineage → failure explanation → scope lock → known commits → execute/runner feedback → verify feedback → original prompt → git policy.

Insufficient data → `retry_checkpoint_failed` (no hollow resume).

## See also

- [retry-resume.md](./retry-resume.md) — operator Resume from failure
- [cycle-commits.md](./cycle-commits.md) — task-wide commit ledger
- [harness.md](./harness.md) — cycle loop and worker boundary
- [ADR-0093](../adr/ADR-0093-mcp-commit-register.md)
- [ADR-0032](../adr/ADR-0032-agent-claimed-commit-index.md)
