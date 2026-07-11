import { errorMessage } from "@/lib/errorMessage";
import type {
  Task,
  TaskDependencyEdge,
  TaskDependencySatisfies,
  TaskListResponse,
} from "@/types";
import {
  isRecord,
  parseBooleanField,
  parseFiniteNumber,
  parseISO8601Required,
  parseNonEmptyString,
  parsePriority,
  parseStatus,
  parseString,
} from "./parseTaskApiCore";
import { parseTaskGate } from "./parseGate";

/** Validates JSON from GET /tasks before the UI relies on it. */
export function parseTaskListResponse(value: unknown): TaskListResponse {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: list payload must be an object");
  }
  const rawTasks = value.tasks;
  if (rawTasks === null || rawTasks === undefined) {
    return {
      tasks: [],
      limit: parseFiniteNumber(value.limit, "limit"),
      offset: parseFiniteNumber(value.offset, "offset"),
      has_more: parseBooleanField(value.has_more, "has_more"),
    };
  }
  if (!Array.isArray(rawTasks)) {
    throw new Error("Invalid API response: tasks must be an array");
  }
  const tasks = rawTasks.map((item, i) => {
    try {
      return parseTask(item);
    } catch (e) {
      throw new Error(`Invalid API response: tasks[${i}]: ${errorMessage(e)}`);
    }
  });
  return {
    tasks,
    limit: parseFiniteNumber(value.limit, "limit"),
    offset: parseFiniteNumber(value.offset, "offset"),
    has_more: parseBooleanField(value.has_more, "has_more"),
  };
}

function parseDependencySatisfies(
  value: unknown,
  field: string,
): TaskDependencySatisfies {
  if (value === undefined || value === null || value === "") {
    return "done";
  }
  const s = parseString(value, field);
  if (s === "done") {
    return s;
  }
  throw new Error(`Invalid API response: ${field} must be done`);
}

function parseDependsOnEdge(raw: unknown, path: string): TaskDependencyEdge {
  if (typeof raw === "string") {
    return { task_id: parseNonEmptyString(raw, path), satisfies: "done" };
  }
  if (!isRecord(raw)) {
    throw new Error(`Invalid API response: ${path} must be a string or object`);
  }
  return {
    task_id: parseNonEmptyString(raw.task_id, `${path}.task_id`),
    satisfies: parseDependencySatisfies(raw.satisfies, `${path}.satisfies`),
  };
}

/** Validates a single task object from POST/PATCH responses. */
export function parseTask(value: unknown): Task {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: task must be an object");
  }
  const initial =
    value.initial_prompt === undefined
      ? ""
      : parseString(value.initial_prompt, "initial_prompt");
  const base: Task = {
    id: parseNonEmptyString(value.id, "id"),
    title: parseString(value.title, "title"),
    initial_prompt: initial,
    status: parseStatus(value.status),
    priority: parsePriority(value.priority),
    runner:
      value.runner !== undefined && value.runner !== null
        ? parseString(value.runner, "runner")
        : "cursor",
    cursor_model:
      value.cursor_model !== undefined && value.cursor_model !== null
        ? parseString(value.cursor_model, "cursor_model")
        : "",
  };
  if (
    value.pickup_not_before !== undefined &&
    value.pickup_not_before !== null
  ) {
    base.pickup_not_before = parseString(
      value.pickup_not_before,
      "pickup_not_before",
    );
  }
  if (value.created_at !== undefined && value.created_at !== null) {
    base.created_at = parseISO8601Required(value.created_at, "created_at");
  }
  if (value.project_id !== undefined && value.project_id !== null) {
    const projectID = parseString(value.project_id, "project_id").trim();
    if (projectID !== "") {
      base.project_id = projectID;
    }
  }
  if (Array.isArray(value.project_context_item_ids)) {
    base.project_context_item_ids = value.project_context_item_ids.map((raw, i) =>
      parseNonEmptyString(raw, `project_context_item_ids[${i}]`),
    );
  }
  if (value.worktree_id !== undefined && value.worktree_id !== null) {
    const wtID = parseString(value.worktree_id, "worktree_id").trim();
    if (wtID !== "") {
      base.worktree_id = wtID;
    }
  }
  if (Array.isArray(value.tags)) {
    base.tags = value.tags.map((raw, i) => parseNonEmptyString(raw, `tags[${i}]`));
  } else if (value.tags === undefined) {
    base.tags = [];
  }
  if (value.milestone !== undefined && value.milestone !== null) {
    const m = parseString(value.milestone, "milestone").trim();
    base.milestone = m === "" ? null : m;
  }
  if (Array.isArray(value.depends_on)) {
    base.depends_on = value.depends_on.map((raw, i) =>
      parseDependsOnEdge(raw, `depends_on[${i}]`),
    );
  } else if (value.depends_on === undefined) {
    base.depends_on = [];
  }
  if (value.criteria_satisfied_at !== undefined && value.criteria_satisfied_at !== null) {
    base.criteria_satisfied_at = parseString(
      value.criteria_satisfied_at,
      "criteria_satisfied_at",
    );
  }
  if (value.gate !== undefined && value.gate !== null) {
    base.gate = parseTaskGate(value.gate);
  }
  return base;
}
