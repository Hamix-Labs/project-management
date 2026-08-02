export type EditorMode = "compose" | "edit-task" | "polish" | "template";

export type ResolveEditorTitleContext = {
  formTitle?: string;
  taskName?: string;
  templateName?: string;
};

/** Single title authority for crumb + doc header (call once; pass the string to both). */
export function resolveEditorTitle(
  mode: EditorMode,
  ctx: ResolveEditorTitleContext,
): string {
  switch (mode) {
    case "compose":
      return ctx.formTitle?.trim() || "Untitled task";
    case "edit-task":
    case "polish":
      return ctx.taskName?.trim() || "Untitled task";
    case "template":
      return ctx.templateName?.trim() || "Untitled template";
  }
}
