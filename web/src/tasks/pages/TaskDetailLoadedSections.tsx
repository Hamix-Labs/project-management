import {
  TaskCommitsPanel,
  TaskCyclesPanel,
  TaskDetailChecklistContainer,
  TaskDetailPromptSection,
  TaskDependenciesPanel,
  TaskGatePanel,
} from "../components/task-detail";
import type { TaskDetailLoadedViewProps } from "./TaskDetailLoadedView";

type TaskDetailLoadedSectionsProps = Pick<
  TaskDetailLoadedViewProps,
  | "task"
  | "taskId"
  | "taskQuerySuccess"
  | "saving"
  | "dependencySummaries"
  | "scheduling"
> & {
  sanitizedInitialPrompt: string;
  dependenciesUiEnabled: boolean;
  releaseGatesUiEnabled: boolean;
};

export function TaskDetailLoadedSections({
  task,
  taskId,
  taskQuerySuccess,
  saving,
  dependencySummaries,
  scheduling,
  sanitizedInitialPrompt,
  dependenciesUiEnabled,
  releaseGatesUiEnabled,
}: TaskDetailLoadedSectionsProps) {
  return (
    <>
      {dependenciesUiEnabled ? (
        <TaskDependenciesPanel dependencies={dependencySummaries} />
      ) : null}

      {releaseGatesUiEnabled ? (
        <TaskGatePanel
          gate={task.gate}
          editable
          onAction={(action) => scheduling.gateMutation.mutate(action)}
          actionPending={scheduling.gateMutation.isPending}
          error={scheduling.gateMutation.error ? scheduling.schedulingError : null}
        />
      ) : null}

      <TaskDetailChecklistContainer
        taskId={taskId}
        saving={saving}
        taskStatus={task.status}
      />

      <TaskDetailPromptSection
        initialPrompt={task.initial_prompt}
        sanitizedInitialPrompt={sanitizedInitialPrompt}
      />

      <TaskCyclesPanel taskId={taskId} enabled={taskQuerySuccess} />

      <TaskCommitsPanel taskId={taskId} enabled={taskQuerySuccess} />
    </>
  );
}
