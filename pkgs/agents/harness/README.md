# pkgs/agents/harness

Cycle choreography around `runner.Run`. The worker (`pkgs/agents/worker`) handles queue admission; the harness drives one task from `StartCycle` through terminal `TerminateCycle`, or resumes an open cycle after process restart.

**Behavioral reference:** [docs/domain/harness.md](../../docs/domain/harness.md). See also [docs/architecture.md](../../docs/architecture.md), [docs/domain/execute-agent.md](../../docs/domain/execute-agent.md), [docs/domain/verify-agent.md](../../docs/domain/verify-agent.md), [ADR-0005](../../docs/adr/ADR-0005-extract-agent-harness.md), [ADR-0006](../../docs/adr/ADR-0006-phase-boundary-resume.md), [ADR-0017](../../docs/adr/ADR-0017-harness-internal-domains.md), [ADR-0018](../../docs/adr/ADR-0018-harness-orchestration-fsm.md), [ADR-0021](../../docs/adr/ADR-0021-harness-execute-orchestration.md), and [ADR-0028](../../docs/adr/ADR-0028-in-cycle-verify-only-retry.md).

## Internal layout

Domain logic lives under `internal/` (importable only from `harness` and sibling `internal/*`):

| Package | Role |
|---------|------|
| [`internal/reports/`](internal/reports/) | Side-channel JSON paths, parse/validate, `schema_version` |
| [`internal/git/`](internal/git/) | Commits, reset, verify integrity (`GitRepo` port) |
| [`internal/prompt/`](internal/prompt/) | Execute/verify prompt assembly |
| [`internal/verify/`](internal/verify/) | Verification pipeline stages |
| [`internal/resume/`](internal/resume/) | Checkpoint load, retry routing, continuation bundles |
| [`internal/cursorresume/`](internal/cursorresume/) | Pure ADR-0031 Cursor CLI `--resume` Decide + recovery-kind helpers |
| [`internal/execute/`](internal/execute/) | Execute-phase I/O pipeline (git snap, runner ports, commit ingest, post-run facts) |
| `internal/orchestration/` | Pure cycle Decide functions (`DecideVerifyRetry`, `DecideVerifyRetryWithValidity`, `ClassifyVerifyRetryMode`, `DecideExecutePostRun`, loop-level finalize/legacy) |

Root `harness` owns `Harness`, cycle entrypoints, effect application (`cycle_effects.go`), recovery, and metrics.

## File map (root package)

| File | Responsibility |
|------|----------------|
| `harness.go` | `Harness`, `New`, `Options`, `CancelCurrentRun`, SSE notifiers, metrics interface |
| `cycle.go` | `Run` entry, `processState`, cycle start/terminate |
| `cycle_runner.go` | Execute phase start, runner invoke, complete execute phase |
| `cycle_loop.go` | Shared execute/verify loop coordinator; I/O then orchestration Decide |
| `execution.go` | Thin wiring to `internal/execute` (`executeSvc`, phase ports) |
| `cursor_resume.go` | ADR-0031 Cursor `--resume` I/O planners ([cursor-session-resume.md](../../docs/domain/cursor-session-resume.md)) |
| `cursor_resume_resolve.go` | Session lookup + recovery context assembly; pure Decide in [`internal/cursorresume`](internal/cursorresume/) |
| `cursor_resume_decide.go` | Root aliases for `internal/cursorresume` Decide types |
| `cycle_effects.go` | Applies orchestration effects (store writes, publish, metrics) |
| `cycle_execute_adapter.go` | Thin re-exports of execute adapter helpers used by effect apply |
| `verify_retry_eligibility.go` | Post-execute anchors + `gatherRetryClassifyInput` (ADR-0028) |
| `cycle_verify_only_test.go` | Integration tests for in-cycle verify-only retry (EC-xx) |
| `resume.go` | `Resume` — continue an open cycle after `process_restart` finalization |
| `retry_run.go` | `RunWithRetry` — operator fresh/resume retry modes |
| `verification.go` | Thin delegators to `internal/verify` |
| `git_alias.go` | Thin delegators to `internal/git` |
| `resume_alias.go` | Thin delegators to `internal/resume` |
| `reports_alias.go` | Re-exports report sentinel errors |
| `prompt_helpers.go` | Checklist/continuation helpers for prompt assembly |
| `execute_criteria_mirror.go` | Best-effort criteria mirror after execute |
| `meta.go` | Cycle `MetaJSON` and phase `details_json` normalization |
| `metrics.go` | `RunMetrics` seam and observation helpers |
| `effective_model.go` | Model resolution for execute/verify runners |
| `recovery.go` | Panic, shutdown, and best-effort cycle closeout paths |
| `invariant_test.go` | Durability/orchestration contract tests |

## Public entry points

```go
h := harness.New(store, runner, harness.Options{...})
h.Run(ctx, task)    // task must be StatusReady → worker sets running and starts new cycle
h.Resume(ctx, task, cycle) // task StatusRunning, cycle StatusRunning — same attempt continues
```

Callers outside tests typically use `worker.NewWorker`, which constructs the harness internally and chooses `Run` vs `Resume` at admission.

## Testing

Three tiers (pure, contract, wrapper) are documented in [docs/domain/harness-testing.md](../../docs/domain/harness-testing.md).

| Package / file | Role in tests |
|----------------|---------------|
| [`storefake/`](storefake/) | `contract.Store` double (SQLite via `tasktestdb`, isolated per test) |
| [`notifierfake/`](notifierfake/) | Recording cycle/progress notifiers |
| [`metricsfake/`](metricsfake/) | Recording `RunMetrics` for verdict/duration assertions |
| `testhelpers_test.go` | `newHarnessWithFakes`, `runHarness` for `harness_test` package |
| `internal/verify/integration_testhelpers_test.go` | Same pattern for verify integration tests |

Contract-tier tests call `harness.Run` directly with fakes. They do **not** import `internal/tasktestdb` or start `worker.Worker`.

Local: `go test ./pkgs/agents/harness/... -count=1 -timeout 120s` or `.\scripts\check.ps1 -GoOnly` (`--group=harness` in CI).

## Legacy verify / checklist paths (retirement criteria)

"Legacy" names mark **compatibility** paths, not dead code:

| Path | Why retained | Delete when |
| --- | --- | --- |
| `completeChecklistLegacy` / `DecideVerifyDisabledLegacy` / `applyVerifyDisabledLegacyEffects` | Tasks (or settings) with verification disabled still need a deterministic cycle completion that marks checklist items without running the verify agent | Product removes verify-disabled mode **and** a one-time migrate has rewritten historical rows / MetaJSON that encode the old path |
| `taskchecklist` `VerifierLegacy` | Historical completions and store writes that predate typed verifier kinds | No remaining `verified_by = 'legacy'` rows in production DBs **and** create/update APIs reject new legacy writes (already blocked on HTTP create) |

Do not add new call sites to these helpers. Prefer the modern verify loop and typed `VerifierKind` values.

## Checkpoint derivation (resume)

No dedicated checkpoint table. `internal/resume` reconstructs checkpoint from:

- Phase ledger tail → execute vs verify resume branch
- `task_cycle_verify_reports` → locked passes, verify attempt, retry feedback
- Task row → base prompt
- `task_cycle_commits` → worker-indexed SHAs for resume/verify prompts (see [cycle-commits.md](../../docs/domain/cycle-commits.md))

The composed prompt is what the runner sees; `WorkingDir` remains `app_settings.repo_root`.
