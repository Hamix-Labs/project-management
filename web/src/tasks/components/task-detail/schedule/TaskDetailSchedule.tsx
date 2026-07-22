import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { Status, Task } from "@/types";
import { useAppTimezone, formatInAppTimezone } from "@/shared/time/appTimezone";
import { useTaskCycles } from "@/tasks/hooks/useTaskCycles";
import { PhaseCompleteGlyph, ScheduleGlyph } from "./TaskDetailScheduleGlyphs";
import {
  earliestCycleStartedAt,
  formatTaskCompletionDuration,
} from "./taskCompletionDuration";

type Props = {
  task: Pick<Task, "id" | "status" | "pickup_not_before" | "criteria_satisfied_at">;
};

const TERMINAL_STATUSES: ReadonlySet<Status> = new Set(["done", "failed"]);

/** Task is executing; completion time is not known until criteria are satisfied. */
const PHASE_COMPLETE_PLACEHOLDER_STATUSES: ReadonlySet<Status> = new Set([
  "running",
]);

/**
 * Read-only pickup schedule line on the task detail toolbar. Mutations
 * live in the edit-task modal (`TaskCreateModal` edit mode + `SchedulePicker`).
 */
export function TaskDetailSchedule({ task }: Props) {
  const tz = useAppTimezone();
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const isTerminal = TERMINAL_STATUSES.has(task.status);
  const hasSchedule = Boolean(task.pickup_not_before);
  const phaseCompleteAt = (task.criteria_satisfied_at ?? "").trim();
  const hasPhaseComplete = phaseCompleteAt !== "";
  const showPhaseCompletePlaceholder =
    !hasPhaseComplete && PHASE_COMPLETE_PLACEHOLDER_STATUSES.has(task.status);

  const cyclesQuery = useTaskCycles(task.id, {
    enabled: hasPhaseComplete && Boolean(task.id),
  });
  const startedAt = earliestCycleStartedAt(cyclesQuery.data?.cycles ?? []);
  const durationLabel = hasPhaseComplete
    ? formatTaskCompletionDuration(startedAt, phaseCompleteAt)
    : null;

  if (!hasPhaseComplete && !scheduleUiEnabled && !showPhaseCompletePlaceholder) {
    return null;
  }

  if (!hasSchedule && isTerminal && !hasPhaseComplete) {
    return null;
  }

  const formatted = task.pickup_not_before
    ? formatInAppTimezone(task.pickup_not_before, tz)
    : null;
  const phaseFormatted = hasPhaseComplete
    ? formatInAppTimezone(phaseCompleteAt, tz)
    : null;
  const phaseAria = durationLabel
    ? `Phase completed, ${phaseFormatted}, took ${durationLabel}`
    : `Phase completed, ${phaseFormatted}`;

  return (
    <div
      className="task-detail-schedule"
      data-testid="task-detail-schedule"
      data-state={hasSchedule ? "scheduled" : "unscheduled"}
    >
      {hasPhaseComplete ? (
        <div
          className="task-detail-schedule-row task-detail-schedule-row--phase"
          data-testid="task-detail-phase-complete"
          aria-label={phaseAria}
        >
          <span className="task-detail-schedule-row-icon" aria-hidden="true">
            <PhaseCompleteGlyph />
          </span>
          <div className="task-detail-schedule-row-body">
            <span className="task-detail-schedule-row-label">Completed</span>
            <span className="task-detail-schedule-row-sep" aria-hidden="true">
              ·
            </span>
            <time dateTime={phaseCompleteAt}>{phaseFormatted}</time>
            {durationLabel ? (
              <>
                <span className="task-detail-schedule-row-sep" aria-hidden="true">
                  ·
                </span>
                <span data-testid="task-detail-phase-duration">{durationLabel}</span>
              </>
            ) : null}
          </div>
        </div>
      ) : showPhaseCompletePlaceholder ? (
        <div
          className="task-detail-schedule-row task-detail-schedule-row--phase task-detail-schedule-row--phase-pending"
          data-testid="task-detail-phase-complete-pending"
          aria-label="Completed, pending"
        >
          <span className="task-detail-schedule-row-icon" aria-hidden="true">
            <PhaseCompleteGlyph />
          </span>
          <div className="task-detail-schedule-row-body">
            <span className="task-detail-schedule-row-label">Completed</span>
            <span className="task-detail-schedule-row-sep" aria-hidden="true">
              ·
            </span>
            <span className="task-detail-schedule-row-pending">—</span>
          </div>
        </div>
      ) : null}
      {scheduleUiEnabled && hasSchedule ? (
        <div
          className="task-detail-schedule-row task-detail-schedule-row--scheduled"
          data-testid="task-detail-schedule-badge"
          aria-label={`Scheduled for pickup, ${formatted}`}
        >
          <span className="task-detail-schedule-row-icon" aria-hidden="true">
            <ScheduleGlyph />
          </span>
          <div className="task-detail-schedule-row-body">
            <span className="task-detail-schedule-row-label">Scheduled</span>
            <span className="task-detail-schedule-row-sep" aria-hidden="true">
              ·
            </span>
            <time dateTime={task.pickup_not_before}>{formatted}</time>
          </div>
        </div>
      ) : scheduleUiEnabled && !hasPhaseComplete ? (
        <span className="task-detail-schedule-empty muted">
          No pickup scheduled.
        </span>
      ) : null}
    </div>
  );
}
