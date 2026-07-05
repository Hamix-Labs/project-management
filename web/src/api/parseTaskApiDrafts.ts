import {
  type TaskDraftDetail,
  type TaskDraftPayload,
  type TaskDraftSummary,
} from "@/types";
import { parseComposePayloadCore } from "./parseTaskApiCompose";
import {
  isRecord,
  parseNamedEntitySummaryList,
  parseNonEmptyString,
  parseString,
} from "./parseTaskApiCore";

function parseDraftPayload(value: unknown): TaskDraftPayload {
  return parseComposePayloadCore(value);
}

/** Validates GET /task-drafts list JSON (`drafts` array). */
export function parseTaskDraftSummaryList(value: unknown): TaskDraftSummary[] {
  return parseNamedEntitySummaryList(value, "drafts", "draft");
}

/** Validates GET /task-drafts/{id} JSON. */
export function parseTaskDraftDetail(value: unknown): TaskDraftDetail {
  if (!isRecord(value)) throw new Error("Invalid API response: draft detail must be object");
  return {
    id: parseNonEmptyString(value.id, "id"),
    name: parseString(value.name, "name"),
    created_at: parseString(value.created_at, "created_at"),
    updated_at: parseString(value.updated_at, "updated_at"),
    payload: parseDraftPayload(value.payload),
  };
}
