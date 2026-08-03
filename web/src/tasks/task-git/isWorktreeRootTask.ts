import type { Task } from "@/types/taskCore";

/** True when this task owns the worktree's stack (allocate-time root layer). */
export function isWorktreeRootTask(
  task: Pick<Task, "id" | "worktree_root_task_id">,
): boolean {
  const root = task.worktree_root_task_id?.trim();
  if (!root) {
    // No computed root yet — treat as root so Open PR stays available for
    // single-task worktrees while provisioning completes enrichment.
    return true;
  }
  return root === task.id;
}
