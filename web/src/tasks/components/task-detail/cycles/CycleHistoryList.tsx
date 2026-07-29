import { useMemo } from "react";
import { CycleStatusBadge } from "@/components/task-status";
import { useAppTimezone } from "@/shared/time/appTimezone";
import type { TaskCycle, TaskTokenUsageAttempt } from "@/types/cycle";
import { useTaskTokenUsage } from "../../../hooks/useTaskTokenUsage";
import {
  formatShareOfTaskPct,
  formatTokenCount,
} from "../../../task-display/formatTokenCount";
import {
  ChevronRightGlyph,
  ClockGlyph,
  CoinsGlyph,
  TimerGlyph,
  UserGlyph,
} from "../ExecutionBarGlyphs";
import { formatAttemptTiming } from "./cyclePanelUtils";

type CycleHistoryListProps = {
  taskId: string;
  cycles: TaskCycle[];
};

export function CycleHistoryList({ taskId, cycles }: CycleHistoryListProps) {
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
  attemptUsage,
}: {
  taskId: string;
  cycle: TaskCycle;
  attemptUsage?: TaskTokenUsageAttempt;
}) {
  const tz = useAppTimezone();
  const timing = formatAttemptTiming(cycle, tz);
  const tokenSummary = formatCycleTokenSummary(cycle, attemptUsage);
  const detailsHref = `/tasks/${encodeURIComponent(taskId)}/cycles/${encodeURIComponent(cycle.id)}`;

  return (
    <li className="task-cycle-row" data-cycle-status={cycle.status}>
      <div className="task-cycle-row-summary">
        <div className="task-cycle-row-main">
          <div className="task-cycle-row-identity">
            <span className="task-cycle-row-attempt">
              Attempt #{cycle.attempt_seq}
            </span>
            {!timing.inProgress ? (
              <CycleStatusBadge
                status={cycle.status}
                className="task-cycle-row-status"
                data-testid="task-cycle-row-status"
              />
            ) : null}
          </div>

          <div className="task-cycle-row-aside">
            <span
              className="task-cycle-row-duration"
              data-in-progress={timing.inProgress ? "true" : undefined}
            >
              <TimerGlyph className="task-cycle-row-fact-icon" />
              <span>
                {timing.inProgress
                  ? "In progress"
                  : (timing.duration ?? "—")}
              </span>
            </span>
            <a className="task-cycle-row-attempt-link" href={detailsHref}>
              Details
              <ChevronRightGlyph className="task-cycle-row-attempt-link-icon" />
            </a>
          </div>
        </div>

        <div
          className="task-cycle-row-meta"
          aria-label={timing.ariaLabel}
          data-testid="task-cycle-row-when"
        >
          <span className="task-cycle-row-fact">
            <ClockGlyph className="task-cycle-row-fact-icon" />
            <span className="task-cycle-row-fact-text">
              {timing.inProgress || !timing.endedAt ? (
                timing.startedAt
              ) : (
                <>
                  <span>{timing.startedAt}</span>
                  <span className="task-cycle-row-fact-arrow" aria-hidden="true">
                    →
                  </span>
                  <span>{timing.endedAt}</span>
                </>
              )}
            </span>
          </span>
          <span className="task-cycle-row-fact">
            <UserGlyph className="task-cycle-row-fact-icon" />
            <span className="task-cycle-row-fact-text">{cycle.triggered_by}</span>
          </span>
          {tokenSummary ? (
            <span
              className="task-cycle-row-fact"
              data-testid="task-cycle-row-tokens"
              aria-label={tokenSummary.ariaLabel}
            >
              <CoinsGlyph className="task-cycle-row-fact-icon" />
              <span className="task-cycle-row-fact-text">{tokenSummary.label}</span>
            </span>
          ) : null}
        </div>
      </div>
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
