import type { TaskDraftChecklistItem } from "@/types/taskCore";
import type { TaskDraftPayload } from "@/types/taskDrafts";
import { parseChecklistItemWire } from "./parseTaskApiChecklist";
import {
  isRecord,
  parsePriorityChoice,
  parseString,
} from "./parseTaskApiCore";

/** Shared compose fields used by task drafts and templates. */
export function parseComposePayloadCore(value: unknown): TaskDraftPayload {
  if (!isRecord(value)) throw new Error("Invalid API response: payload must be object");
  const checklistRaw = value.checklist_items;
  if (!Array.isArray(checklistRaw)) {
    throw new Error("Invalid API response: payload.checklist_items must be array");
  }
  return {
    title: parseString(value.title, "payload.title"),
    initial_prompt: parseString(value.initial_prompt, "payload.initial_prompt"),
    priority: parsePriorityChoice(value.priority),
    checklist_items: checklistRaw.map((row, i) =>
      parseChecklistItemWire(row, `payload.checklist_items[${i}]`) as TaskDraftChecklistItem,
    ),
    ...(typeof value.runner === "string"
      ? { runner: parseString(value.runner, "payload.runner") }
      : {}),
    ...(typeof value.cursor_model === "string"
      ? {
          cursor_model: parseString(
            value.cursor_model,
            "payload.cursor_model",
          ),
        }
      : {}),
    ...(typeof value.project_id === "string"
      ? {
          project_id: parseString(value.project_id, "payload.project_id"),
        }
      : {}),
    ...(typeof value.repository_id === "string"
      ? {
          repository_id: parseString(value.repository_id, "payload.repository_id"),
        }
      : {}),
    ...(typeof value.worktree_id === "string"
      ? {
          worktree_id: parseString(value.worktree_id, "payload.worktree_id"),
        }
      : {}),
  };
}
