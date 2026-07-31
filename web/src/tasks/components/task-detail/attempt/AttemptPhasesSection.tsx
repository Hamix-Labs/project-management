import type {
  TaskCycleDetail,
  TaskCyclePhase,
  TaskCycleStreamEvent,
} from "@/types";
import { phaseLabel } from "@/tasks/cycleDisplay/cyclesViewModel";
import type { AttemptTimelineDisplay } from "./attemptTimelineDisplay";
import { AttemptPhaseList } from "./AttemptPhaseList";
import { latestAgentReplyByPhase } from "./latestAgentReplyByPhase";

type AttemptPhasesSectionProps = {
  taskId: string;
  cycleId: string;
  cycle: TaskCycleDetail;
  timelineDisplay: AttemptTimelineDisplay;
  streamEvents: readonly TaskCycleStreamEvent[];
  filterPhaseSeq: number | null;
  onSelectPhase: (seq: number | null) => void;
  phaseFilterEnabled: boolean;
};

export function AttemptPhasesSection({
  taskId,
  cycleId,
  cycle,
  timelineDisplay,
  streamEvents,
  filterPhaseSeq,
  onSelectPhase,
  phaseFilterEnabled,
}: AttemptPhasesSectionProps) {
  const agentReplies = latestAgentReplyByPhase(streamEvents, cycle.phases);

  return (
    <section
      className="task-attempt-section task-attempt-section--phases"
      aria-labelledby="attempt-phases"
    >
      <div className="task-attempt-section-heading-row">
        <h3 className="task-detail-subheading" id="attempt-phases">
          <span>Phases</span>
        </h3>
        {phaseFilterEnabled ? (
          <AttemptPhaseFilterBar
            phases={cycle.phases}
            filterPhaseSeq={filterPhaseSeq}
            onSelectPhase={onSelectPhase}
          />
        ) : null}
      </div>
      <AttemptPhaseList
        taskId={taskId}
        cycleId={cycleId}
        cycle={cycle}
        timelineDisplay={timelineDisplay}
        agentReplies={agentReplies}
      />
    </section>
  );
}

function AttemptPhaseFilterBar({
  phases,
  filterPhaseSeq,
  onSelectPhase,
}: {
  phases: readonly TaskCyclePhase[];
  filterPhaseSeq: number | null;
  onSelectPhase: (seq: number | null) => void;
}) {
  return (
    <div
      className="task-attempt-activity-tabs task-attempt-phase-filter"
      role="group"
      aria-label="Filter activity by phase"
    >
      <button
        type="button"
        className={
          filterPhaseSeq === null
            ? "task-attempt-activity-tab task-attempt-activity-tab--active"
            : "task-attempt-activity-tab"
        }
        aria-pressed={filterPhaseSeq === null}
        onClick={() => onSelectPhase(null)}
      >
        All phases
      </button>
      {phases.map((phase) => {
        const active = filterPhaseSeq === phase.phase_seq;
        return (
          <button
            key={phase.id}
            type="button"
            className={
              active
                ? "task-attempt-activity-tab task-attempt-activity-tab--active"
                : "task-attempt-activity-tab"
            }
            aria-pressed={active}
            onClick={() => onSelectPhase(active ? null : phase.phase_seq)}
          >
            {phaseLabel(phase.phase)}
          </button>
        );
      })}
    </div>
  );
}
