import type { AppSettings } from "@/api/settings";
import { TASK_DRAFTS } from "@/constants/tasks";
import {
  type ChecklistItemDraft,
  type PriorityChoice,
  type TaskDraftChecklistItem,
  type TaskDraftDetail,
} from "@/types";
import { normalizeChecklistItems } from "../task-compose/checklistRequirement";
import { draftAutosaveSignature } from "../task-drafts";
import {
  defaultCursorModelFromSettings,
  defaultRunnerFromSettings,
} from "./defaults";
import type { DraftSavePayload, TaskCreateFormFields } from "./types";

export function mapDraftChecklistItems(
  items: TaskDraftChecklistItem[] | undefined,
): ChecklistItemDraft[] {
  return (items ?? []).map((item) => ({
    text: item.text,
    ...(item.verify_commands?.length ? { verify_commands: item.verify_commands } : {}),
  }));
}

function resumedRunnerFromDraft(draftRunner: unknown, settings: AppSettings | undefined): string {
  if (typeof draftRunner === "string" && draftRunner.trim()) {
    return draftRunner.trim();
  }
  return defaultRunnerFromSettings(settings);
}

function resumedCursorModelFromDraft(
  draftModel: unknown,
  settings: AppSettings | undefined,
): string {
  if (typeof draftModel === "string") {
    return draftModel;
  }
  return defaultCursorModelFromSettings(settings);
}

function optionalDraftId(value: unknown): string {
  return typeof value === "string" ? value : "";
}

export function buildResumedDraftAutosaveBaseline(input: {
  draftID: string;
  title: string;
  prompt: string;
  priority: PriorityChoice;
  runner: string;
  cursorModel: string;
  projectID: string;
  repositoryID: string;
  worktreeID: string;
  checklistItems: TaskDraftChecklistItem[];
}): string {
  return draftAutosaveSignature({
    id: input.draftID,
    name: input.title.trim() || TASK_DRAFTS.untitledDraftName,
    title: input.title,
    prompt: input.prompt,
    priority: input.priority,
    runner: input.runner,
    cursorModel: input.cursorModel,
    projectId: input.projectID,
    repositoryId: input.repositoryID,
    worktreeId: input.worktreeID,
    checklistItems: input.checklistItems,
  });
}

export function computeDraftAutosaveSignature(fields: TaskCreateFormFields): string {
  return draftAutosaveSignature({
    id: fields.newDraftID,
    name: fields.newTitle.trim() || TASK_DRAFTS.untitledDraftName,
    title: fields.newTitle,
    prompt: fields.newPrompt,
    priority: fields.newPriority,
    projectId: fields.newProjectID,
    repositoryId: fields.newRepositoryID,
    worktreeId: fields.newWorktreeID,
    checklistItems: normalizeChecklistItems(fields.newChecklistItems),
    runner: fields.newTaskRunner,
    cursorModel: fields.newTaskCursorModel,
  });
}

export function buildDraftSavePayload(fields: TaskCreateFormFields): DraftSavePayload {
  return {
    id: fields.newDraftID,
    name: fields.newTitle.trim() || TASK_DRAFTS.untitledDraftName,
    payload: {
      title: fields.newTitle,
      initial_prompt: fields.newPrompt,
      priority: fields.newPriority,
      runner: fields.newTaskRunner,
      cursor_model: fields.newTaskCursorModel,
      project_id: fields.newProjectID,
      repository_id: fields.newRepositoryID,
      worktree_id: fields.newWorktreeID,
      checklist_items: normalizeChecklistItems(fields.newChecklistItems),
    },
  };
}

export function applyResumedDraftToForm(input: {
  draft: TaskDraftDetail;
  settings: AppSettings | undefined;
  setNewTaskRunner: (runner: string) => void;
  setNewTaskCursorModel: (model: string) => void;
  setNewSchedule: (schedule: string | null) => void;
  setNewAutonomyEnabled: (enabled: boolean) => void;
  setNewDraftID: (id: string) => void;
  setNewTitle: (title: string) => void;
  setNewPrompt: (prompt: string) => void;
  setNewPriority: (priority: PriorityChoice) => void;
  setNewChecklistItems: (items: ChecklistItemDraft[]) => void;
  setNewProjectID: (id: string) => void;
  setNewRepositoryID: (id: string) => void;
  setNewWorktreeID: (id: string) => void;
  setDraftAutosaveBaseline: (baseline: string) => void;
  setDraftAutosaveBaselineID: (id: string) => void;
}) {
  const resumedRunner = resumedRunnerFromDraft(input.draft.payload.runner, input.settings);
  const resumedModel = resumedCursorModelFromDraft(
    input.draft.payload.cursor_model,
    input.settings,
  );
  input.setNewTaskRunner(resumedRunner);
  input.setNewTaskCursorModel(resumedModel);
  input.setNewSchedule(null);
  input.setNewAutonomyEnabled(true);
  input.setNewDraftID(input.draft.id);
  input.setNewTitle(input.draft.payload.title ?? "");
  input.setNewPrompt(input.draft.payload.initial_prompt ?? "");
  input.setNewPriority(input.draft.payload.priority ?? "");
  input.setNewChecklistItems(mapDraftChecklistItems(input.draft.payload.checklist_items));
  const resumedProjectID = optionalDraftId(input.draft.payload.project_id);
  const resumedRepositoryID = optionalDraftId(input.draft.payload.repository_id);
  const resumedWorktreeID = optionalDraftId(input.draft.payload.worktree_id);
  input.setNewProjectID(resumedProjectID);
  input.setNewRepositoryID(resumedRepositoryID);
  input.setNewWorktreeID(resumedWorktreeID);
  const resumedTitle = input.draft.payload.title ?? "";
  input.setDraftAutosaveBaseline(
    buildResumedDraftAutosaveBaseline({
      draftID: input.draft.id,
      title: resumedTitle,
      prompt: input.draft.payload.initial_prompt ?? "",
      priority: input.draft.payload.priority ?? "",
      runner: resumedRunner,
      cursorModel: resumedModel,
      projectID: resumedProjectID,
      repositoryID: resumedRepositoryID,
      worktreeID: resumedWorktreeID,
      checklistItems: input.draft.payload.checklist_items ?? [],
    }),
  );
  input.setDraftAutosaveBaselineID(input.draft.id);
}
