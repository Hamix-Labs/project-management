import { useCallback, useRef, type ReactNode } from "react";
import type { ChecklistItemDraft, PriorityChoice, Status } from "@/types";
import type { RichPromptEditorProjectContextProps } from "../rich-prompt";
import { TaskCreateModalAdvancedOptions } from "./TaskCreateModalAdvancedOptions";
import { TaskCreateModalAutonomyToggle } from "./fields/TaskCreateModalAutonomyToggle";
import { TaskCreateModalCriteriaFields } from "./fields/TaskCreateModalCriteriaFields";
import { TaskCreateModalEssentialsFields } from "./fields/TaskCreateModalEssentialsFields";
import { TaskCreateModalPromptFields } from "./fields/TaskCreateModalPromptFields";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";
import { TaskCreateModalTemplateCategoryField } from "./fields/TaskCreateModalTemplateCategoryField";
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
    dependsOn,
    onTagsCsvChange,
    onMilestoneChange,
    onDependsOnChange,
  } = props;

  const openNewCriterionRef = useRef<(() => void) | null>(null);
  const registerOpenNew = useCallback((open: (() => void) | null) => {
    openNewCriterionRef.current = open;
  }, []);
  const editorKey = presentation.isTaskEdit
    ? editingTaskId ?? "edit-prompt-modal"
    : presentation.isTemplateMode
      ? "template-prompt-modal"
      : "create-prompt-modal";
  const checklistRequirement = presentation.isTaskEdit ? "optional" : "required";

  return (
    <div className="task-create-modal-scroll">
      <TaskCreateModalSection
        variant="essentials"
        title="Essentials"
        lede="What to do, how urgent it is, and how success is judged."
      >
        <TaskCreateModalEssentialsFields
          idsPrefix={presentation.idsPrefix}
          title={title}
          priority={priority}
          repositoryId={repositoryId}
          projectId={projectId}
          worktreeId={worktreeId}
          disabled={presentation.disabled}
          showWorktree={!presentation.isTaskEdit}
          onTitleChange={onTitleChange}
          onPriorityChange={onPriorityChange}
          onRepositoryChange={onRepositoryChange}
          onProjectChange={onProjectChange}
          onWorktreeChange={onWorktreeChange}
        />
      </TaskCreateModalSection>

      <TaskCreateModalSection
        variant="prompt"
        title="Initial prompt"
        lede="The full brief the agent starts from. Supports Markdown."
      >
        <TaskCreateModalPromptFields
          idsPrefix={presentation.idsPrefix}
          editorKey={editorKey}
          prompt={prompt}
          disabled={presentation.disabled}
          onPromptChange={onPromptChange}
          projectContext={promptProjectContext}
          worktreeId={worktreeId.trim() || undefined}
        />
      </TaskCreateModalSection>

      <TaskCreateModalSection
        variant="criteria"
        title="Done criteria"
        lede="Clear, checkable conditions that define when this task is complete."
        requirement={checklistRequirement}
        action={
          <button
            type="button"
            className="task-detail-add-checklist-btn"
            disabled={presentation.disabled || presentation.isTaskEdit}
            onClick={() => openNewCriterionRef.current?.()}
          >
            New criterion
          </button>
        }
      >
        <TaskCreateModalCriteriaFields
          checklistItems={checklistItems}
          checklistRequirement={checklistRequirement}
          checklistDisabled={presentation.isTaskEdit}
          disabled={presentation.disabled}
          onAppendChecklistCriterion={onAppendChecklistCriterion}
          onUpdateChecklistRow={onUpdateChecklistRow}
          onRemoveChecklistRow={onRemoveChecklistRow}
          registerOpenNew={registerOpenNew}
        />
        {presentation.isTemplateMode ? (
          <TaskCreateModalTemplateCategoryField
            idsPrefix={presentation.idsPrefix}
            tagsCsv={tagsCsv}
            disabled={presentation.disabled}
            onTagsCsvChange={onTagsCsvChange}
          />
        ) : null}
      </TaskCreateModalSection>

      {projectAssignment ? (
        <TaskCreateModalSection
          variant="context"
          title="Project"
          lede="Scope this task to a project and attach context the agent can reference."
        >
          {projectAssignment}
        </TaskCreateModalSection>
      ) : null}

      <TaskCreateModalSection
        variant="execution"
        title="Execution"
        lede="Choose whether the agent should start on its own and which runner to use."
      >
        <TaskCreateModalAutonomyToggle
          enabled={autonomyEnabled}
          disabled={presentation.disabled || autonomyDisabled}
          onChange={onAutonomyChange}
        />

        <TaskCreateModalAdvancedOptions
          presentation={presentation}
          editingTaskRunner={editingTaskRunner}
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
      </TaskCreateModalSection>
    </div>
  );
}
