import { useState } from "react";
import { errorMessage } from "@/lib/errorMessage";
import { useNow } from "@/shared/useNow";
import {
  cycleStatusLabel,
  cycleStatusFillClass,
  cycleRunnerChipClass,
  formatRunnerModel,
  phaseLabel,
  phaseStatusFillClass,
  phaseStatusLabel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import type { TaskCycle } from "@/types/cycle";
import { useTaskCycle } from "../../../hooks/useTaskCycles";
import { formatCycleLineageLabel } from "../../../cycleDisplay/cycleLineage";
import { CycleRowVerdicts } from "./CycleRowVerdicts";
import {
  formatPhaseDuration,
  formatStartedToEnded,
} from "./cyclePanelUtils";

type CycleHistoryListProps = {
  taskId: string;
  cycles: TaskCycle[];
  runningCycleId: string | null;
  cyclesById: ReadonlyMap<string, TaskCycle>;
};

export function CycleHistoryList({
  taskId,
  cycles,
  runningCycleId,
  cyclesById,
}: CycleHistoryListProps) {
  if (cycles.length === 0) {
    return null;
  }
  return (
    <ol className="task-cycles-list" data-testid="task-cycles-list">
      {cycles.map((cycle) => (
        <CycleRow
          key={cycle.id}
          taskId={taskId}
          cycle={cycle}
          isLiveAbove={cycle.id === runningCycleId}
          cyclesById={cyclesById}
        />
      ))}
    </ol>
  );
}

function CycleRow({
  taskId,
  cycle,
  isLiveAbove,
  cyclesById,
}: {
  taskId: string;
  cycle: TaskCycle;
  isLiveAbove: boolean;
  cyclesById: ReadonlyMap<string, TaskCycle>;
}) {
  const [open, setOpen] = useState(false);
  const lineage = formatCycleLineageLabel(cycle, cyclesById);

  return (
    <li className="task-cycle-row" data-cycle-status={cycle.status}>
      <details
        open={open}
        onToggle={(e) => setOpen((e.currentTarget as HTMLDetailsElement).open)}
      >
        <summary className="task-cycle-row-summary">
          <span
            className={`cell-pill ${cycleStatusFillClass(cycle.status)}`}
            data-testid="task-cycle-row-status"
          >
            {cycleStatusLabel(cycle.status)}
          </span>
          <span className="task-cycle-row-attempt">
            Attempt #{cycle.attempt_seq}
            {lineage ? (
              <span className="task-cycle-lineage muted"> · {lineage}</span>
            ) : null}
          </span>
          <span className="task-cycle-row-when muted">
            {formatStartedToEnded(cycle)}
          </span>
          <span className="task-cycle-row-trigger muted">
            by {cycle.triggered_by}
          </span>
          <span
            className={`cell-pill ${cycleRunnerChipClass()}`}
            data-testid="task-cycle-row-runner"
          >
            {formatRunnerModel(cycle.cycle_meta)}
          </span>
          {isLiveAbove ? (
            <span
              className="task-cycle-row-livehint"
              aria-label="This cycle is shown in the live ticker above"
            >
              ↑ live
            </span>
          ) : null}
          <a
            className="task-cycle-row-attempt-link"
            href={`/tasks/${encodeURIComponent(taskId)}/cycles/${encodeURIComponent(cycle.id)}`}
            onClick={(e) => e.stopPropagation()}
          >
            View run details
          </a>
        </summary>
        {open ? (
          <>
            <CycleRowPhases taskId={taskId} cycleId={cycle.id} />
            <CycleRowVerdicts taskId={taskId} cycleId={cycle.id} />
          </>
        ) : null}
      </details>
    </li>
  );
}

function CycleRowPhases({
  taskId,
  cycleId,
}: {
  taskId: string;
  cycleId: string;
}) {
  const detailQuery = useTaskCycle(taskId, cycleId);
  const phases = detailQuery.data?.phases ?? [];
  const hasRunningPhase = phases.some((phase) => phase.status === "running");
  const now = useNow({ enabled: hasRunningPhase });

  if (detailQuery.isPending) {
    return (
      <p className="task-cycle-row-phases muted" aria-busy="true">
        Loading phases…
      </p>
    );
  }
  if (detailQuery.isError) {
    return (
      <p className="task-cycle-row-phases err" role="alert">
        {errorMessage(detailQuery.error, "Could not load phases.")}
      </p>
    );
  }
  if (phases.length === 0) {
    return (
      <p className="task-cycle-row-phases muted">
        No phases recorded for this cycle.
      </p>
    );
  }
  return (
    <ol className="task-cycle-phase-list" aria-label="Phases for this cycle">
      {phases.map((phase) => (
        <li
          key={phase.id}
          className="task-cycle-phase-item"
          data-phase-status={phase.status}
        >
          <span className="task-cycle-phase-name">{phaseLabel(phase.phase)}</span>
          <span className={`cell-pill ${phaseStatusFillClass(phase.status)}`}>
            {phaseStatusLabel(phase.status)}
          </span>
          <span className="task-cycle-phase-duration muted">
            {formatPhaseDuration(phase, now)}
          </span>
          {phase.summary ? (
            <span className="task-cycle-phase-summary">{phase.summary}</span>
          ) : null}
        </li>
      ))}
    </ol>
  );
}
