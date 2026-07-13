import {
  TaskCommitsPanel,
  TaskCyclesPanel,
  TaskDetailChecklistSection,
  TaskDetailPromptSection,
  TaskDependenciesPanel,
  TaskGatePanel,
} from "../components/task-detail";
import { canMutateTaskCriteria } from "../task-display/canMutateTaskCriteria";
import type { TaskDetailLoadedViewProps } from "./TaskDetailLoadedView";

type TaskDetailLoadedSectionsProps = Pick<
  TaskDetailLoadedViewProps,
  | "task"
  | "taskId"
  | "taskQuerySuccess"
  | "saving"
  | "checklistQuery"
  | "checklistState"
  | "dependencySummaries"
  | "scheduling"
> & {
  doneCount: number;
  totalCount: number;
  sanitizedInitialPrompt: string;
  dependenciesUiEnabled: boolean;
  releaseGatesUiEnabled: boolean;
};

export function TaskDetailLoadedSections({
  task,
  taskId,
  taskQuerySuccess,
  saving,
  checklistQuery,
  checklistState,
  dependencySummaries,
  scheduling,
  doneCount,
  totalCount,
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

      <TaskDetailChecklistSection
        saving={saving}
        canAddCriterion={canMutateTaskCriteria(task.status)}
        taskStatus={task.status}
        checklistQuery={checklistQuery}
        doneCount={doneCount}
        totalCount={totalCount}
        modalOpen={checklistState.checklistModalOpen}
        newCriterionText={checklistState.newChecklistText}
        onNewCriterionTextChange={checklistState.setNewChecklistText}
        newCriterionVerifyCommands={checklistState.newChecklistVerifyCommands}
        onNewCriterionVerifyCommandsChange={checklistState.setNewChecklistVerifyCommands}
        onOpenAddModal={checklistState.openChecklistModal}
        onCloseAddModal={checklistState.closeChecklistModal}
        onSubmitNewCriterion={checklistState.submitNewChecklistCriterion}
        addCriterionPending={checklistState.addChecklistMutation.isPending}
        editModalOpen={checklistState.editCriterionModalOpen}
        editingItemId={checklistState.editingChecklistItemId}
        editCriterionText={checklistState.editChecklistText}
        onEditCriterionTextChange={checklistState.setEditChecklistText}
        editCriterionVerifyCommands={checklistState.editChecklistVerifyCommands}
        onEditCriterionVerifyCommandsChange={checklistState.setEditChecklistVerifyCommands}
        onOpenEditCriterionModal={checklistState.openEditCriterionModal}
        onCloseEditCriterionModal={checklistState.closeEditCriterionModal}
        onSubmitEditCriterion={checklistState.submitEditChecklistCriterion}
        editCriterionPending={checklistState.updateChecklistTextMutation.isPending}
        onRemoveChecklistItem={(id) => checklistState.deleteChecklistMutation.mutate(id)}
        removeItemPending={checklistState.deleteChecklistMutation.isPending}
        addCriterionError={checklistState.addChecklistMutation.error}
        editCriterionError={checklistState.updateChecklistTextMutation.error}
        removeItemError={checklistState.deleteChecklistMutation.error}
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
