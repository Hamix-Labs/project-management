import type { ProjectContextItem } from "@/types";

/**
 * Server cap from `pkgs/tasks/store/internal/tasks/project_context_selection.go`
 * (`maxSelectedProjectContextItems`). Mirrored here so the SPA can refuse to
 * add an item that would push past the limit instead of letting the API
 * reject the create/patch with a 400.
 */
export const MAX_SELECTED_PROJECT_CONTEXT_ITEMS = 20;

/**
 * Length of the human-facing short ID we render inside `#` mention chips and
 * the read-only REFERENCES block. Six characters is enough to distinguish a
 * few dozen siblings without bloating the chip and matches the convention
 * `#Decision title · a1b2c3`.
 */
export const PROJECT_CONTEXT_SHORT_ID_LENGTH = 6;

/**
 * Render a stable short identifier for a project context item. We strip
 * dashes (UUIDs) and underscores, then take the first N alphanumeric
 * characters lowercased. Falls back to the trimmed id when the cleaned
 * string is shorter than `PROJECT_CONTEXT_SHORT_ID_LENGTH` so callers
 * always get something printable.
 */
export function projectContextShortId(rawId: string): string {
  const trimmed = (rawId ?? "").trim();
  if (!trimmed) return "";
  const cleaned = trimmed.replace(/[^A-Za-z0-9]/g, "").toLowerCase();
  if (cleaned.length === 0) return trimmed.slice(0, PROJECT_CONTEXT_SHORT_ID_LENGTH);
  return cleaned.slice(0, PROJECT_CONTEXT_SHORT_ID_LENGTH);
}

/**
 * Merge `incoming` ids into `existing`, preserving order, deduping, and
 * stopping at `MAX_SELECTED_PROJECT_CONTEXT_ITEMS`. Returns the existing
 * array unchanged when nothing new would be appended (to keep React's
 * referential equality and skip needless re-renders).
 */
export function mergeProjectContextSelection(
  existing: readonly string[],
  incoming: readonly string[],
): string[] {
  if (incoming.length === 0) return existing.slice();
  const seen = new Set(existing);
  const merged = existing.slice();
  let changed = false;
  for (const id of incoming) {
    const trimmed = (id ?? "").trim();
    if (!trimmed) continue;
    if (seen.has(trimmed)) continue;
    if (merged.length >= MAX_SELECTED_PROJECT_CONTEXT_ITEMS) break;
    merged.push(trimmed);
    seen.add(trimmed);
    changed = true;
  }
  if (!changed) return existing.slice();
  return merged;
}

/**
 * Order the resolved `ProjectContextItem` records to match `selectedIds`,
 * dropping ids that no longer exist in the supplied items list. Used by the
 * REFERENCES block and the selected-summary panel so the visual order tracks
 * the operator's selection order.
 */
export function selectedProjectContextItems(
  items: readonly ProjectContextItem[],
  selectedIds: readonly string[],
): ProjectContextItem[] {
  if (selectedIds.length === 0 || items.length === 0) return [];
  const byId = new Map(items.map((item) => [item.id, item]));
  const out: ProjectContextItem[] = [];
  for (const id of selectedIds) {
    const item = byId.get(id);
    if (item) out.push(item);
  }
  return out;
}
