# ADR-0093: MCP commit register

**Date:** 2026-07-31
**Status:** Accepted
**Deciders:** Engineering
**Supersedes:** Ingest-source decision of [ADR-0032](./ADR-0032-agent-claimed-commit-index.md)

## Context

ADR-0032 indexed execute commits from agent-declared `commits[]` in
`criteria-report.json`. Agents transcribed 40-character SHAs by hand. When the
model fabricated a full SHA that only shared a short prefix with the real
object, execute terminated with `execute_invalid_commit` even though the
worktree commit was correct.

Staff bar: the harness must never require the model to copy object ids. SHAs
enter the ledger only from `git rev-parse` after a commit Hamix itself created.

## Decision

1. **MCP tool `hamix.commit`** — commits the **current index only** (no staging,
   no `-a`, no `--amend`). Agents stage via Shell `git add`.
2. **Commit register** — each successful tool call appends the full HEAD SHA to
   `ReportDir/<cycle_id>/commit-register.json` (atomic rewrite).
3. **Ingest source** — post-execute reads the register only. Criteria-report
   `commits[]` is removed from the product contract and ignored if present.
4. **Exact set equality (I2)** — let `H` be full SHAs from
   `git rev-list --reverse <cycle_base_sha>..HEAD` and `R` the normalized
   register SHAs. Require `set(R) == set(H)`. `rev-list` is an equality oracle,
   not a fallback indexer.
5. **Hard fail (behavior break vs ADR-0032)** — empty register when the git
   snapshot is not skipped → `execute_missing_commits`. `H \ R` →
   `execute_unregistered_commits`. `R \ H` or unresolvable/corrupt register →
   `execute_invalid_commit`.
6. **Unchanged** — task-wide append-only `task_cycle_commits` ledger;
   `ListCommitsForTask` for verify; non-git (`snap.Skipped`) skips ingest.

## Consequences

### Positive

- Eliminates hallucinated-SHA execute failures for correctly committed work.
- Single authoritative write path for new SHAs.
- Shell-only commits fail closed instead of silently omitting ledger rows.

### Negative / Trade-offs

- Agents must use `hamix.commit` (prompt + I2 enforcement); Shell `git commit`
  alone fails execute.
- Empty-claims soft success from ADR-0032 is gone — every git execute visit
  needs ≥1 registered commit.
- Does not cryptographically prevent Shell commit + forged register of a *true*
  SHA (ledger would still be accurate).

## Alternatives Considered

| Alternative | Reason Rejected |
|-------------|-----------------|
| Keep claim `commits[]` + early validate | Still requires model transcription |
| Rev-list as primary indexer | Two sources of truth; bypasses MCP contract |
| Stage inside MCP | Agents already stage via Shell; keep tool minimal |

## See also

- [docs/domain/cycle-commits.md](../domain/cycle-commits.md)
- [docs/domain/agent-mcp.md](../domain/agent-mcp.md)
