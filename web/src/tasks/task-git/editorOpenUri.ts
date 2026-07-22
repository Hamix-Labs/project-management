/**
 * Build an editor deep link that opens a local folder (Mac + Windows).
 * Same shape as VS Code / Cursor `scheme://file/.../` URL handlers.
 *
 * Always appends `?windowId=_blank` so the folder opens in a new window
 * instead of replacing the current workspace (important when an agent is
 * still running in the existing window). Editors that ignore the query
 * still receive a valid folder URI.
 */
export function buildEditorOpenFolderUri(
  fsPath: string,
  scheme: string,
): string {
  const trimmedScheme = scheme.trim().replace(/:\/\/*$/, "");
  if (trimmedScheme === "") {
    throw new Error("scheme is required");
  }

  let path = fsPath.trim().replace(/\\/g, "/");
  if (path === "") {
    throw new Error("path is required");
  }

  const winDrive = /^[A-Za-z]:\//.exec(path);
  if (winDrive) {
    path = path.replace(/\/+$/, "");
    return `${trimmedScheme}://file/${path}/?windowId=_blank`;
  }

  if (!path.startsWith("/")) {
    path = `/${path}`;
  }
  path = path.replace(/\/+$/, "");
  return `${trimmedScheme}://file${path}/?windowId=_blank`;
}
