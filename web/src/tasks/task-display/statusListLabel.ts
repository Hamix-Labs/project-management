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
      return "Review";
    case "done":
      return "Done";
    case "failed":
      return "Failed";
    case "on_hold":
      return "On hold";
  }
}
