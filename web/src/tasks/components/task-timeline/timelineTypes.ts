export type TimelineCategory = "tasks" | "agents" | "verification";

export type TimelineKind =
  | "task-created"
  | "status-changed"
  | "verification-passed"
  | "verification-failed"
  | "agent-started"
  | "agent-finished"
  | "comment";

export type TimelineEvent = {
  id: string;
  kind: TimelineKind;
  category: TimelineCategory;
  /** ISO timestamp — source of truth for range filtering and grouping. */
  at: string;
  title: string;
  highlight: string;
  detail: string;
  /** Full task id for deep links when available. */
  taskId?: string;
  /** Short display id (e.g. first 8 hex chars). */
  taskRef?: string;
  meta?: string[];
};

export type TimelineFilterId = "all" | "tasks" | "verification";

export type TimelineRangeId = "24h" | "7d" | "30d" | "90d" | "all";

export type TimelineDateGroup = {
  label: string;
  events: TimelineEvent[];
};
