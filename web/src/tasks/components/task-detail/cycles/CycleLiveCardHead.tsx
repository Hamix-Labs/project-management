import {
  cycleStatusFillClass,
  cycleStatusLabel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import type { CycleStatus } from "@/types/cycle";

type CycleLiveCardHeadProps = {
  cycleStatus: CycleStatus;
  /** Phase name shown as bordered mono badge when a phase is in flight. */
  phaseName?: string | null;
  /** Formatted phase elapsed (e.g. from formatDurationSeconds). */
  phaseElapsed?: string | null;
};

/**
 * Live card header: ping + Live + status pill on the left;
 * optional phase badge + elapsed on the right.
 */
export function CycleLiveCardHead({
  cycleStatus,
  phaseName = null,
  phaseElapsed = null,
}: CycleLiveCardHeadProps) {
  const showPhase = Boolean(phaseName);

  return (
    <div className="task-cycle-ticker-head">
      <div className="task-cycle-ticker-head-start">
        <span className="cycle-live-dot-wrap" aria-hidden="true">
          <span className="cycle-live-dot-ping" />
          <span className="cycle-live-dot" />
        </span>
        <span className="task-cycle-ticker-eyebrow">Live</span>
        <span
          className={`cell-pill ${cycleStatusFillClass(cycleStatus)}`}
          data-testid="task-cycle-ticker-status"
        >
          {cycleStatusLabel(cycleStatus)}
        </span>
      </div>
      {showPhase ? (
        <div
          className="task-cycle-ticker-head-end"
          data-testid="task-cycle-ticker-phase"
        >
          <span className="task-cycle-ticker-phase-badge">{phaseName}</span>
          {phaseElapsed ? (
            <span
              className="task-cycle-ticker-focus-elapsed"
              aria-hidden="true"
            >
              {phaseElapsed}
            </span>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
