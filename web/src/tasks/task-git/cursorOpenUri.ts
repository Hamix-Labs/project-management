/**
 * Build a Cursor deep link that opens a local folder (Mac + Windows).
 * Same shape as VS Code `vscode://file/.../` URL handlers.
 *
 * Always appends `?windowId=_blank` so the folder opens in a new Cursor
 * window instead of replacing the current workspace (important when an
 * agent is still running in the existing window).
 */
export function buildCursorOpenFolderUri(fsPath: string): string {
  let path = fsPath.trim().replace(/\\/g, "/");
  if (path === "") {
    throw new Error("path is required");
  }

  const winDrive = /^[A-Za-z]:\//.exec(path);
  if (winDrive) {
    path = path.replace(/\/+$/, "");
    return `cursor://file/${path}/?windowId=_blank`;
  }

  if (!path.startsWith("/")) {
    path = `/${path}`;
  }
  path = path.replace(/\/+$/, "");
  return `cursor://file${path}/?windowId=_blank`;
}
