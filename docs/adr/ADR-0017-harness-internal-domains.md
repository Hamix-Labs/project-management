# ADR-0017: Harness Internal Domain Packages

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-18
**Status:** Accepted
**Deciders:** Engineering

## Context

After ADR-0005 extracted cycle choreography into `pkgs/agents/harness`, the package grew to ~26 production files (~4,900 lines) in a single flat `package harness`. Several files exceed the repo file-size yellow bar (`verification.go`, `git_commits.go`, `cycle_loop.go`). Domains (verify, git, resume, prompt, reports) are distinguishable by filename but not enforced by the compiler.

Contributors extending harness behavior must grep a large package. Cross-domain coupling is easy to introduce. Tests for one domain often require constructing a full `Harness`.

## Decision

Split harness implementation into **`pkgs/agents/harness/internal/<domain>/`** subpackages, importable only from `harness` and sibling `internal/*` (Go `internal/` rule).

| Subpackage | Responsibility |
|------------|----------------|
| `internal/reports` | Report-dir paths, parse/validate side-channel JSON, sentinel errors |
| `internal/git` | Commit observe/admit, fresh-retry reset, verify integrity snapshots (`GitRepo` port at I/O edge) |
| `internal/prompt` | Execute/verify prompt assembly from DTOs (no store writes) |
| `internal/verify` | Verify pipeline: gate, commands, integrity, LLM, persist |
| `internal/resume` | Checkpoint reconstruction, operator retry routing, continuation bundles |
| `internal/cursorresume` | Pure Cursor CLI `--resume` Decide + recovery-kind helpers (ADR-0031); distinct from harness checkpoint resume |
| `internal/execute` | Execute-phase I/O pipeline: phase start ports, git snapshot, runner invoke ports, commit ingest, post-run fact mapping |
| `internal/orchestration` | Pure cycle state machine: event → effects (Track B) |

**Public API unchanged:** `harness.Harness`, `New`, `Run`, `Resume`, `RunWithRetry`, `CancelCurrentRun`, seam interfaces, and report sentinel errors (re-exported from root where needed). `pkgs/agents/worker` remains the only production importer of `harness`.

**Orchestration stays at root** for Track A: `Harness`, `processState`, cycle start/terminate, recovery, metrics integrate domain packages.

### Dependency rules

```
harness (root)
  → internal/resume, internal/verify, internal/prompt, internal/git, internal/reports, internal/orchestration, internal/cursorresume, internal/execute

internal/resume → internal/prompt, internal/git, internal/reports
internal/verify → internal/git, internal/reports, runner, adapterkit
internal/execute → internal/git, internal/reports, internal/orchestration, contract, runner
internal/prompt → internal/git, internal/reports
internal/cursorresume → internal/prompt, domain (pure Decide; no store/runner)
internal/git    → store, domain, stdlib/git CLI
internal/reports → domain, stdlib (leaf)
internal/orchestration → domain only (leaf for Track B pure machine)
```

**Forbidden:** any `internal/*` importing `package harness` (root). **Forbidden:** `internal/verify` ↔ `internal/resume` mutual imports. **Forbidden:** `internal/execute` ↔ `internal/verify` mutual imports (criteria mirror/probe stay at root or use hooks). **Forbidden:** folding Cursor CLI policy into `internal/resume` (different layer — ADR-0031).

### Migration

Track A (phases 0–6): **move-only** — behavior, report formats, SSE, and metrics unchanged. One commit (+ push to `main`) per phase.

Track B–C (follow-on ADRs): orchestrator/effects split, durability tier docs, versioned report schema.

## Consequences

### Positive

- Domain boundaries are compiler-enforced; reviews scope to one subpackage.
- Unit tests can target `internal/verify`, `internal/reports`, etc. without full harness setup.
- File-size bar achievable per domain; orchestration extension has a named home (`internal/orchestration`).

### Negative / Trade-offs

- DTO mapping at package boundaries (no `*Harness` in `internal/*`).
- `funclogmeasure` allowlist and docs links must track moved symbols.
- Temporary churn in import paths during phased migration.

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Public subpackages (`harness/verify`) | No external importers today; `internal/` prevents accidental coupling |
| Single flat package + lint rules | Not enforceable by Go; already at scale limits |
| Harness interface + registry | No second orchestration implementation (same as ADR-0005) |
| Merge domains into `pkgs/agents/` siblings | Harness-specific; breaks cohesion with cycle orchestration |

## Related

- [ADR-0005](ADR-0005-extract-agent-harness.md) — harness extraction from worker
- [ADR-0018](ADR-0018-harness-orchestration-fsm.md) — orchestrator/effects split (Track B)
- [docs/domain/harness.md](../domain/harness.md) — behavioral reference
