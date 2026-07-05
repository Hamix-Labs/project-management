import { errorMessage } from "@/lib/errorMessage";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { Task, TaskChecklistResponse } from "@/types";
import type { UseQueryResult } from "@tanstack/react-query";
import {
  TaskCyclesPanel,
  TaskCommitsPanel,
  TaskDetailChecklistSection,
  TaskDetailToolbarActions,
  TaskDetailHeader,
  TaskDetailPromptSection,
  TaskDetailSchedule,
  TaskDependenciesPanel,
  TaskGatePanel,
  TaskModelConfigModal,
} from "../components/task-detail";
import { AutonomyConfirmDialog, TaskRetryConfirmDialog } from "../components/dialogs";
import type { TaskRetryMode } from "../components/dialogs/TaskRetryConfirmDialog";
import { sanitizePromptHtml } from "../task-prompt";
import { canEditTask } from "../task-display/canEditTask";
import { canMutateTaskCriteria } from "../task-display/canMutateTaskCriteria";
import { useTaskDetailChecklist } from "../hooks/useTaskDetailChecklist";
import { useTaskDetailMutations } from "../hooks/useTaskDetailMutations";
import { useTaskDetailScheduling } from "../hooks/useTaskDetailScheduling";
import { resolveTaskDependencySummaries } from "../task-query";
import { useTasksAppModals } from "../app/TasksAppProvider";

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
  const autonomyEnable = autonomyMode === "on_hold";
  const dependenciesUiEnabled = !isUiFeatureOmitted("tagsAndDependencies");
  const releaseGatesUiEnabled = !isUiFeatureOmitted("releaseGates");

  return (
    <section className="panel task-detail-panel task-detail-content--enter">
      <TaskDetailHeader task={task} />

      <div className="task-detail-toolbar">
        <TaskDetailSchedule task={task} />
        <TaskDetailToolbarActions
          saving={saving}
          canEdit={canEditTask(task.status)}
          onEdit={() => modals.openEdit(task)}
          onDelete={() => modals.requestDelete(task)}
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
    </section>
  );
}
