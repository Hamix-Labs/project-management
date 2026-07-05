import {
  type Task,
  type TaskComposePayload,
  type TaskTemplateDetail,
  type TaskTemplateSummary,
} from "@/types";
import { parseChecklistItemWire, parseTask } from "./parseTaskApiTasks";
import {
  isRecord,
  parseFiniteNumber,
  parseNonEmptyString,
  parsePriorityChoice,
  parseStatus,
  parseString,
} from "./parseTaskApiCore";

function parseOptionalPrimaryTag(value: unknown): string | undefined {
  if (value === undefined || value === null) return undefined;
  const tag = parseString(value, "primary_tag");
  const trimmed = tag.trim();
  return trimmed === "" ? undefined : trimmed;
}

function parseInstantiateCount(value: unknown, field: string): number {
  if (value === undefined || value === null) return 0;
  const n = parseFiniteNumber(value, field);
  if (!Number.isInteger(n) || n < 0) {
    throw new Error(`Invalid API response: ${field} must be a non-negative integer`);
  }
  return n;
}

function parseTaskTemplateSummaryFields(
  item: Record<string, unknown>,
  pathPrefix: string,
): TaskTemplateSummary {
  const summary: TaskTemplateSummary = {
    id: parseNonEmptyString(item.id, `${pathPrefix}.id`),
    name: parseString(item.name, `${pathPrefix}.name`),
    created_at: parseString(item.created_at, `${pathPrefix}.created_at`),
    updated_at: parseString(item.updated_at, `${pathPrefix}.updated_at`),
    instantiate_count: parseInstantiateCount(
      item.instantiate_count,
      `${pathPrefix}.instantiate_count`,
    ),
  };
  const primaryTag = parseOptionalPrimaryTag(item.primary_tag);
  if (primaryTag !== undefined) {
    summary.primary_tag = primaryTag;
  }
  return summary;
}

function parseComposeChecklistItem(
  value: unknown,
  path: string,
): TaskComposePayload["checklist_items"][number] {
  return parseChecklistItemWire(value, path);
}

function parseDependsOnWire(value: unknown): TaskComposePayload["depends_on"] {
  if (value === undefined || value === null) return [];
  if (!Array.isArray(value)) {
    throw new Error("Invalid API response: depends_on must be array");
  }
  return value.map((edge, i) => {
    if (typeof edge === "string") {
      return { task_id: parseString(edge, `depends_on[${i}]`), satisfies: "done" as const };
    }
    if (!isRecord(edge)) {
      throw new Error(`Invalid API response: depends_on[${i}] must be object or string`);
    }
    return {
      task_id: parseString(edge.task_id, `depends_on[${i}].task_id`),
      satisfies: "done" as const,
    };
  });
}

export function parseTaskComposePayload(value: unknown): TaskComposePayload {
  if (!isRecord(value)) throw new Error("Invalid API response: payload must be object");
  const checklistRaw = value.checklist_items;
  if (!Array.isArray(checklistRaw)) {
    throw new Error("Invalid API response: payload.checklist_items must be array");
  }
  return {
    title: parseString(value.title, "payload.title"),
    initial_prompt: parseString(value.initial_prompt, "payload.initial_prompt"),
    status: parseStatus(value.status),
    priority: parsePriorityChoice(value.priority) as TaskComposePayload["priority"],
    checklist_items: checklistRaw.map((row, i) =>
      parseComposeChecklistItem(row, `payload.checklist_items[${i}]`),
    ),
    ...(typeof value.runner === "string"
      ? { runner: parseString(value.runner, "payload.runner") }
      : {}),
    ...(typeof value.cursor_model === "string"
      ? { cursor_model: parseString(value.cursor_model, "payload.cursor_model") }
      : {}),
    ...(typeof value.project_id === "string"
      ? { project_id: parseString(value.project_id, "payload.project_id") }
      : {}),
    ...(Array.isArray(value.project_context_item_ids)
      ? {
          project_context_item_ids: value.project_context_item_ids.map((id, i) =>
            parseString(id, `payload.project_context_item_ids[${i}]`),
          ),
        }
      : {}),
    ...(typeof value.pickup_not_before === "string"
      ? {
          pickup_not_before: parseString(
            value.pickup_not_before,
            "payload.pickup_not_before",
          ),
        }
      : {}),
    ...(Array.isArray(value.tags)
      ? {
          tags: value.tags.map((tag, i) => parseString(tag, `payload.tags[${i}]`)),
        }
      : {}),
    ...(typeof value.milestone === "string"
      ? { milestone: parseString(value.milestone, "payload.milestone") }
      : {}),
    depends_on: parseDependsOnWire(value.depends_on),
  };
}

export function parseTaskTemplateSummaryList(value: unknown): TaskTemplateSummary[] {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: template list must be object");
  }
  const raw = value.templates;
  if (!Array.isArray(raw)) {
    throw new Error("Invalid API response: templates must be array");
  }
  return raw.map((item, i) => {
    if (!isRecord(item)) {
      throw new Error(`Invalid API response: templates[${i}] must be object`);
    }
    return parseTaskTemplateSummaryFields(item, `templates[${i}]`);
  });
}

export function parseTaskTemplateDetail(value: unknown): TaskTemplateDetail {
  if (!isRecord(value)) throw new Error("Invalid API response: template detail must be object");
  return {
    ...parseTaskTemplateSummaryFields(value, "template"),
    payload: parseTaskComposePayload(value.payload),
  };
}

export function parseTaskTemplateInstantiateResponse(value: unknown): {
  tasks: Task[];
  errors: { template_id: string; error: string }[];
} {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: instantiate response must be object");
  }
  const tasksRaw = value.tasks;
  const errorsRaw = value.errors;
  if (!Array.isArray(tasksRaw)) {
    throw new Error("Invalid API response: tasks must be array");
  }
  if (!Array.isArray(errorsRaw)) {
    throw new Error("Invalid API response: errors must be array");
  }
  return {
    tasks: tasksRaw.map((t) => parseTask(t)),
    errors: errorsRaw.map((row, i) => {
      if (!isRecord(row)) {
        throw new Error(`Invalid API response: errors[${i}] must be object`);
      }
      return {
        template_id: parseString(row.template_id, `errors[${i}].template_id`),
        error: parseString(row.error, `errors[${i}].error`),
      };
    }),
  };
}
