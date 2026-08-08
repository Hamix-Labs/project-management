import type { AppSettings } from "@/api/settings";
import type {
  ChecklistItemDraft,
  TaskDraftChecklistItem,
  TaskDraftDetail,
} from "@/types";
import {
  defaultCursorModelFromSettings,
  defaultRunnerFromSettings,
} from "./defaults";
import type { DraftPayloadFields } from "./types";

export function mapDraftChecklistItems(
  items: TaskDraftChecklistItem[] | undefined,
): ChecklistItemDraft[] {
  return (items ?? []).map((item) => ({
    text: item.text,
    ...(item.verify_commands?.length ? { verify_commands: item.verify_commands } : {}),
  }));
}

/** Older drafts omit runner; fall back to the operator's configured default. */
function resumedRunner(draftRunner: unknown, settings: AppSettings | undefined): string {
  if (typeof draftRunner === "string" && draftRunner.trim()) {
    return draftRunner.trim();
  }
  return defaultRunnerFromSettings(settings);
}

/** An explicitly empty model is a real choice; only absence falls back. */
function resumedCursorModel(
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

/**
 * Form fields a resumed draft restores, as one value.
 *
 * Resume derives both the form state and the autosave baseline from this single
 * result, so a field cannot be restored into the form yet missing from its own
 * baseline (which would make an untouched resumed draft look dirty).
 */
export function resumedDraftFields(
  draft: TaskDraftDetail,
  settings: AppSettings | undefined,
): DraftPayloadFields {
  return {
    newDraftID: draft.id,
    newTitle: draft.payload.title ?? "",
    newPrompt: draft.payload.initial_prompt ?? "",
    newPriority: draft.payload.priority ?? "",
    newTaskRunner: resumedRunner(draft.payload.runner, settings),
    newTaskCursorModel: resumedCursorModel(draft.payload.cursor_model, settings),
    newProjectID: optionalDraftId(draft.payload.project_id),
    newRepositoryID: optionalDraftId(draft.payload.repository_id),
    newWorktreeID: optionalDraftId(draft.payload.worktree_id),
    newChecklistItems: mapDraftChecklistItems(draft.payload.checklist_items),
  };
}

/** Payload-relevant fields of a brand-new draft, for the opening baseline. */
export function freshDraftFields(
  settings: AppSettings | undefined,
  generatedID: string,
): DraftPayloadFields {
  return {
    newDraftID: generatedID,
    newTitle: "",
    newPrompt: "",
    newPriority: "",
    newTaskRunner: defaultRunnerFromSettings(settings),
    newTaskCursorModel: defaultCursorModelFromSettings(settings),
    newProjectID: "",
    newRepositoryID: "",
    newWorktreeID: "",
    newChecklistItems: [],
  };
}
