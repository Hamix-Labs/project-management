import type { TaskStatsResponse } from "@/types";
import type { TaskListLifecycleFilter } from "./taskListClientFilter";

export type TaskListStatusTab = {
  value: TaskListLifecycleFilter;
  label: string;
};

/** GitHub-style Open | Closed lifecycle tabs (fine status lives in filters). */
export function taskListStatusTabs(): TaskListStatusTab[] {
  return [
    { value: "open", label: "Open" },
    { value: "closed", label: "Closed" },
  ];
}

export function taskListLifecycleTabCount(
  value: TaskListLifecycleFilter,
  stats: TaskStatsResponse | null | undefined,
): number | null {
  if (!stats) return null;
  const closed = stats.by_status?.closed ?? 0;
  switch (value) {
    case "open":
      return Math.max(0, stats.total - closed);
    case "closed":
      return closed;
    default:
      return null;
  }
}
