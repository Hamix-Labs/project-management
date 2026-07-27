import { useMemo } from "react";
import { errorMessage } from "@/lib/errorMessage";
import {
  EmptyState,
  EmptyStateTimelineGlyph,
} from "@/shared/EmptyState";
import type { Phase, PhaseStatus } from "@/types/cycle";
import { useTaskCycles } from "../../../hooks/useTaskCycles";
import { TaskDetailCollapsibleSection } from "../layout/TaskDetailCollapsibleSection";
import { CycleHistoryList } from "./CycleHistoryList";
import { CurrentPhaseTicker } from "./CurrentPhaseTicker";
import { CyclesLoading } from "./CyclesLoading";
import { indexCyclesById, splitRunningAndHistory } from "./cyclePanelUtils";

type Props = {
  taskId: string;
  /**
   * Defaults to true. Pass `false` to suppress the panel entirely
   * (e.g. while the parent task query is still pending) so we don't
   * race the task fetch with a 404 from `/tasks/{id}/cycles` when
   * the id is still being resolved.
   */
  enabled?: boolean;
};

/**
 * Per-task observability surface mounted on TaskDetailPage. Composes
 * a live current-phase ticker and a history list of every cycle.
 */
export function TaskCyclesPanel({ taskId, enabled = true }: Props) {
  const cyclesQuery = useTaskCycles(taskId, { enabled });
  const retryCycles = cyclesQuery.refetch;

  const { runningCycle, historyCycles } = useMemo(
    () => splitRunningAndHistory(cyclesQuery.data),
    [cyclesQuery.data],
  );

  const cyclesById = useMemo(
    () => indexCyclesById(cyclesQuery.data?.cycles ?? []),
    [cyclesQuery.data?.cycles],
  );

  const count =
    !cyclesQuery.isPending && !cyclesQuery.isError
      ? (cyclesQuery.data?.cycles ?? []).length
      : null;

  return (
    <TaskDetailCollapsibleSection
      className="task-cycles-panel"
      title="Attempts"
      headingId="task-detail-cycles-heading"
      count={count}
      defaultOpen
    >
      {cyclesQuery.isPending ? (
        <CyclesLoading />
      ) : cyclesQuery.isError ? (
        <div className="err" role="alert">
          <p>
            {errorMessage(
              cyclesQuery.error,
              "Could not load attempts.",
            )}
          </p>
          <div className="task-detail-error-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => {
                void retryCycles();
              }}
            >
              Try again
            </button>
          </div>
        </div>
      ) : historyCycles.length === 0 && !runningCycle ? (
        <EmptyState
          icon={<EmptyStateTimelineGlyph />}
          title="No attempts yet"
          description="Each agent run is recorded here as execute → verify."
        />
      ) : (
        <>
          {runningCycle ? (
            <CurrentPhaseTicker
              taskId={taskId}
              cycle={runningCycle}
              cyclesById={cyclesById}
            />
          ) : null}
          <CycleHistoryList
            taskId={taskId}
            cycles={historyCycles}
            runningCycleId={runningCycle?.id ?? null}
          />
        </>
      )}
    </TaskDetailCollapsibleSection>
  );
}

// Re-exported for tests so they can construct fixtures without
// owning the Phase/PhaseStatus type imports.
export type { Phase, PhaseStatus };
