import {
  cycleStatusFillClass,
  cycleStatusLabel,
  formatRunnerModel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import type { TaskCycleDetail } from "@/types";
import type { AttemptTimelineDisplay } from "./attemptTimelineDisplay";

type Props = {
  cycle: TaskCycleDetail;
  timelineDisplay: AttemptTimelineDisplay;
};

export function AttemptDetailHeader({ cycle, timelineDisplay }: Props) {
  return (
    <header className="task-attempt-header">
      <div className="task-attempt-title-group">
        <div className="task-attempt-title-row">
          <h2 className="task-detail-title">Attempt #{cycle.attempt_seq}</h2>
          <span className={`cell-pill ${cycleStatusFillClass(cycle.status)}`}>
            {cycleStatusLabel(cycle.status)}
          </span>
        </div>
        <p className="task-attempt-meta-inline">
          <span className="task-attempt-meta-inline-item">
            {formatRunnerModel(cycle.cycle_meta)}
          </span>
          <time
            className="task-attempt-meta-inline-item"
            dateTime={cycle.started_at}
          >
            {timelineDisplay.startedParts.date} at {timelineDisplay.startedParts.time}
          </time>
          <span className="task-attempt-meta-inline-item">
            {timelineDisplay.durationLabel}
          </span>
        </p>
      </div>
    </header>
  );
}
