import type { Phase, TaskEvent, TaskEventType } from "@/types";
import { phaseLabel } from "@/tasks/cycleDisplay/cyclesViewModel";

const LABELS: Record<TaskEventType, string> = {
  task_created: "Task created",
  status_changed: "Status changed",
  priority_changed: "Priority changed",
  prompt_appended: "Prompt updated",
  context_added: "Context added",
  constraint_added: "Constraint added",
  success_criterion_added: "Success criterion added",
  non_goal_added: "Non-goal added",
  plan_added: "Plan added",
  checklist_item_added: "Checklist item added",
  checklist_item_toggled: "Checklist item updated",
  checklist_item_updated: "Criterion text updated",
  checklist_item_removed: "Criterion removed",
  message_added: "Title or message updated",
  artifact_added: "Artifact added",
  approval_requested: "Approval requested",
  approval_granted: "Approval granted",
  task_completed: "Task completed",
  on_task_done: "Task marked done",
  task_failed: "Task failed",
  task_retry_requested: "Retry requested",
  task_polish_requested: "Polish requested",
  open_pr_requested: "Open PR requested",
  pr_opened: "Pull request opened",
  task_pickup_failed: "Agent pickup failed",
  cycle_started: "Attempt started",
  cycle_completed: "Attempt completed",
  cycle_failed: "Attempt failed",
  phase_started: "Phase started",
  phase_completed: "Phase completed",
  phase_failed: "Phase failed",
  phase_skipped: "Phase skipped",
  sync_ping: "Live sync check (legacy dev ping)",
};

const PHASE_EVENT_ACTION: Partial<Record<TaskEventType, string>> = {
  phase_started: "started",
  phase_completed: "completed",
  phase_failed: "failed",
  phase_skipped: "skipped",
};

function isPhase(value: string): value is Phase {
  return (
    value === "execute" ||
    value === "verify" ||
    value === "diagnose" ||
    value === "persist"
  );
}

export function eventTypeLabel(type: TaskEventType): string {
  return LABELS[type] ?? type;
}

/**
 * Human label for compact timelines. Phase mirror events include the phase
 * kind (Execute, Verify) so rows are scannable without opening detail.
 */
export function eventDisplayLabel(ev: TaskEvent): string {
  const action = PHASE_EVENT_ACTION[ev.type];
  if (action) {
    switch (ev.type) {
      case "phase_started":
      case "phase_completed":
      case "phase_failed":
      case "phase_skipped": {
        const phase = ev.data.phase;
        if (typeof phase === "string" && isPhase(phase)) {
          return `${phaseLabel(phase)} ${action}`;
        }
        break;
      }
      default:
        break;
    }
  }
  return eventTypeLabel(ev.type);
}
