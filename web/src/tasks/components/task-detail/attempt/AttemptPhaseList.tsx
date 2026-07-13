import type { CycleStatus, TaskCycleDetail, TaskCyclePhase } from "@/types";
import {
  phaseLabel,
  phaseStatusFillClass,
  phaseStatusLabel,
} from "@/observability";
import { useNow } from "@/shared/useNow";
import { CycleLiveProgressList } from "../cycles/CycleLiveProgressList";
import { useAgentRunProgress } from "@/tasks/hooks/useAgentRunProgress";
import type { AttemptTimelineDisplay } from "./attemptTimelineDisplay";
import { PhaseSeqBadge } from "./AttemptPhaseSeqBadge";

type AttemptPhaseListProps = {
  taskId: string;
  cycleId: string;
  cycle: TaskCycleDetail;
  timelineDisplay: AttemptTimelineDisplay;
};

export function AttemptPhaseList({
  taskId,
  cycleId,
  cycle,
  timelineDisplay,
}: AttemptPhaseListProps) {
  const {
    showPhaseBadge,
    showEndcap,
    endcapLabel,
    endcapTime,
    showStartCap,
    startCapTime,
  } = timelineDisplay;

  return (
    <div className="task-attempt-phase-timeline">
      {showStartCap ? (
        <AttemptStartCap
          startedAt={cycle.started_at}
          startedTime={startCapTime}
        />
      ) : null}
      <ol
        className={[
          "task-attempt-phase-track",
          showPhaseBadge && "task-attempt-phase-track--numbered",
          showEndcap && "task-attempt-phase-track--with-endcap",
        ]
          .filter(Boolean)
          .join(" ")}
      >
        {cycle.phases.map((phase, index) => (
          <AttemptPhaseStep
            key={phase.id}
            taskId={taskId}
            cycleId={cycleId}
            phase={phase}
            index={index}
            phaseCount={cycle.phases.length}
            showPhaseBadge={showPhaseBadge}
            showEndcap={showEndcap}
          />
        ))}
      </ol>
      {showEndcap && endcapLabel ? (
        <AttemptTerminalEndcap
          status={cycle.status}
          label={endcapLabel}
          endedAt={cycle.ended_at}
          endedTime={endcapTime}
        />
      ) : null}
    </div>
  );
}

function AttemptPhaseStep({
  taskId,
  cycleId,
  phase,
  index,
  phaseCount,
  showPhaseBadge,
  showEndcap,
}: {
  taskId: string;
  cycleId: string;
  phase: TaskCyclePhase;
  index: number;
  phaseCount: number;
  showPhaseBadge: boolean;
  showEndcap: boolean;
}) {
  const stepClass = "task-attempt-phase-step";
  const main = (
    <div className="task-attempt-phase-step-main">
      <span className="task-attempt-phase-step-name">
        {phaseLabel(phase.phase)}
      </span>
      <span className={`cell-pill ${phaseStatusFillClass(phase.status)}`}>
        {phaseStatusLabel(phase.status)}
      </span>
      {showPhaseBadge ? <PhaseSeqBadge seq={phase.phase_seq} /> : null}
    </div>
  );

  return (
    <li
      className={stepClass}
      data-status={phase.status}
      data-last={
        !showEndcap && index === phaseCount - 1 ? "true" : undefined
      }
    >
      <span className="task-attempt-phase-step-marker" aria-hidden="true" />
      {main}
      <LivePhaseTail taskId={taskId} cycleId={cycleId} phase={phase} />
    </li>
  );
}

function LivePhaseTail({
  taskId,
  cycleId,
  phase,
}: {
  taskId: string;
  cycleId: string;
  phase: TaskCyclePhase;
}) {
  const live = useAgentRunProgress(taskId, cycleId, phase.phase_seq);
  const now = useNow({
    enabled: phase.status === "running" && live.length > 0,
    intervalMs: 1000,
  });
  if (phase.status !== "running" || live.length === 0) return null;
  return (
    <div className="task-attempt-live-tail" aria-live="polite">
      <div className="task-attempt-live-tail-heading">
        <span className="cycle-live-dot task-attempt-live-dot" aria-hidden="true" />
        <span>Live</span>
      </div>
      <CycleLiveProgressList
        items={live}
        now={now}
        listAriaLabel="Recent live updates"
        timestampMode="clock"
        showPendingRow
      />
    </div>
  );
}

/**
 * Closes the phase rail with a single marker representing the cycle's
 * terminal outcome. The phase track on its own ends abruptly at the last
 * phase row, leaving the reader to look up at the header pill to learn
 * how the attempt as a whole ended; this endcap puts the rollup at the
 * natural end of the timeline. Only rendered for terminal statuses —
 * running attempts intentionally leave the rail open so the brand-color
 * halo on the running phase remains the dominant liveness signal.
 */
function AttemptTerminalEndcap({
  status,
  label,
  endedAt,
  endedTime,
}: {
  status: CycleStatus;
  label: string;
  endedAt: string | undefined;
  endedTime: string | null;
}) {
  return (
    <div
      className="task-attempt-phase-endcap"
      data-status={status}
      aria-label={
        endedTime ? `${label} at ${endedTime}` : label
      }
    >
      <span className="task-attempt-phase-endcap-marker" aria-hidden="true" />
      <span className="task-attempt-phase-endcap-name">{label}</span>
      {endedTime && endedAt ? (
        <time className="task-attempt-phase-endcap-time" dateTime={endedAt}>
          {endedTime}
        </time>
      ) : null}
    </div>
  );
}

/**
 * Opens the phase rail so the timeline reads as a complete arc:
 * "Attempt started → phases → Attempt {completed/failed/aborted}".
 * Always rendered when the cycle has phases.
 */
function AttemptStartCap({
  startedAt,
  startedTime,
}: {
  startedAt: string;
  startedTime: string | null;
}) {
  const label = "Attempt started";
  return (
    <div
      className="task-attempt-phase-startcap"
      aria-label={startedTime ? `${label} at ${startedTime}` : label}
    >
      <span className="task-attempt-phase-startcap-marker" aria-hidden="true" />
      <span className="task-attempt-phase-startcap-name">{label}</span>
      {startedTime ? (
        <time className="task-attempt-phase-startcap-time" dateTime={startedAt}>
          {startedTime}
        </time>
      ) : null}
    </div>
  );
}
