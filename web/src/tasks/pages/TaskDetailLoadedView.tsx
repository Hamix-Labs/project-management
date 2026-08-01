import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { Task } from "@/types";
import { TaskDetailHeader } from "../components/task-detail";
import { sanitizePromptHtml } from "@/lib/promptFormat";
import { useTaskDetailMutations } from "../hooks/useTaskDetailMutations";
import { useTaskDetailScheduling } from "../hooks/useTaskDetailScheduling";
import { resolveTaskDependencySummaries } from "../task-query";
import { useTasksAppModals } from "../app/TasksAppProvider";
import { TaskDetailLoadedDialogs } from "./TaskDetailLoadedDialogs";
import { TaskDetailLoadedSections } from "./TaskDetailLoadedSections";
import { TaskDetailLoadedToolbar } from "./TaskDetailLoadedToolbar";

export type AutonomyMode = "hidden" | "ready" | "on_hold";

export function resolveAutonomyMode(taskStatus: Task["status"]): AutonomyMode {
  if (taskStatus === "ready") return "ready";
  if (taskStatus === "on_hold") return "on_hold";
  return "hidden";
}

export type TaskDetailLoadedViewProps = {
  task: Task;
  taskId: string;
  taskQuerySuccess: boolean;
  saving: boolean;
  modals: ReturnType<typeof useTasksAppModals>;
  scheduling: ReturnType<typeof useTaskDetailScheduling>;
  dependencySummaries: ReturnType<typeof resolveTaskDependencySummaries>;
  autonomyMode: AutonomyMode;
  autonomyConfirmOpen: boolean;
  setAutonomyConfirmOpen: (open: boolean) => void;
  autonomyMutation: ReturnType<typeof useTaskDetailMutations>["autonomyMutation"];
  approveConfirmOpen: boolean;
  setApproveConfirmOpen: (open: boolean) => void;
  approveMutation: ReturnType<typeof useTaskDetailMutations>["approveMutation"];
  openPrConfirmOpen: boolean;
  setOpenPrConfirmOpen: (open: boolean) => void;
  openPrMutation: ReturnType<typeof useTaskDetailMutations>["openPrMutation"];
  polishDialogOpen: boolean;
  setPolishDialogOpen: (open: boolean) => void;
  polishMutation: ReturnType<typeof useTaskDetailMutations>["polishMutation"];
  modelConfigOpen: boolean;
  setModelConfigOpen: (open: boolean) => void;
};

export function TaskDetailLoadedView({
  task,
  taskId,
  taskQuerySuccess,
  saving,
  modals,
  scheduling,
  dependencySummaries,
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
}: TaskDetailLoadedViewProps) {
  const sanitizedInitialPrompt = sanitizePromptHtml(task.initial_prompt);
  const dependenciesUiEnabled = !isUiFeatureOmitted("taskDependencies");
  const releaseGatesUiEnabled = !isUiFeatureOmitted("releaseGates");

  return (
    <section className="panel task-detail-panel task-detail-content--enter">
      <TaskDetailHeader task={task} />

      <TaskDetailLoadedToolbar
        task={task}
        saving={saving}
        modals={modals}
        autonomyMode={autonomyMode}
        setAutonomyConfirmOpen={setAutonomyConfirmOpen}
        setApproveConfirmOpen={setApproveConfirmOpen}
        setOpenPrConfirmOpen={setOpenPrConfirmOpen}
        setPolishDialogOpen={setPolishDialogOpen}
        setModelConfigOpen={setModelConfigOpen}
        approveMutation={approveMutation}
        openPrMutation={openPrMutation}
        polishMutation={polishMutation}
        autonomyMutation={autonomyMutation}
      />

      <TaskDetailLoadedDialogs
        task={task}
        saving={saving}
        modals={modals}
        autonomyMode={autonomyMode}
        autonomyConfirmOpen={autonomyConfirmOpen}
        setAutonomyConfirmOpen={setAutonomyConfirmOpen}
        autonomyMutation={autonomyMutation}
        approveConfirmOpen={approveConfirmOpen}
        setApproveConfirmOpen={setApproveConfirmOpen}
        approveMutation={approveMutation}
        openPrConfirmOpen={openPrConfirmOpen}
        setOpenPrConfirmOpen={setOpenPrConfirmOpen}
        openPrMutation={openPrMutation}
        polishDialogOpen={polishDialogOpen}
        setPolishDialogOpen={setPolishDialogOpen}
        polishMutation={polishMutation}
        modelConfigOpen={modelConfigOpen}
        setModelConfigOpen={setModelConfigOpen}
      />

      <TaskDetailLoadedSections
        task={task}
        taskId={taskId}
        taskQuerySuccess={taskQuerySuccess}
        saving={saving}
        dependencySummaries={dependencySummaries}
        scheduling={scheduling}
        sanitizedInitialPrompt={sanitizedInitialPrompt}
        dependenciesUiEnabled={dependenciesUiEnabled}
        releaseGatesUiEnabled={releaseGatesUiEnabled}
      />
    </section>
  );
}
