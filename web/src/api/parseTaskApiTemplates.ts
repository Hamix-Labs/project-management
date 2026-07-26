import type { Task, TaskComposePayload } from "@/types/taskCore";
import type {
  TaskTemplateDetail,
  TaskTemplateSummary,
} from "@/types/taskTemplates";
import { parseComposePayloadCore } from "./parseTaskApiCompose";
import { parseDependsOnList, parseTask } from "./parseTaskApiTasks";
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
  if (item.is_function === true) {
    summary.is_function = true;
  }
  if (Array.isArray(item.input_kinds)) {
    const kinds = item.input_kinds
      .map((k) => (typeof k === "string" ? k.trim() : ""))
      .filter((k): k is "dir" | "file" | "function" =>
        k === "dir" || k === "file" || k === "function",
      );
    if (kinds.length > 0) {
      summary.input_kinds = kinds;
    }
  }
  return summary;
}

function parseFunctionInputs(
  value: unknown,
): import("@/types").TemplateFunctionInputDef[] | undefined {
  if (value === undefined || value === null) return undefined;
  if (!Array.isArray(value)) {
    throw new Error("Invalid API response: payload.function_inputs must be array");
  }
  if (value.length === 0) return undefined;
  return value.map((row, i) => {
    if (!isRecord(row)) {
      throw new Error(`Invalid API response: payload.function_inputs[${i}] must be object`);
    }
    const kind = parseString(row.kind, `payload.function_inputs[${i}].kind`);
    if (kind !== "dir" && kind !== "file" && kind !== "function") {
      throw new Error(
        `Invalid API response: payload.function_inputs[${i}].kind must be dir, file, or function`,
      );
    }
    const def: import("@/types").TemplateFunctionInputDef = {
      id: parseNonEmptyString(row.id, `payload.function_inputs[${i}].id`),
      kind,
      label: parseNonEmptyString(row.label, `payload.function_inputs[${i}].label`),
    };
    if (typeof row.required === "boolean") {
      def.required = row.required;
    }
    if (row.multiple === true) {
      def.multiple = true;
    }
    return def;
  });
}

function parseDependsOnWire(value: unknown): TaskComposePayload["depends_on"] {
  if (value === undefined || value === null) return [];
  return parseDependsOnList(value);
}

export function parseTaskComposePayload(value: unknown): TaskComposePayload {
  if (!isRecord(value)) throw new Error("Invalid API response: payload must be object");
  const core = parseComposePayloadCore(value);
  const functionInputs = parseFunctionInputs(value.function_inputs);
  return {
    ...core,
    status: parseStatus(value.status),
    priority: parsePriorityChoice(value.priority) as TaskComposePayload["priority"],
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
    ...(functionInputs ? { function_inputs: functionInputs } : {}),
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
