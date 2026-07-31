import { useEffect } from "react";
import { errorMessage } from "@/lib/errorMessage";
import { useNow } from "@/shared/useNow";
import { formatDurationSeconds } from "@/observability";
import {
  phaseLabel,
  phaseStatusFillClass,
  phaseStatusLabel,
} from "@/tasks/cycleDisplay/cyclesViewModel";
import {
  HANDOFF_CLAIMS_MESSAGE,
  resolveAgentProgressMessage,
} from "@/tasks/cycleDisplay/agentProgressDisplay";
import type { TaskCycle } from "@/types/cycle";
import {
  hydrateAgentRunProgress,
  useAgentRunProgress,
  type AgentRunProgressItem,
} from "../../../hooks/useAgentRunProgress";
import { useTaskCycle, useTaskCycleStream } from "../../../hooks/useTaskCycles";
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
          {HANDOFF_CLAIMS_MESSAGE}
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
        now={now}
      />
    </div>
  );
}

function idlePendingMessage(items: ReadonlyArray<AgentRunProgressItem>): string {
  const last = items[items.length - 1];
  if (last) {
    const { kind, subtype, tool, message } = last.progress;
    if (
      (kind === "tool_call" || kind === "tool") &&
      subtype !== "completed" &&
      subtype !== "success" &&
      subtype !== "done" &&
      subtype !== "failed" &&
      subtype !== "error"
    ) {
      const toolName = tool?.trim();
      if (toolName) return `Running ${toolName}…`;
      if (message?.trim()) return message;
      return "Tool in progress…";
    }
  }
  for (let i = items.length - 1; i >= 0; i -= 1) {
    const entry = items[i];
    if (entry.progress.kind !== "run_state") continue;
    const remapped = resolveAgentProgressMessage(
      entry.progress.subtype,
      entry.progress.message,
      entry.progress.tool,
      "",
    );
    if (remapped.trim()) return remapped;
  }
  return "Agent working…";
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
  const stream = useTaskCycleStream(taskId, cycleId, { enabled: true });
  useEffect(() => {
    if (stream.events.length === 0) return;
    hydrateAgentRunProgress(taskId, cycleId, phaseSeq, stream.events);
  }, [taskId, cycleId, phaseSeq, stream.events]);

  const items = useAgentRunProgress(taskId, cycleId, phaseSeq);
  return (
    <CycleLiveProgressList
      items={items}
      now={now}
      showPendingRow={items.length > 0}
      emptyMessage="Waiting for agent updates…"
      pendingMessage={idlePendingMessage(items)}
    />
  );
}
