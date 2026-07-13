import type { ReactNode } from "react";
import type { ChecklistItemDraft, PriorityChoice, Status } from "@/types";
import type { RichPromptEditorProjectContextProps } from "@/components/rich-prompt";
import { TaskCreateModalCriteriaSection } from "./TaskCreateModalCriteriaSection";
import { TaskCreateModalEssentialsSection } from "./TaskCreateModalEssentialsSection";
import { TaskCreateModalExecutionSection } from "./TaskCreateModalExecutionSection";
import { TaskCreateModalProjectSection } from "./TaskCreateModalProjectSection";
import { TaskCreateModalPromptSection } from "./TaskCreateModalPromptSection";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  editingTaskId: string | null;
  editingTaskRunner: string;
  title: string;
  prompt: string;
  priority: PriorityChoice;
  checklistItems: ChecklistItemDraft[];
  onTitleChange: (v: string) => void;
  onPromptChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
  onAppendChecklistCriterion: (item: ChecklistItemDraft | string) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  promptProjectContext?: RichPromptEditorProjectContextProps;
  projectAssignment?: ReactNode;
  taskRunner: string;
  taskCursorModel: string;
  onTaskRunnerChange: (runner: string) => void;
  onTaskCursorModelChange: (v: string) => void;
  onComposeStatusChange?: (status: Status) => void;
  autonomyEnabled: boolean;
  autonomyDisabled: boolean;
  onAutonomyChange: (enabled: boolean) => void;
  schedule: string | null;
  onScheduleChange: (next: string | null) => void;
  appTimezone: string;
  tagsCsv: string;
  milestone: string;
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  onRepositoryChange: (repositoryId: string) => void;
  onProjectChange: (projectId: string) => void;
  onWorktreeChange: (worktreeId: string) => void;
  onProjectContextClear: () => void;
  dependsOn: string[];
  onTagsCsvChange: (value: string) => void;
  onMilestoneChange: (value: string) => void;
  onDependsOnChange: (value: string[]) => void;
};

export function TaskCreateModalFormBody(props: Props) {
  const {
    presentation,
    editingTaskId,
    editingTaskRunner,
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
    promptProjectContext,
    projectAssignment,
    taskRunner,
    taskCursorModel,
    onTaskRunnerChange,
    onTaskCursorModelChange,
    onComposeStatusChange,
    autonomyEnabled,
    autonomyDisabled,
    onAutonomyChange,
    schedule,
    onScheduleChange,
    appTimezone,
    tagsCsv,
    milestone,
    repositoryId,
    projectId,
    worktreeId,
    onRepositoryChange,
    onProjectChange,
    onWorktreeChange,
    onProjectContextClear,
    dependsOn,
    onTagsCsvChange,
    onMilestoneChange,
    onDependsOnChange,
  } = props;

  const editorKey = presentation.isTaskEdit
    ? editingTaskId ?? "edit-prompt-modal"
    : presentation.isTemplateMode
      ? "template-prompt-modal"
      : "create-prompt-modal";
  const checklistRequirement = presentation.isTaskEdit ? "optional" : "required";

  return (
    <div className="task-create-modal-scroll">
      <TaskCreateModalEssentialsSection
        presentation={presentation}
        title={title}
        priority={priority}
        repositoryId={repositoryId}
        projectId={projectId}
        worktreeId={worktreeId}
        onTitleChange={onTitleChange}
        onPriorityChange={onPriorityChange}
        onRepositoryChange={onRepositoryChange}
        onProjectChange={onProjectChange}
        onWorktreeChange={onWorktreeChange}
        onProjectContextClear={onProjectContextClear}
      />

      <TaskCreateModalPromptSection
        presentation={presentation}
        editorKey={editorKey}
        prompt={prompt}
        worktreeId={worktreeId}
        onPromptChange={onPromptChange}
        promptProjectContext={promptProjectContext}
      />

      <TaskCreateModalCriteriaSection
        presentation={presentation}
        checklistItems={checklistItems}
        checklistRequirement={checklistRequirement}
        tagsCsv={tagsCsv}
        onAppendChecklistCriterion={onAppendChecklistCriterion}
        onUpdateChecklistRow={onUpdateChecklistRow}
        onRemoveChecklistRow={onRemoveChecklistRow}
        onTagsCsvChange={onTagsCsvChange}
      />

      {projectAssignment ? (
        <TaskCreateModalProjectSection projectAssignment={projectAssignment} />
      ) : null}

      <TaskCreateModalExecutionSection
        presentation={presentation}
        editingTaskRunner={editingTaskRunner}
        autonomyEnabled={autonomyEnabled}
        autonomyDisabled={autonomyDisabled}
        onAutonomyChange={onAutonomyChange}
        taskRunner={taskRunner}
        taskCursorModel={taskCursorModel}
        onTaskRunnerChange={onTaskRunnerChange}
        onTaskCursorModelChange={onTaskCursorModelChange}
        onComposeStatusChange={onComposeStatusChange}
        schedule={schedule}
        onScheduleChange={onScheduleChange}
        appTimezone={appTimezone}
        tagsCsv={tagsCsv}
        milestone={milestone}
        projectId={projectId}
        dependsOn={dependsOn}
        onTagsCsvChange={onTagsCsvChange}
        onMilestoneChange={onMilestoneChange}
        onDependsOnChange={onDependsOnChange}
      />
    </div>
  );
}
