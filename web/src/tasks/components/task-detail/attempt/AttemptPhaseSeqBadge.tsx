export function PhaseSeqBadge({ seq }: { seq: number }) {
  return (
    <span className="task-attempt-phase-seq" aria-label={`Phase ${seq}`}>
      PHASE {seq}
    </span>
  );
}
