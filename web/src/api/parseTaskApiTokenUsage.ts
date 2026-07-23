import { errorMessage } from "@/lib/errorMessage";
import type {
  TaskTokenUsageAttempt,
  TaskTokenUsageResponse,
  TokenUsageProjection,
} from "@/types";
import {
  isRecord,
  parseBooleanField,
  parseFiniteNumber,
  parseNonEmptyString,
} from "./parseTaskApiCore";

/** Validates the token_usage projection on cycle rows and usage endpoints. */
export function parseTokenUsageProjection(value: unknown): TokenUsageProjection {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: token_usage must be an object");
  }
  return {
    consumed_tokens: parseFiniteNumber(value.consumed_tokens, "consumed_tokens"),
    execute_consumed_tokens: parseFiniteNumber(
      value.execute_consumed_tokens,
      "execute_consumed_tokens",
    ),
    verify_consumed_tokens: parseFiniteNumber(
      value.verify_consumed_tokens,
      "verify_consumed_tokens",
    ),
    input_tokens: parseFiniteNumber(value.input_tokens, "input_tokens"),
    output_tokens: parseFiniteNumber(value.output_tokens, "output_tokens"),
    cache_read_tokens: parseFiniteNumber(value.cache_read_tokens, "cache_read_tokens"),
    cache_write_tokens: parseFiniteNumber(value.cache_write_tokens, "cache_write_tokens"),
    known: parseBooleanField(value.known, "known"),
  };
}

export function parseOptionalTokenUsageProjection(
  value: unknown,
): TokenUsageProjection | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }
  return parseTokenUsageProjection(value);
}

/** Validates `GET /tasks/{id}/token-usage` envelope. */
export function parseTaskTokenUsageResponse(
  value: unknown,
): TaskTokenUsageResponse {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: token usage payload must be an object");
  }
  const rawAttempts = value.attempts;
  if (!Array.isArray(rawAttempts)) {
    throw new Error("Invalid API response: attempts must be an array");
  }
  const attempts = rawAttempts.map((item, i) => {
    try {
      return parseTaskTokenUsageAttempt(item);
    } catch (e) {
      throw new Error(`Invalid API response: attempts[${i}]: ${errorMessage(e)}`);
    }
  });
  return {
    task_id: parseNonEmptyString(value.task_id, "task_id"),
    token_usage: parseTokenUsageProjection(value.token_usage),
    attempts,
  };
}

function parseTaskTokenUsageAttempt(value: unknown): TaskTokenUsageAttempt {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: attempt must be an object");
  }
  const share =
    value.share_of_task_pct === undefined || value.share_of_task_pct === null
      ? null
      : parseFiniteNumber(value.share_of_task_pct, "share_of_task_pct");
  return {
    cycle_id: parseNonEmptyString(value.cycle_id, "cycle_id"),
    attempt_seq: parseFiniteNumber(value.attempt_seq, "attempt_seq"),
    token_usage: parseTokenUsageProjection(value.token_usage),
    share_of_task_pct: share,
  };
}
