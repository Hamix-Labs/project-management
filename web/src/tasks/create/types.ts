import type { UseQueryResult } from "@tanstack/react-query";
import type {
  ChecklistItemDraft,
  Priority,
  PriorityChoice,
  Status,
  TaskDependencyEdge,
  TaskDraftSummary,
} from "@/types";

export type { DraftSavePayload } from "@/types";

export type TaskDraftsQuery = UseQueryResult<TaskDraftSummary[], Error>;

export type CreateTaskMutationInput = {
  title: string;
  initial_prompt: string;
  status: Status;
  priority: Priority;
  checklistItems: ChecklistItemDraft[];
  draft_id: string;
  runner: string;
  cursor_model: string;
  pickup_not_before: string | null;
  project_id: string;
  repository_id: string;
  worktree_id: string;
  tags: string[];
  milestone?: string;
  depends_on: TaskDependencyEdge[];
};

export type CreateModalPrefill = {
  projectID: string;
  repositoryID?: string;
  worktreeID?: string;
  lockProjectAssignment: boolean;
  /** When true, repo/project (and bound worktree) cannot be changed — enqueue mode. */
  lockGitAssignment?: boolean;
};

export type ComposeTarget = "task" | "template";
export type ComposeOperation = "create" | "edit";

export type TaskCreateFormFields = {
  newTitle: string;
  newPrompt: string;
  newPriority: PriorityChoice;
  newTaskRunner: string;
  newTaskCursorModel: string;
  newProjectID: string;
  newRepositoryID: string;
  newWorktreeID: string;
  newSchedule: string | null;
  newAutonomyEnabled: boolean;
  newTagsCsv: string;
  newMilestone: string;
  newDependsOn: string[];
  newChecklistItems: ChecklistItemDraft[];
  newDraftID: string;
  /** Template-only function input schema (empty for ordinary templates/tasks). */
  newFunctionInputs: import("@/types").TemplateFunctionInputDef[];
};

/**
 * The form fields a draft save actually persists. Narrower than
 * `TaskCreateFormFields` so pure mappers (fresh / resumed) can produce exactly
 * what the payload builder consumes, with the compiler rejecting omissions.
 */
export type DraftPayloadFields = Pick<
  TaskCreateFormFields,
  | "newDraftID"
  | "newTitle"
  | "newPrompt"
  | "newPriority"
  | "newTaskRunner"
  | "newTaskCursorModel"
  | "newProjectID"
  | "newRepositoryID"
  | "newWorktreeID"
  | "newChecklistItems"
>;
