import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { StatusBadge } from "@/components/task-status";
import type { Status, Task } from "@/types";
import { useAppTimezone, formatInAppTimezone } from "@/shared/time/appTimezone";
import { statusNeedsUserInput } from "../../../task-display";
import { ScheduleGlyph } from "./TaskDetailScheduleGlyphs";

type Props = {
  task: Pick<Task, "status" | "pickup_not_before">;
};

const TERMINAL_STATUSES: ReadonlySet<Status> = new Set(["done", "failed"]);

/**
 * Toolbar strip: task status plus optional read-only pickup schedule.
 * Schedule mutations live in the edit-task modal (`SchedulePicker`).
 */
export function TaskDetailSchedule({ task }: Props) {
  const tz = useAppTimezone();
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const isTerminal = TERMINAL_STATUSES.has(task.status);
  const hasSchedule = Boolean(task.pickup_not_before);
  const needsUser = statusNeedsUserInput(task.status);

  const formatted = task.pickup_not_before
    ? formatInAppTimezone(task.pickup_not_before, tz)
    : null;

  return (
    <div
      className="task-detail-schedule"
      data-testid="task-detail-schedule"
      data-state={hasSchedule ? "scheduled" : "unscheduled"}
    >
      <div
        className="task-detail-schedule-row task-detail-schedule-row--status"
        data-testid="task-detail-status"
      >
        <StatusBadge
          status={task.status}
          className="task-detail-status-badge"
          data-needs-user={needsUser ? "true" : undefined}
        />
      </div>
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
      ) : scheduleUiEnabled && !isTerminal ? (
        <span className="task-detail-schedule-empty muted">
          No pickup scheduled.
        </span>
      ) : null}
    </div>
  );
}
