import {
  DEFAULT_EDITOR_ID,
  EDITORS,
  getEditorById,
  isEditorId,
  type EditorDefinition,
  type EditorId,
} from "./registry";

export const LAST_EDITOR_STORAGE_KEY = "hamix.openInEditor";

export function getLastEditorId(): EditorId {
  try {
    const raw = localStorage.getItem(LAST_EDITOR_STORAGE_KEY)?.trim() ?? "";
    if (raw !== "" && isEditorId(raw)) {
      return raw;
    }
  } catch {
    // localStorage may be unavailable (private mode / SSR); fall through.
  }
  return DEFAULT_EDITOR_ID;
}

export function setLastEditorId(id: EditorId): void {
  try {
    localStorage.setItem(LAST_EDITOR_STORAGE_KEY, id);
  } catch {
    // Ignore quota / privacy errors — preference is best-effort.
  }
}

/** Editors ordered with the last-used id first, then registry order. */
export function editorsForMenu(preferredId?: EditorId): EditorDefinition[] {
  const preferred = preferredId ?? getLastEditorId();
  const head = getEditorById(preferred);
  if (!head) {
    return [...EDITORS];
  }
  return [head, ...EDITORS.filter((editor) => editor.id !== preferred)];
}
