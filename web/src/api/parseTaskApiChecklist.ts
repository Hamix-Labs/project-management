import type { ChecklistVerifyCommandInput } from "@/types";
import type { TaskChecklistItemView, TaskChecklistResponse } from "@/types";
import {
  isRecord,
  parseFiniteNumber,
  parseNonEmptyString,
  parseString,
} from "./parseTaskApiCore";

/** Validates one verify_commands row on checklist API responses. */
export function parseChecklistVerifyCommand(
  value: unknown,
  path: string,
): ChecklistVerifyCommandInput {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be an object`);
  }
  return {
    command: parseString(value.command, `${path}.command`),
    expected_outcome:
      typeof value.expected_outcome === "string"
        ? value.expected_outcome
        : undefined,
  };
}

export type ChecklistItemWire = {
  text: string;
  verify_commands?: ChecklistVerifyCommandInput[];
};

/** Validates a draft/template checklist row (string or object with optional verify_commands). */
export function parseChecklistItemWire(
  value: unknown,
  path: string,
): ChecklistItemWire {
  if (typeof value === "string") {
    return { text: parseString(value, path) };
  }
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be string or object`);
  }
  let verify_commands: ChecklistItemWire["verify_commands"];
  if (value.verify_commands !== undefined && value.verify_commands !== null) {
    if (!Array.isArray(value.verify_commands)) {
      throw new Error(`Invalid API response: ${path}.verify_commands must be an array`);
    }
    verify_commands = value.verify_commands.map((cmd, j) =>
      parseChecklistVerifyCommand(cmd, `${path}.verify_commands[${j}]`),
    );
  }
  return {
    text: parseString(value.text, `${path}.text`),
    ...(verify_commands !== undefined && verify_commands.length > 0
      ? { verify_commands }
      : {}),
  };
}

/** Validates GET /tasks/{id}/checklist JSON. */
export function parseTaskChecklistResponse(value: unknown): TaskChecklistResponse {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: checklist payload must be an object");
  }
  const raw = value.items;
  if (!Array.isArray(raw)) {
    throw new Error("Invalid API response: items must be an array");
  }
  const items: TaskChecklistItemView[] = raw.map((row, i) => {
    if (!isRecord(row)) {
      throw new Error(`Invalid API response: items[${i}] must be an object`);
    }
    let verify_commands: TaskChecklistItemView["verify_commands"];
    if (row.verify_commands !== undefined && row.verify_commands !== null) {
      if (!Array.isArray(row.verify_commands)) {
        throw new Error(`Invalid API response: items[${i}].verify_commands must be an array`);
      }
      verify_commands = row.verify_commands.map((cmd, j) =>
        parseChecklistVerifyCommand(cmd, `items[${i}].verify_commands[${j}]`),
      );
    }
    return {
      id: parseNonEmptyString(row.id, "id"),
      sort_order: parseFiniteNumber(row.sort_order, "sort_order"),
      text: parseString(row.text, "text"),
      done: row.done === true,
      evidence: typeof row.evidence === "string" ? row.evidence : undefined,
      verified_by:
        typeof row.verified_by === "string" ? row.verified_by : undefined,
      verifier_reasoning:
        typeof row.verifier_reasoning === "string"
          ? row.verifier_reasoning
          : undefined,
      cycle_id: typeof row.cycle_id === "string" ? row.cycle_id : undefined,
      ...(verify_commands !== undefined ? { verify_commands } : {}),
    };
  });
  return { items };
}
