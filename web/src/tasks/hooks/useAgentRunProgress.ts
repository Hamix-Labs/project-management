import { useSyncExternalStore } from "react";

export type AgentRunProgress = {
  kind: string;
  subtype?: string;
  message?: string;
  tool?: string;
};

export type AgentRunProgressFrame = {
  taskId: string;
  cycleId: string;
  phaseSeq: number;
  progress: AgentRunProgress;
};

export type AgentRunProgressItem = AgentRunProgressFrame & {
  receivedAt: number;
};

const MAX_ITEMS_PER_PHASE = 5;
const MAX_TRACKED_PHASES = 50;
const EMPTY_PROGRESS: AgentRunProgressItem[] = [];

const progressByPhase = new Map<string, AgentRunProgressItem[]>();
const listeners = new Set<() => void>();

function keyFor(taskId: string, cycleId: string, phaseSeq: number): string {
  return `${taskId}:${cycleId}:${phaseSeq}`;
}

function emitChange(): void {
  for (const listener of listeners) {
    listener();
  }
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function snapshotFor(taskId: string, cycleId: string, phaseSeq: number): AgentRunProgressItem[] {
  return progressByPhase.get(keyFor(taskId, cycleId, phaseSeq)) ?? EMPTY_PROGRESS;
}

export function pushAgentRunProgress(frame: AgentRunProgressFrame): void {
  if (
    frame.taskId.trim() === "" ||
    frame.cycleId.trim() === "" ||
    frame.phaseSeq <= 0 ||
    frame.progress.kind.trim() === ""
  ) {
    return;
  }
  const key = keyFor(frame.taskId, frame.cycleId, frame.phaseSeq);
  const current = progressByPhase.get(key) ?? [];
  progressByPhase.set(key, [
    ...current,
    { ...frame, receivedAt: Date.now() },
  ].slice(-MAX_ITEMS_PER_PHASE));

  while (progressByPhase.size > MAX_TRACKED_PHASES) {
    const oldest = progressByPhase.keys().next().value as string | undefined;
    if (oldest === undefined) break;
    progressByPhase.delete(oldest);
  }
  emitChange();
}

/**
 * Seed the ephemeral live ticker from durable cycle stream events when the
 * SSE map is empty (reload / missed frames). Does not overwrite existing
 * live items. `receivedAt` uses each event's `at` when parseable.
 */
export function hydrateAgentRunProgress(
  taskId: string,
  cycleId: string,
  phaseSeq: number,
  events: ReadonlyArray<{
    kind: string;
    subtype?: string;
    message?: string;
    tool?: string;
    at?: string;
    phase_seq?: number;
  }>,
): void {
  if (
    taskId.trim() === "" ||
    cycleId.trim() === "" ||
    phaseSeq <= 0 ||
    events.length === 0
  ) {
    return;
  }
  const key = keyFor(taskId, cycleId, phaseSeq);
  if ((progressByPhase.get(key)?.length ?? 0) > 0) {
    return;
  }
  const forPhase = events.filter(
    (ev) => (ev.phase_seq ?? phaseSeq) === phaseSeq && ev.kind.trim() !== "",
  );
  if (forPhase.length === 0) return;
  const items: AgentRunProgressItem[] = forPhase.slice(-MAX_ITEMS_PER_PHASE).map((ev) => {
    const parsed = ev.at ? Date.parse(ev.at) : Number.NaN;
    return {
      taskId,
      cycleId,
      phaseSeq,
      receivedAt: Number.isFinite(parsed) ? parsed : Date.now(),
      progress: {
        kind: ev.kind,
        subtype: ev.subtype,
        message: ev.message,
        tool: ev.tool,
      },
    };
  });
  progressByPhase.set(key, items);
  while (progressByPhase.size > MAX_TRACKED_PHASES) {
    const oldest = progressByPhase.keys().next().value as string | undefined;
    if (oldest === undefined) break;
    progressByPhase.delete(oldest);
  }
  emitChange();
}

/** Test helper — clears the in-memory progress map. */
export function resetAgentRunProgressForTests(): void {
  progressByPhase.clear();
  emitChange();
}

export function useAgentRunProgress(
  taskId: string,
  cycleId: string,
  phaseSeq: number,
): AgentRunProgressItem[] {
  return useSyncExternalStore(
    subscribe,
    () => snapshotFor(taskId, cycleId, phaseSeq),
    () => EMPTY_PROGRESS,
  );
}
