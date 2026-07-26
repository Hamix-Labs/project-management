import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { Status, Task } from "@/types";
import { useAppTimezone, formatInAppTimezone } from "@/shared/time/appTimezone";
import { ScheduleGlyph } from "./TaskDetailScheduleGlyphs";

type Props = {
  task: Pick<Task, "status" | "pickup_not_before">;
};

const TERMINAL_STATUSES: ReadonlySet<Status> = new Set(["done", "failed"]);

/**
 * Read-only pickup schedule strip for the task-detail toolbar.
 * Status lives in the execution bar; schedule mutations live in the
 * edit-task modal (`SchedulePicker`).
 */
export function TaskDetailSchedule({ task }: Props) {
  const tz = useAppTimezone();
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const isTerminal = TERMINAL_STATUSES.has(task.status);
  const hasSchedule = Boolean(task.pickup_not_before);

  const formatted = task.pickup_not_before
    ? formatInAppTimezone(task.pickup_not_before, tz)
    : null;

  if (!scheduleUiEnabled) {
    return null;
  }

  if (!hasSchedule && isTerminal) {
    return null;
  }

  return (
    <div
      className="task-detail-schedule"
      data-testid="task-detail-schedule"
      data-state={hasSchedule ? "scheduled" : "unscheduled"}
    >
      {hasSchedule ? (
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
      ) : (
        <span className="task-detail-schedule-empty muted">
          No pickup scheduled.
        </span>
      )}
    </div>
  );
}
