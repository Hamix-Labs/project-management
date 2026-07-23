import {
  TaskDetailSchedule,
  TaskDetailToolbarActions,
} from "../components/task-detail";
import { canEditTask } from "../task-display/canEditTask";
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

  return (
    <div className="task-detail-toolbar">
      <TaskDetailSchedule task={task} />
      <TaskDetailToolbarActions
        saving={saving}
        canEdit={canEditTask(task.status)}
        onEdit={() => modals.openEdit(task)}
        onDelete={() => modals.requestDelete(task)}
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
