import type { Status } from "@/types";
import { TaskCreateModalCriteriaSection } from "./TaskCreateModalCriteriaSection";
import { TaskCreateModalEssentialsSection } from "./TaskCreateModalEssentialsSection";
import { TaskCreateModalExecutionSection } from "./TaskCreateModalExecutionSection";
import { TaskCreateModalPromptSection } from "./TaskCreateModalPromptSection";
import type {
  TaskCreateModalCriteria,
  TaskCreateModalEssentials,
  TaskCreateModalExecution,
  TaskCreateModalGitBinding,
  TaskCreateModalPromptFields,
} from "./taskCreateModalProps";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  editingTaskId: string | null;
  editingTemplateId: string | null;
  editingTaskRunner: string;
  onComposeStatusChange?: (status: Status) => void;
  essentials: TaskCreateModalEssentials;
  prompt: TaskCreateModalPromptFields;
  criteria: TaskCreateModalCriteria;
  git: TaskCreateModalGitBinding;
  execution: TaskCreateModalExecution & { autonomyDisabled: boolean };
  appTimezone: string;
};

export function TaskCreateModalFormBody({
  presentation,
  editingTaskId,
  editingTemplateId,
  editingTaskRunner,
  onComposeStatusChange,
  essentials,
  prompt,
  criteria,
  git,
  execution,
  appTimezone,
}: Props) {
  const editorKey = presentation.isTaskEdit
    ? editingTaskId ?? "edit-prompt-modal"
    : presentation.isTemplateMode && presentation.isEdit
      ? editingTemplateId ?? "template-edit-prompt-modal"
      : presentation.isTemplateMode
        ? "template-prompt-modal"
        : "create-prompt-modal";
  const checklistRequirement = presentation.isTaskEdit ? "optional" : "required";

  return (
    <div className="task-create-modal-scroll">
      <TaskCreateModalEssentialsSection
        presentation={presentation}
        title={essentials.title}
        priority={essentials.priority}
        repositoryId={git.repositoryId}
        projectId={git.projectId}
        worktreeId={git.worktreeId}
        onTitleChange={essentials.onTitleChange}
        onPriorityChange={essentials.onPriorityChange}
        onRepositoryChange={git.onRepositoryChange}
        onProjectChange={git.onProjectChange}
        onWorktreeChange={git.onWorktreeChange}
      />

      <TaskCreateModalPromptSection
        presentation={presentation}
        editorKey={editorKey}
        prompt={prompt.prompt}
        worktreeId={git.worktreeId}
        onPromptChange={prompt.onPromptChange}
      />

      <TaskCreateModalCriteriaSection
        presentation={presentation}
        checklistItems={criteria.checklistItems}
        checklistRequirement={checklistRequirement}
        tagsCsv={criteria.tagsCsv}
        functionInputs={criteria.functionInputs}
        onAppendChecklistCriterion={criteria.onAppendChecklistCriterion}
        onUpdateChecklistRow={criteria.onUpdateChecklistRow}
        onRemoveChecklistRow={criteria.onRemoveChecklistRow}
        onTagsCsvChange={criteria.onTagsCsvChange}
        onFunctionInputsChange={criteria.onFunctionInputsChange}
      />

      <TaskCreateModalExecutionSection
        presentation={presentation}
        editingTaskRunner={editingTaskRunner}
        autonomyEnabled={execution.autonomyEnabled}
        autonomyDisabled={execution.autonomyDisabled}
        onAutonomyChange={execution.onAutonomyChange}
        taskRunner={execution.taskRunner}
        taskCursorModel={execution.taskCursorModel}
        onTaskRunnerChange={execution.onTaskRunnerChange}
        onTaskCursorModelChange={execution.onTaskCursorModelChange}
        onComposeStatusChange={onComposeStatusChange}
        schedule={execution.schedule}
        onScheduleChange={execution.onScheduleChange}
        appTimezone={appTimezone}
        tagsCsv={criteria.tagsCsv}
        milestone={execution.milestone}
        projectId={git.projectId}
        dependsOn={execution.dependsOn}
        onTagsCsvChange={criteria.onTagsCsvChange}
        onMilestoneChange={execution.onMilestoneChange}
        onDependsOnChange={execution.onDependsOnChange}
      />
    </div>
  );
}
