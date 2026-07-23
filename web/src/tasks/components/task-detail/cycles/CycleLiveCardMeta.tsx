import { runnerLabel } from "@/tasks/cycleDisplay/cyclesViewModel";
import type { CycleMeta } from "@/types/cycle";

type CycleLiveCardMetaProps = {
  attemptSeq: number;
  lineage: string | null;
  cycleMeta: CycleMeta;
  startedLabel: string;
};

/**
 * Muted middot footer: Attempt · runner · model · Started …
 * No runtime pill — plain text so status color stays in the head.
 */
export function CycleLiveCardMeta({
  attemptSeq,
  lineage,
  cycleMeta,
  startedLabel,
}: CycleLiveCardMetaProps) {
  const runner = runnerLabel(cycleMeta.runner);
  const model =
    runner === "unknown runner"
      ? null
      : cycleMeta.cursor_model_effective.trim() || "default model";

  return (
    <div className="task-cycle-ticker-meta">
      <span className="task-cycle-ticker-attempt">
        Attempt #{attemptSeq}
        {lineage ? (
          <span className="task-cycle-lineage muted"> · {lineage}</span>
        ) : null}
      </span>
      <span className="task-cycle-ticker-meta-sep" aria-hidden="true">
        ·
      </span>
      <span data-testid="task-cycle-ticker-runner">
        <span className="task-cycle-ticker-runner-label">{runner}</span>
        {model ? (
          <>
            <span className="task-cycle-ticker-meta-sep" aria-hidden="true">
              {" "}
              ·{" "}
            </span>
            <span className="task-cycle-ticker-model-label">{model}</span>
          </>
        ) : null}
      </span>
      <span className="task-cycle-ticker-meta-sep" aria-hidden="true">
        ·
      </span>
      <span
        className="task-cycle-ticker-elapsed"
        data-testid="task-cycle-ticker-elapsed"
      >
        Started {startedLabel} ago
      </span>
    </div>
  );
}
