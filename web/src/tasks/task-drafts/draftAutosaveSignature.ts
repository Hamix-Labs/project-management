import type { DraftSavePayload } from "@/types";

/**
 * Treat editor-empty TipTap markup as the empty string when computing the
 * autosave signature. Without this, opening the create modal and tabbing
 * through fields would produce `<p></p>` / `<p><br></p>` / NBSP-only
 * paragraphs that look identical to the user but flip the dirty bit and
 * trigger pointless POST /tasks/drafts writes.
 */
export function normalizeDraftPromptForDirty(prompt: string): string {
  const compact = prompt.replace(/[\s\u200B\uFEFF]/g, "").toLowerCase();
  if (
    compact === "" ||
    compact === "<p></p>" ||
    compact === "<p><br></p>" ||
    /^<p>(<br\/?>|&nbsp;|&#160;)*<\/p>$/.test(compact)
  ) {
    return "";
  }
  return prompt;
}

/**
 * Stable, JSON-serialized fingerprint of a create-task draft.
 * Used by the autosave loop to short-circuit no-op writes: when the current
 * fingerprint equals the last-saved baseline, the debounce timer skips POST.
 *
 * Derived from the save payload itself rather than a parallel field list. That
 * is the point: a field that reaches the server necessarily reaches the dirty
 * bit, so the two cannot drift (which is how `repository_id` once became
 * unsaveable — it was persisted but invisible to the gate).
 *
 * Key order is inherited from the payload builder. Overwriting an existing key
 * via spread keeps its original position, so `initial_prompt` stays put.
 */
export function draftPayloadFingerprint(input: DraftSavePayload): string {
  return JSON.stringify({
    id: input.id,
    name: input.name,
    payload: {
      ...input.payload,
      initial_prompt: normalizeDraftPromptForDirty(input.payload.initial_prompt),
    },
  });
}
