import { errorMessage } from "@/lib/errorMessage";
import { TaskModelConfigModal } from "../components/task-detail";
import { AutonomyConfirmDialog, TaskRetryConfirmDialog } from "../components/dialogs";
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
