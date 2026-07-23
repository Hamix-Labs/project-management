import { useMemo, useState } from "react";
import { errorMessage } from "@/lib/errorMessage";
import { useAppTimezone } from "@/shared/time/appTimezone";
import { useNow } from "@/shared/useNow";
import {
  cycleStatusLabel,
  cycleStatusFillClass,
  phaseLabel,
  phaseStatusFillClass,
  phaseStatusLabel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import type { TaskCycle, TaskTokenUsageAttempt } from "@/types/cycle";
import { useTaskCycle } from "../../../hooks/useTaskCycles";
import { useTaskTokenUsage } from "../../../hooks/useTaskTokenUsage";
import { formatCycleLineageLabel } from "../../../cycleDisplay/cycleLineage";
import {
  formatShareOfTaskPct,
  formatTokenCount,
} from "../../../task-display/formatTokenCount";
import { CycleRowVerdicts } from "./CycleRowVerdicts";
import {
  formatAttemptTiming,
  formatPhaseDuration,
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
  const usageQuery = useTaskTokenUsage(taskId);
  const attemptsByCycleId = useMemo(
    () => indexAttemptsByCycleId(usageQuery.data?.attempts ?? []),
    [usageQuery.data?.attempts],
  );

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
          attemptUsage={attemptsByCycleId.get(cycle.id)}
        />
      ))}
    </ol>
  );
}

function indexAttemptsByCycleId(
  attempts: TaskTokenUsageAttempt[],
): Map<string, TaskTokenUsageAttempt> {
  const out = new Map<string, TaskTokenUsageAttempt>();
  for (const attempt of attempts) {
    out.set(attempt.cycle_id, attempt);
  }
  return out;
}

function CycleRow({
  taskId,
  cycle,
  isLiveAbove,
  cyclesById,
  attemptUsage,
}: {
  taskId: string;
  cycle: TaskCycle;
  isLiveAbove: boolean;
  cyclesById: ReadonlyMap<string, TaskCycle>;
  attemptUsage?: TaskTokenUsageAttempt;
}) {
  const [open, setOpen] = useState(false);
  const tz = useAppTimezone();
  const lineage = formatCycleLineageLabel(cycle, cyclesById);
  const timing = formatAttemptTiming(cycle, tz);
  const tokenSummary = formatCycleTokenSummary(cycle, attemptUsage);

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
          <span
            className="task-cycle-row-when muted"
            aria-label={timing.ariaLabel}
            data-testid="task-cycle-row-when"
          >
            {timing.label}
          </span>
          <span className="task-cycle-row-trigger muted">
            by {cycle.triggered_by}
          </span>
          {tokenSummary ? (
            <span
              className="task-cycle-row-tokens muted"
              data-testid="task-cycle-row-tokens"
              aria-label={tokenSummary.ariaLabel}
            >
              {tokenSummary.label}
            </span>
          ) : null}
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

function formatCycleTokenSummary(
  cycle: TaskCycle,
  attemptUsage?: TaskTokenUsageAttempt,
): { label: string; ariaLabel?: string } | null {
  const cycleUsage = cycle.token_usage;
  if (!cycleUsage?.known) {
    return null;
  }

  const tokens = formatTokenCount(cycleUsage.consumed_tokens);
  const sharePct = attemptUsage?.share_of_task_pct;
  if (sharePct == null) {
    return { label: tokens.label, ariaLabel: tokens.ariaLabel };
  }

  const shareLabel = formatShareOfTaskPct(sharePct);
  return {
    label: `${tokens.label} · ${shareLabel} of task`,
    ariaLabel: `${tokens.ariaLabel}, ${shareLabel} of task`,
  };
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
