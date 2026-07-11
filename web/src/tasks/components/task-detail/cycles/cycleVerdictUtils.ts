import type {
  CycleCommandRun,
  CycleCriteriaReport,
  CycleVerifyReport,
  VerifierKind,
} from "@/types/cycle";

export type VerdictRow = {
  criterionId: string;
  verified: boolean;
  verifierKind: VerifierKind | "";
  reasoning: string;
  evidence: string;
};

export type AttemptGroup = {
  attemptSeq: number;
  rows: VerdictRow[];
};

/**
 * Joins criteria-side and verify-side rows by `(attempt_seq, criterion_id)`.
 * The verify row wins when both exist; the criteria row provides
 * `evidence` text used as the fallback display when no verify row was
 * recorded (e.g. the verifier never ran for that criterion).
 */
export function groupVerdictsByAttempt(
  criteria: ReadonlyArray<CycleCriteriaReport>,
  verify: ReadonlyArray<CycleVerifyReport>,
): AttemptGroup[] {
  const byAttempt = new Map<number, Map<string, VerdictRow>>();
  const ensure = (attemptSeq: number): Map<string, VerdictRow> => {
    let m = byAttempt.get(attemptSeq);
    if (!m) {
      m = new Map();
      byAttempt.set(attemptSeq, m);
    }
    return m;
  };
  for (const c of criteria) {
    const m = ensure(c.attempt_seq);
    m.set(c.criterion_id, {
      criterionId: c.criterion_id,
      verified: false,
      verifierKind: "",
      reasoning: "",
      evidence: c.evidence,
    });
  }
  for (const v of verify) {
    const m = ensure(v.attempt_seq);
    const existing = m.get(v.criterion_id);
    m.set(v.criterion_id, {
      criterionId: v.criterion_id,
      verified: v.verified,
      verifierKind: v.verifier_kind,
      reasoning: v.reasoning,
      evidence: existing?.evidence ?? "",
    });
  }
  const groups: AttemptGroup[] = [];
  for (const [attemptSeq, m] of byAttempt) {
    const rows = Array.from(m.values()).sort((a, b) =>
      a.criterionId.localeCompare(b.criterionId),
    );
    groups.push({ attemptSeq, rows });
  }
  groups.sort((a, b) => a.attemptSeq - b.attemptSeq);
  return groups;
}

export function commandRunsForAttempt(
  runs: ReadonlyArray<CycleCommandRun>,
  attemptSeq: number,
): CycleCommandRun[] {
  return runs.filter((r) => r.attempt_seq === attemptSeq);
}

export function commandRunAttemptSeqs(
  runs: ReadonlyArray<CycleCommandRun>,
): number[] {
  const seen = new Set<number>();
  for (const r of runs) {
    seen.add(r.attempt_seq);
  }
  return Array.from(seen).sort((a, b) => a - b);
}
