import type { FormEvent } from "react";
import type { ChecklistItemDraft, PriorityChoice, Status } from "@/types";
import type { TestScenario } from "@/tasks/test-scenarios";

/** Modal open mode, busy flags, and mutation/validation errors. */
export type TaskCreateModalSession = {
  /** When set, the modal edits an existing task using the same layout as create. */
  editingTaskId?: string | null;
  /** When set with composeTarget template + edit, keys the prompt editor for remount. */
  editingTemplateId?: string | null;
  composeTarget?: "task" | "template";
  composeOperation?: "create" | "edit";
  editingTaskRunner?: string;
  composeStatus?: Status;
  onComposeStatusChange?: (status: Status) => void;
  /** Edit-mode PATCH in flight (maps to modal `busy`). */
  patchPending?: boolean;
  patchError?: string | null;
  /** Client-side validation (e.g. missing title) in edit mode. */
  formError?: string | null;
  pending: boolean;
  saving: boolean;
  draftSaving: boolean;
  draftSaveLabel: string | null;
  draftSaveError: boolean;
  createError?: Error | null;
  createFormError?: string | null;
};

export type TaskCreateModalEssentials = {
  title: string;
  priority: PriorityChoice;
  onTitleChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
};

export type TaskCreateModalPromptFields = {
  prompt: string;
  onPromptChange: (v: string) => void;
};

export type TaskCreateModalCriteria = {
  checklistItems: ChecklistItemDraft[];
  onAppendChecklistCriterion: (item: ChecklistItemDraft | string) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  tagsCsv: string;
  onTagsCsvChange: (value: string) => void;
  functionInputs: import("@/types").TemplateFunctionInputDef[];
  onFunctionInputsChange: (next: import("@/types").TemplateFunctionInputDef[]) => void;
};

export type TaskCreateModalGitBinding = {
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  onRepositoryChange: (repositoryId: string) => void;
  onProjectChange: (projectId: string) => void;
  onWorktreeChange: (worktreeId: string) => void;
};

export type TaskCreateModalExecution = {
  taskRunner: string;
  taskCursorModel: string;
  onTaskRunnerChange: (runner: string) => void;
  onTaskCursorModelChange: (v: string) => void;
  schedule: string | null;
  onScheduleChange: (next: string | null) => void;
  autonomyEnabled: boolean;
  onAutonomyChange: (enabled: boolean) => void;
  autonomyDisabled?: boolean;
  milestone: string;
  onMilestoneChange: (value: string) => void;
  dependsOn: string[];
  onDependsOnChange: (value: string[]) => void;
};

export type TaskCreateModalActions = {
  onClose: () => void;
  onSaveDraft: () => void;
  onSubmit: (e: FormEvent) => void;
  onApplyTestScenario?: (scenario: TestScenario) => void;
};

/**
 * Public create/edit modal contract grouped by form section.
 * Prefer {@link buildTaskCreateModalProps} at the layer / test boundary.
 */
export type TaskCreateModalProps = {
  session: TaskCreateModalSession;
  essentials: TaskCreateModalEssentials;
  prompt: TaskCreateModalPromptFields;
  criteria: TaskCreateModalCriteria;
  git: TaskCreateModalGitBinding;
  execution: TaskCreateModalExecution;
  actions: TaskCreateModalActions;
  appTimezone: string;
};

/** Flat field bag accepted by {@link buildTaskCreateModalProps}. */
export type TaskCreateModalFlatInput = TaskCreateModalSession &
  TaskCreateModalEssentials &
  TaskCreateModalPromptFields &
  TaskCreateModalCriteria &
  TaskCreateModalGitBinding &
  TaskCreateModalExecution &
  TaskCreateModalActions & {
    appTimezone: string;
  };
