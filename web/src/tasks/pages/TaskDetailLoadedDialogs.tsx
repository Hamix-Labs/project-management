import { useQueryClient } from "@tanstack/react-query";
import { errorMessage } from "@/lib/errorMessage";
import type { TaskChecklistResponse } from "@/types";
import { TaskModelConfigModal } from "../components/task-detail";
import {
  AutonomyConfirmDialog,
  TaskApproveConfirmDialog,
  TaskPolishDialog,
  TaskRetryConfirmDialog,
} from "../components/dialogs";
import { taskQueryKeys } from "../task-query";
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
  const queryClient = useQueryClient();
  const checklist = queryClient.getQueryData<TaskChecklistResponse>(
    taskQueryKeys.checklist(task.id),
  );
  const polishCriteria =
    checklist?.items.map((item) => ({ id: item.id, text: item.text })) ?? [];

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
          worktreeId={task.worktree_id}
          criteria={polishCriteria}
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
          onConfirm={(payload) => polishMutation.mutate(payload)}
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
