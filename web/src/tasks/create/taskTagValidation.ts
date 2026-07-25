import { errorMessage } from "@/lib/errorMessage";
import { parseTagsFromCsv } from "./composePayload";

/** Matches server `taskTagPattern` after normalize (lowercase). */
export const TASK_TAG_PATTERN = /^[a-z0-9][a-z0-9._-]{0,31}$/;

export const TASK_TAG_RULES_HINT =
  "Tags must be 1–32 characters: lowercase letters, numbers, and . _ - only (no spaces). Separate multiple tags with commas.";

/** Client-side tag CSV check before create/edit submit. */
export function validateTagsCsv(csv: string): string | null {
  for (const tag of parseTagsFromCsv(csv)) {
    if (!TASK_TAG_PATTERN.test(tag)) {
      return `Tag "${tag}" is invalid. ${TASK_TAG_RULES_HINT}`;
    }
  }
  return null;
}

const REQUEST_ID_SUFFIX = /\s*\(request [a-f0-9-]+\)\s*$/i;
const INVALID_TAG = /^invalid tag "([^"]*)"(?::\s*(.+))?$/i;
const DUPLICATE_TAG = /^duplicate tag "([^"]*)"$/i;

/**
 * Map create/PATCH tag failures to operator-facing copy and drop request ids.
 */
export function taskMutationErrorMessage(err: unknown, fallback?: string): string {
  const raw = errorMessage(err, fallback);
  const stripped = raw.replace(REQUEST_ID_SUFFIX, "").trim();
  const invalid = INVALID_TAG.exec(stripped);
  if (invalid) {
    const tag = invalid[1] ?? "";
    const detail = (invalid[2] ?? "").trim();
    if (detail) {
      return `Tag "${tag}" is invalid. ${detail}`;
    }
    return `Tag "${tag}" is invalid. ${TASK_TAG_RULES_HINT}`;
  }
  const dup = DUPLICATE_TAG.exec(stripped);
  if (dup) {
    return `Tag "${dup[1]}" is listed more than once.`;
  }
  return stripped || raw;
}
