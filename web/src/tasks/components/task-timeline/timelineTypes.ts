export type TimelineCategory = "tasks" | "agents" | "verification";

export type TimelineKind =
  | "task-created"
  | "status-changed"
  | "verification-passed"
  | "verification-failed"
  | "agent-started"
  | "agent-finished"
  | "review-approved"
  | "comment";

export type TimelineEvent = {
  id: string;
  kind: TimelineKind;
  category: TimelineCategory;
  /** ISO timestamp — source of truth for grouping. */
  at: string;
  title: string;
  highlight: string;
  detail: string;
  /** Full task id for deep links when available. */
  taskId?: string;
  /** Short display id (e.g. first 8 hex chars). */
  taskRef?: string;
  /**
   * Audit-trail sequence number on the owning task. When present,
   * clicking the event navigates to `/tasks/{taskId}/events/{seq}`.
   */
  seq?: number;
  meta?: string[];
};

/** Range identifier for the date-range dropdown. */
export type TimelineRangeId = "24h" | "7d" | "30d" | "90d" | "all";

export type TimelineDateGroup = {
  label: string;
  events: TimelineEvent[];
};
