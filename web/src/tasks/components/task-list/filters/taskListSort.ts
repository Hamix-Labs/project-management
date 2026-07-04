import type { TaskWithDepth } from "../../../task-tree";
import { PRIORITY_META } from "../../../task-display/priorityMeta";
import { STATUS_META } from "../../../task-display/statusMeta";

export type TaskListSortKey =
  | "title"
  | "status"
  | "priority"
  | "created_at"
  | "project";

export type TaskListSortDir = "asc" | "desc";

/** Newest-first by `created_at`, then id descending as a stable tie-break. */
export function sortTasksByCreatedDesc<T extends { id: string; created_at?: string }>(
  tasks: T[],
): T[] {
  return [...tasks].sort((a, b) => {
    const aMs = a.created_at ? Date.parse(a.created_at) : 0;
    const bMs = b.created_at ? Date.parse(b.created_at) : 0;
    if (bMs !== aMs) return bMs - aMs;
    return b.id.localeCompare(a.id);
  });
}

/** Client-side sort on the current page; preserves relative order for tree siblings when keys tie. */
export function sortTasksForListView(
  tasks: TaskWithDepth[],
  sortKey: TaskListSortKey,
  sortDir: TaskListSortDir,
  projectNameById: Record<string, string>,
): TaskWithDepth[] {
  const dir = sortDir === "asc" ? 1 : -1;
  return [...tasks].sort((a, b) => {
    let cmp = 0;
    switch (sortKey) {
      case "title":
        cmp = a.title.localeCompare(b.title);
        break;
      case "status":
        cmp = STATUS_META[a.status].order - STATUS_META[b.status].order;
        break;
      case "priority":
        cmp = PRIORITY_META[a.priority].weight - PRIORITY_META[b.priority].weight;
        break;
      case "project": {
        const aLabel = a.project_id ? (projectNameById[a.project_id] ?? "") : "";
        const bLabel = b.project_id ? (projectNameById[b.project_id] ?? "") : "";
        cmp = aLabel.localeCompare(bLabel);
        break;
      }
      case "created_at":
      default: {
        const aMs = a.created_at ? Date.parse(a.created_at) : 0;
        const bMs = b.created_at ? Date.parse(b.created_at) : 0;
        cmp = aMs - bMs;
        break;
      }
    }
    if (cmp !== 0) return cmp * dir;
    return a.id.localeCompare(b.id) * dir;
  });
}
