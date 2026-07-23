import type { AgentRunProgressItem } from "@/tasks/hooks/useAgentRunProgress";
import { toLiveTimelineRows } from "@/tasks/cycleDisplay/liveTimelineRows";
import { CycleLiveProgressIcon } from "./CycleLiveProgressIcons";

export type CycleLiveProgressListProps = {
  items: ReadonlyArray<AgentRunProgressItem>;
  now: number;
  maxItems?: number;
  showPendingRow?: boolean;
  emptyMessage?: string;
  testId?: string;
  listAriaLabel?: string;
  timestampMode?: "relative" | "clock";
  pendingMessage?: string;
};

export function CycleLiveProgressList({
  items,
  now,
  maxItems = 3,
  showPendingRow,
  emptyMessage,
  testId = "task-cycle-progress-list",
  listAriaLabel = "Recent agent progress",
  timestampMode = "relative",
  pendingMessage = "Waiting…",
}: CycleLiveProgressListProps) {
  if (items.length === 0) {
    if (!emptyMessage) return null;
    return (
      <p
        className="task-cycle-progress-empty"
        data-testid="task-cycle-progress-empty"
      >
        {emptyMessage}
      </p>
    );
  }

  const shouldShowPending = showPendingRow ?? true;
  const rows = toLiveTimelineRows(items, now, {
    maxItems,
    showPendingRow: shouldShowPending,
    pendingMessage,
    timestampMode,
  });

  return (
    <ol
      className="task-cycle-progress-list"
      aria-label={listAriaLabel}
      data-testid={testId}
    >
      {rows.map((row, index) => {
        const isLatest = !row.isPending && index === (shouldShowPending ? 1 : 0);
        return (
          <li
            key={row.key}
            className={[
              "task-cycle-progress-item",
              row.isPending && "task-cycle-progress-item--pending",
              isLatest && "task-cycle-progress-item--latest",
            ]
              .filter(Boolean)
              .join(" ")}
            aria-label={row.isPending ? row.kindTitle : undefined}
            title={row.kindTitle}
          >
            <span className="task-cycle-progress-icon">
              <CycleLiveProgressIcon role={row.icon} />
            </span>
            <div className="task-cycle-progress-body">
              <span className="task-cycle-progress-kind">{row.kindLabel}</span>
              <span
                className={[
                  "task-cycle-progress-message",
                  row.messageEmphasis === "secondary" &&
                    "task-cycle-progress-message--secondary",
                ]
                  .filter(Boolean)
                  .join(" ")}
                title={row.message}
              >
                {row.message}
              </span>
            </div>
            {timestampMode === "clock" && row.receivedAt != null ? (
              <time
                className="task-cycle-progress-time"
                dateTime={new Date(row.receivedAt).toISOString()}
              >
                {row.timeLabel}
              </time>
            ) : (
              <span className="task-cycle-progress-time" aria-hidden="true">
                {row.timeLabel}
              </span>
            )}
          </li>
        );
      })}
    </ol>
  );
}
