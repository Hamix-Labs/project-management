import { errorMessage } from "@/lib/errorMessage";
import { useNow } from "@/shared/useNow";
import { formatDurationSeconds } from "@/observability";
import {
  cycleStatusLabel,
  cycleStatusFillClass,
  cycleRunnerChipClass,
  formatRunnerModel,
  phaseLabel,
  phaseStatusFillClass,
  phaseStatusLabel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import type { TaskCycle } from "@/types/cycle";
import { useAgentRunProgress, type AgentRunProgressItem } from "../../../hooks/useAgentRunProgress";
import { useTaskCycle } from "../../../hooks/useTaskCycles";
import { formatCycleLineageLabel } from "../../../cycleDisplay/cycleLineage";
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

  return (
    <div className="task-cycle-ticker" data-testid="task-cycle-ticker">
      <div className="task-cycle-ticker-head">
        <span className="cycle-live-dot" aria-hidden="true" />
        <span className="task-cycle-ticker-eyebrow">Live</span>
        <span
          className={`cell-pill ${cycleStatusFillClass(cycle.status)}`}
          data-testid="task-cycle-ticker-status"
        >
          {cycleStatusLabel(cycle.status)}
        </span>
      </div>
      <CurrentPhaseLine
        taskId={taskId}
        cycleId={cycle.id}
        detailQuery={detailQuery}
        now={now}
      />
      <div className="task-cycle-ticker-meta">
        <span className="task-cycle-ticker-attempt">
          Attempt #{cycle.attempt_seq}
          {lineage ? (
            <span className="task-cycle-lineage muted"> · {lineage}</span>
          ) : null}
        </span>
        <span className="task-cycle-ticker-meta-sep" aria-hidden="true">
          ·
        </span>
        <span
          className={`cell-pill ${cycleRunnerChipClass()}`}
          data-testid="task-cycle-ticker-runner"
        >
          {formatRunnerModel(cycle.cycle_meta)}
        </span>
        <span className="task-cycle-ticker-meta-sep" aria-hidden="true">
          ·
        </span>
        <span
          className="task-cycle-ticker-elapsed"
          data-testid="task-cycle-ticker-elapsed"
        >
          Started {formatDurationSeconds(elapsedSeconds(cycle.started_at, now))} ago
        </span>
      </div>
    </div>
  );
}

function CurrentPhaseLine({
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
        Could not resolve current phase ({errorMessage(detailQuery.error, "unknown error")}).
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
          No phase started yet.
        </p>
      );
    }
    return (
      <p className="task-cycle-ticker-focus" data-testid="task-cycle-ticker-phase">
        Between phases · last:{" "}
        <span className={`cell-pill ${phaseStatusFillClass(lastPhase.status)}`}>
          {phaseLabel(lastPhase.phase)} {phaseStatusLabel(lastPhase.status).toLowerCase()}
        </span>
      </p>
    );
  }
  return (
    <>
      <div
        className="task-cycle-ticker-focus task-cycle-ticker-focus--running"
        data-testid="task-cycle-ticker-phase"
      >
        <span className="task-cycle-ticker-focus-label" aria-live="polite">
          <span className={`cell-pill ${phaseStatusFillClass(runningPhase.status)}`}>
            {phaseLabel(runningPhase.phase)}
          </span>
        </span>
        <span className="task-cycle-ticker-focus-elapsed" aria-hidden="true">
          {formatDurationSeconds(elapsedSeconds(runningPhase.started_at, now))}
        </span>
      </div>
      <div className="task-cycle-ticker-feed">
        <PhaseProgress
          taskId={taskId}
          cycleId={cycleId}
          phaseSeq={runningPhase.phase_seq}
          now={now}
        />
      </div>
    </>
  );
}

function idlePendingMessage(items: ReadonlyArray<AgentRunProgressItem>): string {
  for (let i = items.length - 1; i >= 0; i -= 1) {
    const entry = items[i];
    if (entry.progress.kind !== "run_state") continue;
    if (entry.progress.message?.trim()) {
      return entry.progress.message;
    }
    if (entry.progress.subtype === "idle_kill_pending") {
      return "Terminating agent soon if no output";
    }
    if (entry.progress.subtype === "idle_suspicious") {
      return "No agent output for a while — run may be stuck";
    }
  }
  return "Waiting for the next agent update…";
}

function PhaseProgress({
  taskId,
  cycleId,
  phaseSeq,
  now,
}: {
  taskId: string;
  cycleId: string;
  phaseSeq: number;
  now: number;
}) {
  const items = useAgentRunProgress(taskId, cycleId, phaseSeq);
  return (
    <CycleLiveProgressList
      items={items}
      now={now}
      showPendingRow={items.length > 0}
      emptyMessage="Waiting for the next agent update…"
      pendingMessage={idlePendingMessage(items)}
    />
  );
}
