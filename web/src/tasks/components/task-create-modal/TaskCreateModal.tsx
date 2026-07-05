import { useRef, useState } from "react";
import type { TestScenario } from "@/tasks/test-scenarios";
import { taskCreateModalBusyLabel } from "./taskCreateModalBusyLabel";
import { resolveTaskCreateModalPresentation } from "./taskCreateModalPresentation";
import type { TaskCreateModalProps } from "./taskCreateModalProps";
import { TaskCreateModalShell } from "./TaskCreateModalShell";

export type { TaskCreateModalProps };

export function TaskCreateModal({
  editingTaskId = null,
  composeTarget = "task",
  composeOperation = "create",
  editingTaskRunner = "",
  composeStatus,
  onComposeStatusChange,
  patchPending = false,
  patchError = null,
  formError = null,
  pending,
  saving,
  draftSaving,
  draftSaveLabel,
  draftSaveError,
  onClose,
  title,
  prompt,
  priority,
  checklistItems,
  onTitleChange,
  onPromptChange,
  onPriorityChange,
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
  onRemoveChecklistRow,
  taskRunner,
  taskCursorModel,
  onTaskRunnerChange,
  onTaskCursorModelChange,
  projectAssignment,
  promptProjectContext,
  schedule,
  onScheduleChange,
  autonomyEnabled,
  onAutonomyChange,
  autonomyDisabled = false,
  tagsCsv,
  milestone,
  projectId,
  worktreeId,
  onWorktreeChange,
  dependsOn,
  onTagsCsvChange,
  onMilestoneChange,
  onDependsOnChange,
  appTimezone,
  onSaveDraft,
  onSubmit,
  createError = null,
  createFormError = null,
  onApplyTestScenario,
}: TaskCreateModalProps) {
  const presentation = resolveTaskCreateModalPresentation({
    editingTaskId,
    composeTarget,
    composeOperation,
    composeStatus,
    pending,
    saving,
    patchPending,
    draftSaveLabel,
    onApplyTestScenario,
  });

  const [scenariosOpen, setScenariosOpen] = useState(false);
  const scenariosTriggerRef = useRef<HTMLButtonElement>(null);

  const handleScenarioPicked = (scenario: TestScenario) => {
    onApplyTestScenario?.(scenario);
    setScenariosOpen(false);
    scenariosTriggerRef.current?.focus();
  };

  return (
    <TaskCreateModalShell
      presentation={presentation}
      editingTaskId={editingTaskId}
      editingTaskRunner={editingTaskRunner}
      composeStatus={composeStatus}
      onComposeStatusChange={onComposeStatusChange}
      pending={pending}
      saving={saving}
      draftSaving={draftSaving}
      draftSaveLabel={draftSaveLabel}
      draftSaveError={draftSaveError}
      onClose={onClose}
      title={title}
      prompt={prompt}
      priority={priority}
      checklistItems={checklistItems}
      onTitleChange={onTitleChange}
      onPromptChange={onPromptChange}
      onPriorityChange={onPriorityChange}
      onAppendChecklistCriterion={onAppendChecklistCriterion}
      onUpdateChecklistRow={onUpdateChecklistRow}
      onRemoveChecklistRow={onRemoveChecklistRow}
      taskRunner={taskRunner}
      taskCursorModel={taskCursorModel}
      onTaskRunnerChange={onTaskRunnerChange}
      onTaskCursorModelChange={onTaskCursorModelChange}
      projectAssignment={projectAssignment}
      promptProjectContext={promptProjectContext}
      schedule={schedule}
      onScheduleChange={onScheduleChange}
      autonomyEnabled={autonomyEnabled}
      onAutonomyChange={onAutonomyChange}
      autonomyDisabled={autonomyDisabled}
      tagsCsv={tagsCsv}
      milestone={milestone}
      projectId={projectId}
      worktreeId={worktreeId}
      onWorktreeChange={onWorktreeChange}
      dependsOn={dependsOn}
      onTagsCsvChange={onTagsCsvChange}
      onMilestoneChange={onMilestoneChange}
      onDependsOnChange={onDependsOnChange}
      appTimezone={appTimezone}
      onSaveDraft={onSaveDraft}
      onSubmit={onSubmit}
      createError={createError}
      createFormError={createFormError}
      patchError={patchError}
      formError={formError}
      busyLabel={taskCreateModalBusyLabel()}
      scenariosOpen={scenariosOpen}
      scenariosTriggerRef={scenariosTriggerRef}
      onToggleScenarios={() => setScenariosOpen((open) => !open)}
      onScenarioPicked={handleScenarioPicked}
      onCloseScenarios={() => setScenariosOpen(false)}
    />
  );
}
