import { errorMessage } from "@/lib/errorMessage";
import { TaskModelConfigModal } from "../components/task-detail";
import {
  AutonomyConfirmDialog,
  TaskApproveConfirmDialog,
  TaskPolishDialog,
  TaskRetryConfirmDialog,
} from "../components/dialogs";
import type { TaskDetailLoadedViewProps } from "./TaskDetailLoadedView";

type TaskDetailLoadedDialogsProps = Pick<
  TaskDetailLoadedViewProps,
  | "task"
  | "saving"
  | "modals"
  | "autonomyMode"
  | "autonomyConfirmOpen"
  | "setAutonomyConfirmOpen"
  | "autonomyMutation"
  | "retryConfirmMode"
  | "setRetryConfirmMode"
  | "retryMutation"
  | "approveConfirmOpen"
  | "setApproveConfirmOpen"
  | "approveMutation"
  | "polishDialogOpen"
  | "setPolishDialogOpen"
  | "polishMutation"
  | "modelConfigOpen"
  | "setModelConfigOpen"
>;

export function TaskDetailLoadedDialogs({
  task,
  saving,
  modals,
  autonomyMode,
  autonomyConfirmOpen,
  setAutonomyConfirmOpen,
  autonomyMutation,
  retryConfirmMode,
  setRetryConfirmMode,
  retryMutation,
  approveConfirmOpen,
  setApproveConfirmOpen,
  approveMutation,
  polishDialogOpen,
  setPolishDialogOpen,
  polishMutation,
  modelConfigOpen,
  setModelConfigOpen,
}: TaskDetailLoadedDialogsProps) {
  const autonomyEnable = autonomyMode === "on_hold";

  return (
    <>
      {autonomyConfirmOpen && autonomyMode !== "hidden" ? (
        <AutonomyConfirmDialog
          enable={autonomyEnable}
          taskTitle={task.title}
          saving={saving}
          pending={autonomyMutation.isPending}
          error={
            autonomyMutation.isError
              ? errorMessage(
                  autonomyMutation.error,
                  autonomyEnable
                    ? "Couldn't resume autonomous execution."
                    : "Couldn't put this task on hold.",
                )
              : null
          }
          onCancel={() => {
            setAutonomyConfirmOpen(false);
            if (autonomyMutation.isError) autonomyMutation.reset();
          }}
          onConfirm={() =>
            autonomyMutation.mutate(autonomyEnable ? "ready" : "on_hold")
          }
        />
      ) : null}

      {retryConfirmMode ? (
        <TaskRetryConfirmDialog
          mode={retryConfirmMode}
          taskTitle={task.title}
          saving={saving}
          pending={retryMutation.isPending}
          error={
            retryMutation.isError
              ? errorMessage(
                  retryMutation.error,
                  retryConfirmMode === "fresh"
                    ? "Couldn't start over."
                    : "Couldn't resume from failure.",
                )
              : null
          }
          onCancel={() => {
            setRetryConfirmMode(null);
            if (retryMutation.isError) retryMutation.reset();
          }}
          onConfirm={() => retryMutation.mutate(retryConfirmMode)}
        />
      ) : null}

      {approveConfirmOpen ? (
        <TaskApproveConfirmDialog
          taskTitle={task.title}
          saving={saving}
          pending={approveMutation.isPending}
          error={
            approveMutation.isError
              ? errorMessage(approveMutation.error, "Couldn't approve task.")
              : null
          }
          onCancel={() => {
            setApproveConfirmOpen(false);
            if (approveMutation.isError) approveMutation.reset();
          }}
          onConfirm={() => approveMutation.mutate()}
        />
      ) : null}

      {polishDialogOpen ? (
        <TaskPolishDialog
          taskTitle={task.title}
          worktreeId={task.worktree_id}
          projectId={task.project_id}
          projectContextItemIds={task.project_context_item_ids}
          saving={saving}
          pending={polishMutation.isPending}
          error={
            polishMutation.isError
              ? errorMessage(polishMutation.error, "Couldn't queue polish.")
              : null
          }
          onCancel={() => {
            setPolishDialogOpen(false);
            if (polishMutation.isError) polishMutation.reset();
          }}
          onConfirm={(instructions) => polishMutation.mutate(instructions)}
        />
      ) : null}

      {modelConfigOpen ? (
        <TaskModelConfigModal
          taskTitle={task.title}
          saving={saving}
          onChangeModel={() => modals.openChangeModel(task)}
          onClose={() => setModelConfigOpen(false)}
        />
      ) : null}
    </>
  );
}
