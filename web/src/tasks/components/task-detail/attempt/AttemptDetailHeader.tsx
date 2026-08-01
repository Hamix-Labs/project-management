import { CycleStatusBadge } from "@/components/task-status";
import { formatRunnerModel } from "@/tasks/cycleDisplay/cyclesViewModel";
import {
  CREATING_PR_STATUS_LABEL,
  isOpenPrRunKind,
} from "@/tasks/task-display/openPrRunDisplay";
import type { TaskCycleDetail } from "@/types";
import type { AttemptTimelineDisplay } from "./attemptTimelineDisplay";

type Props = {
  cycle: TaskCycleDetail;
  timelineDisplay: AttemptTimelineDisplay;
};

export function AttemptDetailHeader({ cycle, timelineDisplay }: Props) {
  const creatingPr =
    cycle.status === "running" && isOpenPrRunKind(cycle.meta);
  const durationLabel = creatingPr
    ? timelineDisplay.durationLabel.replace(/^Running for/, "Creating PR for")
    : timelineDisplay.durationLabel;
  return (
    <header className="task-attempt-header">
      <div className="task-attempt-title-group">
        <div className="task-attempt-title-row">
          <h2 className="task-detail-title">Attempt #{cycle.attempt_seq}</h2>
          <CycleStatusBadge
            status={cycle.status}
            label={creatingPr ? CREATING_PR_STATUS_LABEL : undefined}
          />
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
            {durationLabel}
          </span>
        </p>
      </div>
    </header>
  );
}
