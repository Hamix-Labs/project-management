import type { AgentRunProgressItem } from "@/tasks/hooks/useAgentRunProgress";
import {
  agentProgressKindDescriptor,
  agentProgressMessage,
  formatAgentProgressClockTime,
  formatAgentProgressElapsed,
  type AgentProgressKindTone,
} from "./agentProgressDisplay";

export type LiveTimelineIconRole =
  | "working"
  | "done"
  | "call"
  | "failed"
  | "neutral";

export type LiveTimelineRow = {
  key: string;
  icon: LiveTimelineIconRole;
  kindLabel: string;
  kindTitle: string;
  message: string;
  messageEmphasis: "primary" | "secondary";
  timeLabel: string;
  receivedAt: number | null;
  isPending?: boolean;
};

export type ToLiveTimelineRowsOptions = {
  maxItems?: number;
  showPendingRow?: boolean;
  pendingMessage?: string;
  timestampMode?: "relative" | "clock";
};

/** Maps descriptor tone to the timeline icon role used by the live feed. */
export function liveTimelineIconForTone(
  tone: AgentProgressKindTone,
): LiveTimelineIconRole {
  if (tone === "done") return "done";
  if (tone === "tool") return "call";
  if (tone === "failed" || tone === "error") return "failed";
  return "neutral";
}

/**
 * Builds newest-first timeline rows for the live progress feed.
 * Optionally prepends a synthetic Working row when showPendingRow is set.
 */
export function toLiveTimelineRows(
  items: ReadonlyArray<AgentRunProgressItem>,
  now: number,
  options: ToLiveTimelineRowsOptions = {},
): LiveTimelineRow[] {
  const maxItems = options.maxItems ?? 3;
  const showPendingRow = options.showPendingRow ?? true;
  const pendingMessage = options.pendingMessage ?? "Waiting…";
  const timestampMode = options.timestampMode ?? "relative";

  const newestFirst = [...items].sort((a, b) => b.receivedAt - a.receivedAt);
  const latest = newestFirst[0];
  const rows: LiveTimelineRow[] = [];

  if (showPendingRow && items.length > 0) {
    rows.push({
      key: "pending",
      icon: "working",
      kindLabel: "Working",
      kindTitle: "Waiting for the next agent update",
      message: pendingMessage,
      messageEmphasis: "secondary",
      timeLabel: latest
        ? formatTimestamp(latest.receivedAt, now, timestampMode)
        : "",
      receivedAt: latest?.receivedAt ?? null,
      isPending: true,
    });
  }

  for (let index = 0; index < newestFirst.length && index < maxItems; index += 1) {
    const entry = newestFirst[index];
    const descriptor = agentProgressKindDescriptor(
      entry.progress.kind,
      entry.progress.subtype,
      entry.progress.tool,
    );
    const icon = liveTimelineIconForTone(descriptor.tone);
    rows.push({
      key: `${entry.receivedAt}:${index}:${entry.progress.kind}:${entry.progress.subtype ?? ""}`,
      icon,
      kindLabel: descriptor.label,
      kindTitle: descriptor.title,
      message: agentProgressMessage(entry),
      messageEmphasis: icon === "working" ? "secondary" : "primary",
      timeLabel: formatTimestamp(entry.receivedAt, now, timestampMode),
      receivedAt: entry.receivedAt,
    });
  }

  return rows;
}

function formatTimestamp(
  receivedAt: number,
  now: number,
  mode: "relative" | "clock",
): string {
  if (mode === "clock") {
    return formatAgentProgressClockTime(receivedAt);
  }
  return formatAgentProgressElapsed(receivedAt, now);
}
