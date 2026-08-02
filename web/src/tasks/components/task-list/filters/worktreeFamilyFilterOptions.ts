import type { Task } from "@/types";
import { taskDisplayRef } from "@/lib/taskShortId";

/** Unique worktree families for the Worktree family filter. */
export function worktreeFamilyFilterOptions(
  tasks: ReadonlyArray<Task>,
): Array<{ value: string; label: string }> {
  const byWt = new Map<string, Task>();
  for (const t of tasks) {
    const wt = t.worktree_id?.trim();
    if (!wt) continue;
    const existing = byWt.get(wt);
    if (!existing) {
      byWt.set(wt, t);
      continue;
    }
    const root = t.worktree_root_task_id?.trim();
    if (root && t.id === root) {
      byWt.set(wt, t);
    }
  }
  return [...byWt.entries()]
    .map(([wt, t]) => ({
      value: wt,
      label: `${taskDisplayRef(t)} ${t.title}`.trim(),
    }))
    .sort((a, b) => a.label.localeCompare(b.label));
}
