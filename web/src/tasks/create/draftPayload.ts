import type { AppSettings } from "@/api/settings";
import { TASK_DRAFTS } from "@/constants/tasks";
import {
  type ChecklistItemDraft,
  type PriorityChoice,
  type TaskDraftDetail,
} from "@/types";
import { normalizeChecklistItems } from "../task-compose/checklistRequirement";
import { draftPayloadFingerprint } from "../task-drafts";
import { resumedDraftFields } from "./resumedDraftFields";
import type { DraftPayloadFields, DraftSavePayload } from "./types";

export { mapDraftChecklistItems } from "./resumedDraftFields";

export function buildDraftSavePayload(fields: DraftPayloadFields): DraftSavePayload {
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

/**
 * Dirty-bit fingerprint for the current form, taken from the payload that would
 * be persisted. Deriving it (instead of listing the fields again) is what keeps
 * "saved" and "dirty" in agreement.
 */
export function computeDraftAutosaveSignature(fields: DraftPayloadFields): string {
  return draftPayloadFingerprint(buildDraftSavePayload(fields));
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
  const fields = resumedDraftFields(input.draft, input.settings);
  input.setNewDraftID(fields.newDraftID);
  input.setNewTitle(fields.newTitle);
  input.setNewPrompt(fields.newPrompt);
  input.setNewPriority(fields.newPriority);
  input.setNewTaskRunner(fields.newTaskRunner);
  input.setNewTaskCursorModel(fields.newTaskCursorModel);
  input.setNewChecklistItems(fields.newChecklistItems);
  input.setNewProjectID(fields.newProjectID);
  input.setNewRepositoryID(fields.newRepositoryID);
  input.setNewWorktreeID(fields.newWorktreeID);
  input.setNewSchedule(null);
  input.setNewAutonomyEnabled(true);
  input.setDraftAutosaveBaseline(computeDraftAutosaveSignature(fields));
  input.setDraftAutosaveBaselineID(fields.newDraftID);
}
