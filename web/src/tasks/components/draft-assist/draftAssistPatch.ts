import type { DraftAssistPatchEventData } from "@/types/draftAssist";

/**
 * Apply a wire-level `patch` frame to the current prompt HTML/string.
 *
 * The backend validates against a TipTap subset before emitting; the
 * SPA mirrors the *operation* list — set / find_replace / append — and
 * treats anything else as a no-op (see `docs/design/task-draft-ai.md`
 * §MCP-first contract). Returns the next value, or `null` when the
 * patch cannot be applied cleanly (unknown op, missing `find` needle,
 * or find not present in current text). Callers should surface the
 * `null` case as an error row in the thread rather than silently
 * dropping the patch.
 */
export function applyDraftAssistPatch(
  current: string,
  patch: DraftAssistPatchEventData,
): string | null {
  switch (patch.op) {
    case "set":
      return patch.value ?? "";

    case "append":
      return `${current}${patch.value ?? ""}`;

    case "find_replace": {
      const needle = patch.find ?? "";
      if (needle === "") return null;
      if (!current.includes(needle)) return null;
      return current.split(needle).join(patch.value ?? "");
    }

    default:
      return null;
  }
}
