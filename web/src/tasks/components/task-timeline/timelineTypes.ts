import type { Priority } from "@/types";

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
  /**
   * Contextual non-ID phrase after the title. Empty when there is nothing
   * useful beyond the task ref (shown once beside the timestamp).
   */
  highlight: string;
  detail: string;
  /** Full task id for deep links when available. */
  taskId?: string;
  /** Short display id (e.g. `#N` or first 8 hex chars) — shown once. */
  taskRef?: string;
  /**
   * Audit-trail sequence number on the owning task. When present,
   * clicking the event navigates to `/tasks/{taskId}/events/{seq}`.
   */
  seq?: number;
  meta?: string[];
  /** Joined task title — Timeline title search only (not rendered). */
  taskTitle?: string;
  /** Joined task priority — Timeline priority filter. */
  taskPriority?: Priority;
  /** Joined task project id — Timeline project filter. */
  taskProjectId?: string;
  /** Joined task tags — Timeline tag filter. */
  taskTags?: string[];
};

/** Range identifier for the date-range dropdown. */
export type TimelineRangeId = "24h" | "7d" | "30d" | "90d" | "all";

export type TimelineDateGroup = {
  label: string;
  events: TimelineEvent[];
};
