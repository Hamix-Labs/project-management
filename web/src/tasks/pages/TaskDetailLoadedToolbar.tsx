import { StatusBadge } from "@/components/task-status";
import {
  TaskDetailSchedule,
  TaskDetailToolbarActions,
} from "../components/task-detail";
import { TokenUsageChip } from "../components/task-detail/TokenUsageChip";
import { canEditTask } from "../task-display/canEditTask";
import { statusNeedsUserInput } from "../task-display";
import type { TaskDetailLoadedViewProps } from "./TaskDetailLoadedView";

type TaskDetailLoadedToolbarProps = Pick<
  TaskDetailLoadedViewProps,
  | "task"
  | "saving"
  | "modals"
  | "autonomyMode"
  | "setAutonomyConfirmOpen"
  | "setRetryConfirmMode"
  | "setApproveConfirmOpen"
  | "setPolishDialogOpen"
  | "setModelConfigOpen"
  | "retryMutation"
  | "approveMutation"
  | "polishMutation"
  | "autonomyMutation"
>;

export function TaskDetailLoadedToolbar({
  task,
  saving,
  modals,
  autonomyMode,
  setAutonomyConfirmOpen,
  setRetryConfirmMode,
  setApproveConfirmOpen,
  setPolishDialogOpen,
  setModelConfigOpen,
  retryMutation,
  approveMutation,
  polishMutation,
  autonomyMutation,
}: TaskDetailLoadedToolbarProps) {
  const inReview = task.status === "review";
  const isClosed = task.status === "closed";
  const needsUser = statusNeedsUserInput(task.status);

  return (
    <div className="task-detail-toolbar">
      <div className="task-detail-toolbar-left">
        <div
          className="task-detail-execution-bar"
          data-testid="task-detail-execution-bar"
        >
          <TokenUsageChip taskId={task.id} />
          <div
            className="task-detail-execution-bar__status"
            data-testid="task-detail-status"
          >
            <StatusBadge
              status={task.status}
              className="task-detail-status-badge"
              data-needs-user={needsUser ? "true" : undefined}
            />
          </div>
        </div>
        <TaskDetailSchedule task={task} />
      </div>
      <TaskDetailToolbarActions
        saving={saving}
        canEdit={canEditTask(task.status) && !isClosed}
        onEdit={() => modals.openEdit(task)}
        onClose={
          isClosed
            ? undefined
            : () =>
                modals.requestClose({
                  id: task.id,
                  title: task.title,
                  number: task.number,
                })
        }
        closePending={modals.closePending}
        onReopen={isClosed ? () => modals.reopen(task.id) : undefined}
        reopenPending={modals.reopenPending}
        onApprove={inReview ? () => setApproveConfirmOpen(true) : undefined}
        approvePending={approveMutation.isPending}
        onPolish={inReview ? () => setPolishDialogOpen(true) : undefined}
        polishPending={polishMutation.isPending}
        onRetryFresh={
          task.status === "failed"
            ? () => setRetryConfirmMode("fresh")
            : undefined
        }
        onRetryResume={
          task.status === "failed"
            ? () => setRetryConfirmMode("resume")
            : undefined
        }
        retryPending={retryMutation.isPending}
        onConfigureModel={() => setModelConfigOpen(true)}
        showModelConfig={task.status === "failed"}
        autonomyMode={autonomyMode}
        onToggleAutonomy={
          autonomyMode !== "hidden"
            ? () => setAutonomyConfirmOpen(true)
            : undefined
        }
        autonomyPending={autonomyMutation.isPending}
      />
    </div>
  );
}
