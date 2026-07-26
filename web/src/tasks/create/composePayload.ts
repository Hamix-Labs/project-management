import type { AppSettings } from "@/api/settings";
import { type ChecklistItemDraft, type Priority, type Status, type TaskComposePayload } from "@/types";
import { normalizeChecklistItems } from "../task-compose/checklistRequirement";
import { createSubmitStatusForAutonomy, defaultCursorModelFromSettings, defaultRunnerFromSettings } from "./defaults";
import type { TaskCreateFormFields } from "./types";

/** Shared tag CSV parse for create + edit (comma / semicolon / newline).
 * Lowercases to match server `NormalizeTaskTags` / wire rules. */
export function parseTagsFromCsv(csv: string): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const part of csv.split(/[,;\n]+/)) {
    const tag = part.trim().toLowerCase();
    if (!tag || seen.has(tag)) continue;
    seen.add(tag);
    out.push(tag);
  }
  return out;
}

export function buildComposePayloadFromForm(
  fields: TaskCreateFormFields,
): TaskComposePayload {
  return {
    title: fields.newTitle.trim(),
    initial_prompt: fields.newPrompt,
    status: createSubmitStatusForAutonomy(fields.newAutonomyEnabled),
    priority: fields.newPriority as Priority,
    runner: fields.newTaskRunner.trim() || "cursor",
    cursor_model: fields.newTaskCursorModel.trim(),
    verify_chat_mode: fields.newTaskVerifyChatMode.trim() || undefined,
    project_id: fields.newProjectID.trim(),
    repository_id: fields.newRepositoryID.trim(),
    project_context_item_ids: fields.newProjectContextItemIDs,
    pickup_not_before: fields.newSchedule ?? undefined,
    tags: parseTagsFromCsv(fields.newTagsCsv),
    milestone: fields.newMilestone.trim() || undefined,
    depends_on: fields.newDependsOn.map((task_id) => ({ task_id, satisfies: "done" as const })),
    checklist_items: normalizeChecklistItems(fields.newChecklistItems),
  };
}

export function hydrateFormFromComposePayload(
  payload: TaskComposePayload,
  settings: AppSettings | undefined,
): {
  title: string;
  prompt: string;
  priority: TaskCreateFormFields["newPriority"];
  runner: string;
  cursorModel: string;
  projectID: string;
  repositoryID: string;
  worktreeID: string;
  projectContextItemIDs: string[];
  schedule: string | null;
  autonomyEnabled: boolean;
  tagsCsv: string;
  milestone: string;
  dependsOn: string[];
  checklistItems: ChecklistItemDraft[];
} {
  const runner =
    typeof payload.runner === "string" && payload.runner.trim()
      ? payload.runner.trim()
      : defaultRunnerFromSettings(settings);
  const cursorModel =
    typeof payload.cursor_model === "string"
      ? payload.cursor_model
      : defaultCursorModelFromSettings(settings);
  const projectID =
    typeof payload.project_id === "string" ? payload.project_id : "";
  const repositoryID =
    typeof payload.repository_id === "string" ? payload.repository_id : "";
  const worktreeID = "";
  const projectContextItemIDs = Array.isArray(payload.project_context_item_ids)
    ? payload.project_context_item_ids
    : [];
  const status: Status = payload.status ?? "ready";
  return {
    title: payload.title ?? "",
    prompt: payload.initial_prompt ?? "",
    priority: payload.priority ?? "",
    runner,
    cursorModel,
    projectID,
    repositoryID,
    worktreeID,
    projectContextItemIDs,
    schedule: payload.pickup_not_before ?? null,
    autonomyEnabled: status === "ready",
    tagsCsv: (payload.tags ?? []).join(", "),
    milestone: payload.milestone ?? "",
    dependsOn: (payload.depends_on ?? []).map((edge) => edge.task_id),
    checklistItems: (payload.checklist_items ?? []).map((item) => ({
      text: item.text,
      ...(item.verify_commands?.length ? { verify_commands: item.verify_commands } : {}),
    })),
  };
}
