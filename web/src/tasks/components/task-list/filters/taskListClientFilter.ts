import type { Priority, Status } from "@/types";
import type { TaskWithDepth } from "../../../task-tree";

/**
 * GitHub-style lifecycle tabs on the task list. `open` is every status
 * except `closed`; `closed` is only the operator-exit status.
 */
export type TaskListLifecycleFilter = "open" | "closed";

/**
 * Fine-grained status filter in the toolbar dropdown. `closed` is not
 * listed here — the Open/Closed tabs own that axis.
 */
export type TaskListClientStatusFilter =
  | "all"
  | "scheduled"
  | Exclude<Status, "closed">;

export type TaskListClientPriorityFilter = "all" | Priority;

/** Max tag chips shown under a list-row title before +N overflow. */
export const TASK_LIST_TAG_CHIP_LIMIT = 3;

/** Sorted unique tags across the given tasks (exact stored strings). */
export function uniqueSortedTagsFromTasks(
  tasks: ReadonlyArray<{ tags?: string[] }>,
): string[] {
  const seen = new Set<string>();
  for (const task of tasks) {
    for (const tag of task.tags ?? []) {
      if (tag) seen.add(tag);
    }
  }
  return [...seen].sort((a, b) => a.localeCompare(b));
}

/** Keep tasks that include `tagFilter`, or all when filter is `"all"`. */
export function filterTasksByTag<T extends { tags?: string[] }>(
  tasks: ReadonlyArray<T>,
  tagFilter: string,
): T[] {
  if (tagFilter === "all") return [...tasks];
  return tasks.filter((t) => (t.tags ?? []).includes(tagFilter));
}

/**
 * Client-side filters for the task list (lifecycle, status, priority,
 * title substring). Open never includes `closed`; Closed is only
 * `status === "closed"` (status dropdown is ignored on that tab).
 */
export function filterTasksForListView(
  tasks: TaskWithDepth[],
  lifecycleFilter: TaskListLifecycleFilter,
  statusFilter: TaskListClientStatusFilter,
  priorityFilter: TaskListClientPriorityFilter,
  titleSearch: string,
  /**
   * Override for `Date.now()` when evaluating the synthetic
   * `scheduled` bucket. Tests pass a fixed clock so the cutoff is
   * deterministic; production callers leave it undefined and we read
   * from `Date.now`.
   */
  nowMs?: number,
): TaskWithDepth[] {
  const q = titleSearch.trim().toLowerCase();
  const now = nowMs ?? Date.now();
  return tasks.filter((t) => {
    if (lifecycleFilter === "closed") {
      if (t.status !== "closed") return false;
    } else if (t.status === "closed") {
      return false;
    } else if (statusFilter === "scheduled") {
      if (t.status !== "ready") return false;
      if (!t.pickup_not_before) return false;
      const ts = Date.parse(t.pickup_not_before);
      if (Number.isNaN(ts)) return false;
      if (ts <= now) return false;
    } else if (statusFilter !== "all" && t.status !== statusFilter) {
      return false;
    }
    if (priorityFilter !== "all" && t.priority !== priorityFilter)
      return false;
    if (q && !t.title.toLowerCase().includes(q)) return false;
    return true;
  });
}
