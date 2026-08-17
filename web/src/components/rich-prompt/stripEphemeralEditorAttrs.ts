/**
 * UniqueID is in-editor for drag handles. It must not leak into stored
 * `initial_prompt` HTML (that would dirty drafts on every mount).
 */
export function stripEphemeralEditorAttrs(html: string): string {
  return html.replace(/\sdata-id="[^"]*"/g, "");
}
