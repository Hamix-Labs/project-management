import type { Priority, Status, TaskDependencyEdge, TaskGate } from "@/types";
import { parseTagsFromCsv } from "../create/composePayload";
import { canEditTaskPickupSchedule } from "../task-pickup/canEditTaskPickupSchedule";

export type BuildPatchMutationInputArgs = {
  id: string;
  title: string;
  initial_prompt: string;
  status: Status;
  priority: Priority;
  project_id: string | null;
  tagsCsv: string;
  milestone: string;
  cursor_model: string;
  /** When null/undefined and schedule not editable, pickup is omitted. */
  pickup_not_before?: string | null;
  gate?: TaskGate | null;
  depends_on?: TaskDependencyEdge[];
};

export type PatchMutationInput = {
  id: string;
  title: string;
  initial_prompt: string;
  status: Status;
  priority: Priority;
  project_id: string | null;
  tags: string[];
  milestone: string | null;
  cursor_model: string;
  pickup_not_before?: string | null;
  gate?: TaskGate | null;
  depends_on?: TaskDependencyEdge[];
};

/**
 * Pure builder for task PATCH payloads from edit-form fields.
 * Mirrors create's tag CSV parsing so edit/create cannot drift.
 */
export function buildPatchMutationInput(
  args: BuildPatchMutationInputArgs,
): PatchMutationInput {
  const input: PatchMutationInput = {
    id: args.id,
    title: args.title.trim(),
    initial_prompt: args.initial_prompt,
    status: args.status,
    priority: args.priority,
    project_id: args.project_id,
    tags: parseTagsFromCsv(args.tagsCsv ?? ""),
    milestone: (args.milestone ?? "").trim() || null,
    cursor_model: (args.cursor_model ?? "").trim(),
  };
  if (canEditTaskPickupSchedule(args.status)) {
    input.pickup_not_before = args.pickup_not_before ?? null;
  }
  if (args.gate !== undefined) {
    input.gate = args.gate;
  }
  if (args.depends_on !== undefined) {
    input.depends_on = args.depends_on;
  }
  return input;
}
