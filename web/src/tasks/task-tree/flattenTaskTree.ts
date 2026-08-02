import type { Task } from "@/types";

/** Flat row for list display. */
export type TaskWithDepth = Task & { depth: number };

/**
 * Maps tasks to list rows. When a worktree family filter is active, root
 * (`id === worktree_root_task_id`) is depth 0 and siblings are depth 1.
 * Otherwise every row stays depth 0 (flat list).
 */
export function flattenTaskTreeRoots(
  nodes: Task[],
  opts?: { worktreeFamilyActive?: boolean },
): TaskWithDepth[] {
  const familyActive = opts?.worktreeFamilyActive === true;
  return nodes.map((n) => {
    if (!familyActive) {
      return { ...n, depth: 0 };
    }
    const rootID = n.worktree_root_task_id?.trim();
    const isRoot = rootID !== undefined && rootID !== "" && n.id === rootID;
    return { ...n, depth: isRoot ? 0 : 1 };
  });
}

/** Sort worktree-family rows: root first, then children by created_at desc. */
export function sortWorktreeFamilyTasks<T extends Task>(tasks: T[]): T[] {
  return [...tasks].sort((a, b) => {
    const rootA = a.worktree_root_task_id?.trim();
    const rootB = b.worktree_root_task_id?.trim();
    const aRoot = rootA !== undefined && rootA !== "" && a.id === rootA;
    const bRoot = rootB !== undefined && rootB !== "" && b.id === rootB;
    if (aRoot !== bRoot) return aRoot ? -1 : 1;
    const ca = a.created_at ?? "";
    const cb = b.created_at ?? "";
    if (ca !== cb) return cb.localeCompare(ca);
    return a.id.localeCompare(b.id);
  });
}
