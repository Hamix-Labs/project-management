/** Optional shell check attached to a done criterion. */
export type ChecklistVerifyCommandInput = {
  command: string;
  expected_outcome?: string;
  /**
   * Wall-clock cap in seconds. Omit or null = no timeout (runs until cycle
   * cancel). When set, must be a positive integer.
   */
  timeout_seconds?: number | null;
};

/** Draft criterion row in create/edit modals before persistence. */
export type ChecklistItemDraft = {
  text: string;
  verify_commands?: ChecklistVerifyCommandInput[];
};

/** One checklist row from GET /tasks/{id}/checklist. */
export type TaskChecklistItemView = {
  id: string;
  sort_order: number;
  text: string;
  done: boolean;
  evidence?: string;
  verified_by?: string;
  verifier_reasoning?: string;
  cycle_id?: string;
  verify_commands?: ChecklistVerifyCommandInput[];
};

export type TaskChecklistResponse = {
  items: TaskChecklistItemView[];
};

/** UI display cap for evidence text (backend store cap is 16 KB). See docs/data-model.md. */
export const CHECKLIST_EVIDENCE_DISPLAY_CAP = 12 * 1024;

/** Checklist row in compose payloads (drafts, templates, create). */
export type TaskDraftChecklistItem = {
  text: string;
  verify_commands?: ChecklistVerifyCommandInput[];
};
