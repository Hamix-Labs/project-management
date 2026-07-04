import type { TaskStatsResponse } from "@/types";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { TaskListClientStatusFilter } from "./taskListClientFilter";

export type TaskListStatusTab = {
  value: TaskListClientStatusFilter;
  label: string;
};

/** Tab labels are inline — not shared with row badge copy (running tab: "In progress"). */
export function taskListStatusTabs(): TaskListStatusTab[] {
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const tabs: TaskListStatusTab[] = [
    { value: "all", label: "All" },
    { value: "ready", label: "Ready" },
    { value: "running", label: "In progress" },
    { value: "review", label: "Review" },
    { value: "blocked", label: "Blocked" },
    { value: "failed", label: "Failed" },
    { value: "done", label: "Done" },
    { value: "on_hold", label: "On hold" },
  ];
  if (scheduleUiEnabled) {
    tabs.push({ value: "scheduled", label: "Scheduled" });
  }
  return tabs;
}

export function taskListStatusTabCount(
  value: TaskListClientStatusFilter,
  stats: TaskStatsResponse | null | undefined,
): number | null {
  if (!stats) return null;
  switch (value) {
    case "all":
      return stats.total;
    case "ready":
      return stats.ready ?? stats.by_status?.ready ?? 0;
    case "scheduled":
      return stats.scheduled ?? 0;
    case "running":
    case "review":
    case "blocked":
    case "failed":
    case "done":
    case "on_hold":
      return stats.by_status?.[value] ?? 0;
    default:
      return null;
  }
}
