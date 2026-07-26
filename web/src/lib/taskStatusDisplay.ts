import type { Status } from "@/types";

/** Human-readable status copy for lists, filters, and detail chrome. */
export function statusListLabel(status: Status): string {
  switch (status) {
    case "ready":
      return "Ready";
    case "running":
      return "Running";
    case "blocked":
      return "Blocked";
    case "review":
      return "Awaiting review";
    case "done":
      return "Done";
    case "failed":
      return "Failed";
    case "on_hold":
      return "On hold";
    case "closed":
      return "Closed";
  }
}

export type StatusTone =
  | "success"
  | "info"
  | "warning"
  | "danger"
  | "neutral"
  | "review";

export type StatusMeta = {
  label: string;
  tone: StatusTone;
  order: number;
  pulse?: boolean;
};

/** Presentation metadata for status badges and client-side list sort. */
export const STATUS_META: Record<Status, StatusMeta> = {
  ready: { label: statusListLabel("ready"), tone: "success", order: 1 },
  running: {
    label: statusListLabel("running"),
    tone: "info",
    order: 2,
    pulse: true,
  },
  // Match cell-pill--status-review (purple), not warning amber.
  review: { label: statusListLabel("review"), tone: "review", order: 3 },
  blocked: { label: statusListLabel("blocked"), tone: "neutral", order: 4 },
  failed: { label: statusListLabel("failed"), tone: "danger", order: 5 },
  done: { label: statusListLabel("done"), tone: "success", order: 6 },
  on_hold: { label: statusListLabel("on_hold"), tone: "neutral", order: 7 },
  closed: { label: statusListLabel("closed"), tone: "neutral", order: 8 },
};
