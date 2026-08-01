import type { Task } from "@/types";

/**
 * Whether agents are waiting on a person soon, from task status and server `approval_pending` on events.
 * Status cases that set `show: true` match `statusNeedsUserInput` in `task-display/taskStatusNeedsUser.ts`.
 */
export function userAttention(
  task: Task,
  meta: { approvalPending: boolean; failureSummary?: string },
): {
  show: boolean;
  headline: string;
  body: string;
} {
  if (meta.approvalPending) {
    return {
      show: true,
      headline: "Approval requested",
      body: "The agent is asking for approval on this task. Review the timeline below.",
    };
  }
  switch (task.status) {
    case "review":
      return {
        show: true,
        headline: "Ready for your review",
        body: "Approve and open a pull request, or send polish instructions.",
      };
    case "pr_ready":
      return {
        show: true,
        headline: "Pull request ready",
        body: "A PR is open for this work. Mark the task done when you are ready for dependents to proceed.",
      };
    case "blocked":
      return {
        show: true,
        headline: "Blocked",
        body: "The agent is blocked. Review context and unblock or adjust the task.",
      };
    case "failed": {
      const summary = (meta.failureSummary ?? "").trim();
      return {
        show: true,
        headline: "Task failed",
        body: summary
          ? summary
          : "The agent reported a failure. Review what happened and decide whether to retry or change scope.",
      };
    }
    default:
      return { show: false, headline: "", body: "" };
  }
}
