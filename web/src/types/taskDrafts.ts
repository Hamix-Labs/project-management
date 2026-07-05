import type { PriorityChoice, TaskDraftChecklistItem } from "./taskCore";

export type TaskDraftPayload = {
  title: string;
  initial_prompt: string;
  priority: PriorityChoice;
  /** Omitted in older drafts; defaults from app settings when missing. */
  runner?: string;
  cursor_model?: string;
  /**
   * Legacy drafts store plain strings; newer drafts store objects with
   * optional `verify_commands`. The parser normalizes both to objects.
   */
  checklist_items: TaskDraftChecklistItem[];
  /**
   * Optional in older drafts. When present, the draft restores the project
   * the operator was composing against on save. Empty string means "no
   * project bound" (falls back to the default project on resume).
   */
  project_id?: string;
  /**
   * Optional in older drafts. Mirrors the task's `project_context_item_ids`
   * so the REFERENCES block in the prompt editor reappears with the same
   * memory nodes the operator selected before closing the modal.
   */
  project_context_item_ids?: string[];
};

export type TaskDraftSummary = {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
};

export type TaskDraftDetail = TaskDraftSummary & {
  payload: TaskDraftPayload;
};
