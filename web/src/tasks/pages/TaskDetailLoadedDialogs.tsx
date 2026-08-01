import { useQueryClient } from "@tanstack/react-query";
import { errorMessage } from "@/lib/errorMessage";
import type { TaskChecklistResponse } from "@/types";
import { TaskModelConfigModal } from "../components/task-detail";
import {
  AutonomyConfirmDialog,
  TaskApproveConfirmDialog,
  TaskOpenPRConfirmDialog,
  TaskPolishDialog,
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
  | "approveConfirmOpen"
  | "setApproveConfirmOpen"
  | "approveMutation"
  | "openPrConfirmOpen"
  | "setOpenPrConfirmOpen"
  | "openPrMutation"
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
  approveConfirmOpen,
  setApproveConfirmOpen,
  approveMutation,
  openPrConfirmOpen,
  setOpenPrConfirmOpen,
  openPrMutation,
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
                    : "Couldn't pause this task.",
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

      {openPrConfirmOpen ? (
        <TaskOpenPRConfirmDialog
          taskTitle={task.title}
          saving={saving}
          pending={openPrMutation.isPending}
          error={
            openPrMutation.isError
              ? errorMessage(
                  openPrMutation.error,
                  "Couldn't queue approve and open PR.",
                )
              : null
          }
          onCancel={() => {
            setOpenPrConfirmOpen(false);
            if (openPrMutation.isError) openPrMutation.reset();
          }}
          onConfirm={() => openPrMutation.mutate()}
        />
      ) : null}

      {approveConfirmOpen ? (
        <TaskApproveConfirmDialog
          taskTitle={task.title}
          saving={saving}
          pending={approveMutation.isPending}
          error={
            approveMutation.isError
              ? errorMessage(approveMutation.error, "Couldn't mark task done.")
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
