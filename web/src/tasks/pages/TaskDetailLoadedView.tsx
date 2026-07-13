import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { Task, TaskChecklistResponse } from "@/types";
import type { UseQueryResult } from "@tanstack/react-query";
import { TaskDetailHeader } from "../components/task-detail";
import type { TaskRetryMode } from "../components/dialogs/TaskRetryConfirmDialog";
import { sanitizePromptHtml } from "@/lib/promptFormat";
import { useTaskDetailChecklist } from "../checklist/hooks/useTaskDetailChecklist";
import { useTaskDetailMutations } from "../hooks/useTaskDetailMutations";
import { useTaskDetailScheduling } from "../hooks/useTaskDetailScheduling";
import { resolveTaskDependencySummaries } from "../task-query";
import { useTasksAppModals } from "../app/TasksAppProvider";
import { TaskDetailLoadedDialogs } from "./TaskDetailLoadedDialogs";
import { TaskDetailLoadedSections } from "./TaskDetailLoadedSections";
import { TaskDetailLoadedToolbar } from "./TaskDetailLoadedToolbar";

export type AutonomyMode = "hidden" | "ready" | "on_hold";

type TaskDetailChecklistState = ReturnType<typeof useTaskDetailChecklist>;

export function resolveAutonomyMode(taskStatus: Task["status"]): AutonomyMode {
  if (taskStatus === "ready") return "ready";
  if (taskStatus === "on_hold") return "on_hold";
  return "hidden";
}

function countChecklistProgress(items: { done: boolean }[]) {
  const doneCount = items.filter((item) => item.done).length;
  return { doneCount, totalCount: items.length };
}

export type TaskDetailLoadedViewProps = {
  task: Task;
  taskId: string;
  taskQuerySuccess: boolean;
  saving: boolean;
  modals: ReturnType<typeof useTasksAppModals>;
  scheduling: ReturnType<typeof useTaskDetailScheduling>;
  checklistQuery: UseQueryResult<TaskChecklistResponse>;
  checklistState: TaskDetailChecklistState;
  dependencySummaries: ReturnType<typeof resolveTaskDependencySummaries>;
  autonomyMode: AutonomyMode;
  autonomyConfirmOpen: boolean;
  setAutonomyConfirmOpen: (open: boolean) => void;
  autonomyMutation: ReturnType<typeof useTaskDetailMutations>["autonomyMutation"];
  retryConfirmMode: TaskRetryMode | null;
  setRetryConfirmMode: (mode: TaskRetryMode | null) => void;
  retryMutation: ReturnType<typeof useTaskDetailMutations>["retryMutation"];
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
  checklistQuery,
  checklistState,
  dependencySummaries,
  autonomyMode,
  autonomyConfirmOpen,
  setAutonomyConfirmOpen,
  autonomyMutation,
  retryConfirmMode,
  setRetryConfirmMode,
  retryMutation,
  modelConfigOpen,
  setModelConfigOpen,
}: TaskDetailLoadedViewProps) {
  const checklistItems = checklistQuery.data?.items ?? [];
  const { doneCount, totalCount } = countChecklistProgress(checklistItems);
  const sanitizedInitialPrompt = sanitizePromptHtml(task.initial_prompt);
  const dependenciesUiEnabled = !isUiFeatureOmitted("tagsAndDependencies");
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
        setRetryConfirmMode={setRetryConfirmMode}
        setModelConfigOpen={setModelConfigOpen}
        retryMutation={retryMutation}
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
        retryConfirmMode={retryConfirmMode}
        setRetryConfirmMode={setRetryConfirmMode}
        retryMutation={retryMutation}
        modelConfigOpen={modelConfigOpen}
        setModelConfigOpen={setModelConfigOpen}
      />

      <TaskDetailLoadedSections
        task={task}
        taskId={taskId}
        taskQuerySuccess={taskQuerySuccess}
        saving={saving}
        checklistQuery={checklistQuery}
        checklistState={checklistState}
        dependencySummaries={dependencySummaries}
        scheduling={scheduling}
        doneCount={doneCount}
        totalCount={totalCount}
        sanitizedInitialPrompt={sanitizedInitialPrompt}
        dependenciesUiEnabled={dependenciesUiEnabled}
        releaseGatesUiEnabled={releaseGatesUiEnabled}
      />
    </section>
  );
}
