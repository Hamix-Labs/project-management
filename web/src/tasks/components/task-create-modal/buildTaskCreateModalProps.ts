import type {
  TaskCreateModalFlatInput,
  TaskCreateModalProps,
} from "./taskCreateModalProps";

/** Groups a flat create-modal field bag into section view-models. */
export function buildTaskCreateModalProps(
  input: TaskCreateModalFlatInput,
): TaskCreateModalProps {
  return {
    session: {
      editingTaskId: input.editingTaskId,
      composeTarget: input.composeTarget,
      composeOperation: input.composeOperation,
      editingTaskRunner: input.editingTaskRunner,
      composeStatus: input.composeStatus,
      onComposeStatusChange: input.onComposeStatusChange,
      patchPending: input.patchPending,
      patchError: input.patchError,
      formError: input.formError,
      pending: input.pending,
      saving: input.saving,
      draftSaving: input.draftSaving,
      draftSaveLabel: input.draftSaveLabel,
      draftSaveError: input.draftSaveError,
      createError: input.createError,
      createFormError: input.createFormError,
    },
    essentials: {
      title: input.title,
      priority: input.priority,
      onTitleChange: input.onTitleChange,
      onPriorityChange: input.onPriorityChange,
    },
    prompt: {
      prompt: input.prompt,
      onPromptChange: input.onPromptChange,
      promptProjectContext: input.promptProjectContext,
    },
    criteria: {
      checklistItems: input.checklistItems,
      onAppendChecklistCriterion: input.onAppendChecklistCriterion,
      onUpdateChecklistRow: input.onUpdateChecklistRow,
      onRemoveChecklistRow: input.onRemoveChecklistRow,
      tagsCsv: input.tagsCsv,
      onTagsCsvChange: input.onTagsCsvChange,
    },
    git: {
      repositoryId: input.repositoryId,
      projectId: input.projectId,
      worktreeId: input.worktreeId,
      onRepositoryChange: input.onRepositoryChange,
      onProjectChange: input.onProjectChange,
      onWorktreeChange: input.onWorktreeChange,
      onProjectContextClear: input.onProjectContextClear,
    },
    execution: {
      taskRunner: input.taskRunner,
      taskCursorModel: input.taskCursorModel,
      taskVerifyChatMode: input.taskVerifyChatMode,
      onTaskRunnerChange: input.onTaskRunnerChange,
      onTaskCursorModelChange: input.onTaskCursorModelChange,
      onTaskVerifyChatModeChange: input.onTaskVerifyChatModeChange,
      schedule: input.schedule,
      onScheduleChange: input.onScheduleChange,
      autonomyEnabled: input.autonomyEnabled,
      onAutonomyChange: input.onAutonomyChange,
      autonomyDisabled: input.autonomyDisabled,
      milestone: input.milestone,
      onMilestoneChange: input.onMilestoneChange,
      dependsOn: input.dependsOn,
      onDependsOnChange: input.onDependsOnChange,
    },
    actions: {
      onClose: input.onClose,
      onSaveDraft: input.onSaveDraft,
      onSubmit: input.onSubmit,
      onApplyTestScenario: input.onApplyTestScenario,
    },
    projectAssignment: input.projectAssignment,
    appTimezone: input.appTimezone,
  };
}
