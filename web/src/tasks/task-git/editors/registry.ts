export type EditorId = "cursor" | "vscode";

export type EditorDefinition = {
  id: EditorId;
  label: string;
  /** Protocol scheme without `://` (e.g. `cursor`, `vscode`). */
  scheme: string;
};

/**
 * Known IDE deep-link targets for “Open in”. Add an entry here to support
 * another editor — URI path normalization stays shared.
 */
export const EDITORS: readonly EditorDefinition[] = [
  { id: "cursor", label: "Cursor", scheme: "cursor" },
  { id: "vscode", label: "VS Code", scheme: "vscode" },
] as const;

export const DEFAULT_EDITOR_ID: EditorId = "cursor";

export function getEditorById(id: string): EditorDefinition | undefined {
  return EDITORS.find((editor) => editor.id === id);
}

export function isEditorId(value: string): value is EditorId {
  return EDITORS.some((editor) => editor.id === value);
}
