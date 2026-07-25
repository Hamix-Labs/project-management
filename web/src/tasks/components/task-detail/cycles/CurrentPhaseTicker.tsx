import { errorMessage } from "@/lib/errorMessage";
import { useNow } from "@/shared/useNow";
import { formatDurationSeconds } from "@/observability";
import {
  phaseLabel,
  phaseStatusFillClass,
  phaseStatusLabel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import type { TaskCycle } from "@/types/cycle";
import {
  useAgentRunProgress,
  type AgentRunProgressItem,
} from "../../../hooks/useAgentRunProgress";
import { useTaskCycle } from "../../../hooks/useTaskCycles";
import { formatCycleLineageLabel } from "../../../cycleDisplay/cycleLineage";
import { CycleLiveCardHead } from "./CycleLiveCardHead";
import { CycleLiveCardMeta } from "./CycleLiveCardMeta";
import { CycleLiveProgressList } from "./CycleLiveProgressList";
import {
  elapsedSeconds,
  pickLatestPhase,
  pickRunningPhase,
} from "./cyclePanelUtils";

type CurrentPhaseTickerProps = {
  taskId: string;
  cycle: TaskCycle;
  cyclesById: ReadonlyMap<string, TaskCycle>;
};

/** Live "what is the agent doing right now?" indicator for the running cycle. */
export function CurrentPhaseTicker({
  taskId,
  cycle,
  cyclesById,
}: CurrentPhaseTickerProps) {
  const detailQuery = useTaskCycle(taskId, cycle.id);
  const now = useNow({ enabled: cycle.status === "running" });
  const lineage = formatCycleLineageLabel(cycle, cyclesById);

  const runningPhase =
    detailQuery.data != null
      ? pickRunningPhase(detailQuery.data.phases)
      : null;

  return (
    <div className="task-cycle-ticker" data-testid="task-cycle-ticker">
      <CycleLiveCardHead
        cycleStatus={cycle.status}
        phaseName={
          runningPhase ? phaseLabel(runningPhase.phase) : null
        }
        phaseElapsed={
          runningPhase
            ? formatDurationSeconds(
                elapsedSeconds(runningPhase.started_at, now),
              )
            : null
        }
      />
      <CurrentPhaseBody
        taskId={taskId}
        cycleId={cycle.id}
        detailQuery={detailQuery}
        now={now}
      />
      <CycleLiveCardMeta
        attemptSeq={cycle.attempt_seq}
        lineage={lineage}
        cycleMeta={cycle.cycle_meta}
        startedLabel={formatDurationSeconds(
          elapsedSeconds(cycle.started_at, now),
        )}
      />
    </div>
  );
}

function CurrentPhaseBody({
  taskId,
  cycleId,
  detailQuery,
  now,
}: {
  taskId: string;
  cycleId: string;
  detailQuery: ReturnType<typeof useTaskCycle>;
  now: number;
}) {
  if (detailQuery.isPending) {
    return (
      <p
        className="task-cycle-ticker-focus task-cycle-ticker-focus--pending"
        data-testid="task-cycle-ticker-phase"
        aria-busy="true"
      >
        Resolving current phase…
      </p>
    );
  }
  if (detailQuery.isError) {
    return (
      <p
        className="task-cycle-ticker-focus task-cycle-ticker-focus--error"
        data-testid="task-cycle-ticker-phase"
      >
        Could not resolve current phase (
        {errorMessage(detailQuery.error, "unknown error")}).
      </p>
    );
  }
  const detail = detailQuery.data;
  const runningPhase = pickRunningPhase(detail.phases);
  if (!runningPhase) {
    const lastPhase = pickLatestPhase(detail.phases);
    if (!lastPhase) {
      return (
        <p
          className="task-cycle-ticker-focus task-cycle-ticker-focus--idle"
          data-testid="task-cycle-ticker-phase"
        >
          Preparing workspace…
        </p>
      );
    }
    if (lastPhase.phase === "execute" && lastPhase.status === "succeeded") {
      return (
        <p
          className="task-cycle-ticker-focus"
          data-testid="task-cycle-ticker-phase"
        >
          Starting verify…
        </p>
      );
    }
    return (
      <p
        className="task-cycle-ticker-focus"
        data-testid="task-cycle-ticker-phase"
      >
        Between phases · last:{" "}
        <span className={`cell-pill ${phaseStatusFillClass(lastPhase.status)}`}>
          {phaseLabel(lastPhase.phase)}{" "}
          {phaseStatusLabel(lastPhase.status).toLowerCase()}
        </span>
      </p>
    );
  }

  return (
    <div className="task-cycle-ticker-feed">
      <PhaseProgress
        taskId={taskId}
        cycleId={cycleId}
        phaseSeq={runningPhase.phase_seq}
        phase={runningPhase.phase}
        now={now}
      />
    </div>
  );
}

function idlePendingMessage(items: ReadonlyArray<AgentRunProgressItem>): string {
  for (let i = items.length - 1; i >= 0; i -= 1) {
    const entry = items[i];
    if (entry.progress.kind !== "run_state") continue;
    if (entry.progress.message?.trim()) {
      return entry.progress.message;
    }
  }
  return "Agent working…";
}

function phaseEmptyMessage(phase: string): string {
  if (phase === "verify") {
    return "Running verify checks…";
  }
  return "Preparing execute…";
}

function PhaseProgress({
  taskId,
  cycleId,
  phaseSeq,
  phase,
  now,
}: {
  taskId: string;
  cycleId: string;
  phaseSeq: number;
  phase: string;
  now: number;
}) {
  const items = useAgentRunProgress(taskId, cycleId, phaseSeq);
  return (
    <CycleLiveProgressList
      items={items}
      now={now}
      showPendingRow={items.length > 0}
      emptyMessage={phaseEmptyMessage(phase)}
      pendingMessage={idlePendingMessage(items)}
    />
  );
}
