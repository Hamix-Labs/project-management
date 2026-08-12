import type { useTasksApp } from "../hooks/useTasksApp";
import { buildTaskCreateModalProps } from "../components/task-create-modal/buildTaskCreateModalProps";
import { resolveTaskCreateModalPresentation } from "../components/task-create-modal/taskCreateModalPresentation";

type App = ReturnType<typeof useTasksApp>;

/** Builds presentation + flat props for {@link TaskComposeForm}. */
export function buildTaskComposeFormBundle(
  app: App,
  opts: { leave: () => void; appTimezone: string },
) {
  const isEditing = app.editingTaskId != null;
  const isTemplateMode = app.composeTarget === "template";
  const isTemplateEdit = isTemplateMode && app.composeOperation === "edit";

  const presentation = resolveTaskCreateModalPresentation({
    editingTaskId: app.editingTaskId,
    composeTarget: app.composeTarget,
    composeOperation: app.composeOperation,
    composeStatus: app.composeStatus,
    pending: isTemplateMode ? app.templateSavePending : app.createPending,
    saving: app.saving,
    patchPending: app.patchPending,
    draftSaveLabel: isEditing || isTemplateMode ? null : app.draftSaveLabel,
    onApplyTestScenario:
      isEditing || isTemplateEdit ? undefined : app.applyTestScenario,
  });

  const props = buildTaskCreateModalProps({
    editingTaskId: app.editingTaskId,
    editingTemplateId: app.editingTemplateId,
    composeTarget: app.composeTarget,
    composeOperation: app.composeOperation,
    editingTaskRunner: app.editingTaskRunner,
    composeStatus: app.composeStatus,
    onComposeStatusChange: app.setComposeStatus,
    patchPending: app.patchPending,
    patchError: app.patchError,
    formError: app.editFormError,
    pending: isTemplateMode ? app.templateSavePending : app.createPending,
    saving: app.saving,
    draftSaving: isEditing || isTemplateMode ? false : app.draftSavePending,
    draftSaveLabel: isEditing || isTemplateMode ? null : app.draftSaveLabel,
    draftSaveError: isEditing || isTemplateMode ? false : app.draftSaveError,
    onClose: opts.leave,
    title: app.newTitle,
    prompt: app.newPrompt,
    priority: app.newPriority,
    checklistItems: app.newChecklistItems,
    tagsCsv: app.newTagsCsv,
    functionInputs: app.newFunctionInputs,
    onTitleChange: app.setNewTitle,
    onPromptChange: app.setNewPrompt,
    onPriorityChange: app.setNewPriority,
    onAppendChecklistCriterion: app.appendNewChecklistCriterion,
    onUpdateChecklistRow: app.updateNewChecklistRow,
    onRemoveChecklistRow: app.removeNewChecklistRow,
    onFunctionInputsChange: app.setNewFunctionInputs,
    taskRunner: isEditing ? app.editingTaskRunner : app.newTaskRunner,
    taskCursorModel: app.newTaskCursorModel,
    onTaskRunnerChange: app.setNewTaskRunner,
    onTaskCursorModelChange: app.setNewTaskCursorModel,
    schedule: app.newSchedule,
    onScheduleChange: app.setNewSchedule,
    autonomyEnabled: isEditing
      ? app.composeStatus === "ready"
      : app.newAutonomyEnabled,
    onAutonomyChange: app.setNewAutonomyEnabled,
    autonomyDisabled: isEditing,
    milestone: app.newMilestone,
    repositoryId: app.newRepositoryID,
    projectId: app.newProjectID,
    worktreeId: app.newWorktreeID,
    assignmentLocked: app.createModalAssignmentLocked,
    onRepositoryChange: (repositoryId) => {
      app.setNewRepositoryID(repositoryId);
      app.setNewProjectID("");
      app.setNewWorktreeID("");
    },
    onProjectChange: (projectId) => {
      app.setNewProjectID(projectId);
    },
    onWorktreeChange: app.setNewWorktreeID,
    dependsOn: app.newDependsOn,
    onTagsCsvChange: app.setNewTagsCsv,
    onMilestoneChange: app.setNewMilestone,
    onDependsOnChange: app.setNewDependsOn,
    appTimezone: opts.appTimezone,
    onSaveDraft: () => {
      if (!isEditing) void app.saveDraftNow();
    },
    onSubmit: (e) => {
      void app.submitComposeModal(e);
    },
    createError: isEditing
      ? null
      : isTemplateMode
        ? app.templateSaveError
        : app.createError,
    createFormError: isEditing ? null : app.createFormError,
    onApplyTestScenario:
      isEditing || isTemplateEdit ? undefined : app.applyTestScenario,
  });

  return { presentation, props };
}
